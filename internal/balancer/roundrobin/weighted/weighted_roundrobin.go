package weighted

import (
	"errors"
	"github.com/maximmihin/cb625/internal/balancer/roundrobin/carousel"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/maximmihin/cb625/internal/balancer/healthcheck"
)

type WeightedCarousel interface {
	SetWithWeight(url string, info any, weight int)
	Next() any
}

type roundRobin struct {
	backendsCarousel WeightedCarousel
	logger           *slog.Logger
}

type Config struct {
	ServerName        string                               `json:"server_name"`
	BackendUrls       []ServerConfig                       `json:"servers"`
	ActiveHealthCheck *healthcheck.ActiveHealthCheckConfig `json:"active_health_check"`
	Logger            *slog.Logger                         `json:"_"` // TODO check is it work
}

type ServerConfig struct {
	URL    string `json:"url"`
	Weight int    `json:"weight,omitempty"`
}

func New(cfg Config) (http.Handler, error) {

	var err error

	rr := new(roundRobin)
	if cfg.ActiveHealthCheck == nil {
		rr.backendsCarousel = carousel.New()
	} else {
		srr := carousel.NewSmart()
		rr.backendsCarousel = srr
		defer func() { // TODO ugly but how?..
			if err == nil {
				healthcheck.RunActiveHealthCheck(cfg.ActiveHealthCheck, srr)
			}
		}()
	}

	var errs []error
	for _, backendUrl := range cfg.BackendUrls {
		backendUrlParsed, err := url.Parse(backendUrl.URL)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		proxy := httputil.NewSingleHostReverseProxy(backendUrlParsed) // TODO inject logger and error handler in proxy

		rr.backendsCarousel.SetWithWeight(backendUrl.URL, proxy, backendUrl.Weight)
	}
	if len(cfg.BackendUrls) == 0 {
		errs = append(errs, errors.New("empty backends list"))
	}
	err = errors.Join(errs...)
	if err != nil {
		return nil, err
	}

	return rr, nil
}

func (rr *roundRobin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO handle error proxy when server on another port
	handler := rr.backendsCarousel.Next().(http.Handler)
	if handler == nil {
		rr.logger.Error("no one alive backend")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable) // TODO not json?
		return
	}
	handler.ServeHTTP(w, r)
}
