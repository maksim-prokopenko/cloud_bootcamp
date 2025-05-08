package servicemanager

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/maximmihin/cb625/internal/balancer/roundrobin"
	weightedroundrobin "github.com/maximmihin/cb625/internal/balancer/roundrobin/weighted"
)

// TODO add feature update all
type ServiceManager struct {
	locator map[string]http.Handler
	logger  *slog.Logger
}

func NewServiceManager(cfg *ServiceManagerConfig, logger *slog.Logger) (*ServiceManager, error) {
	sm := &ServiceManager{
		locator: make(map[string]http.Handler),
		logger:  logger,
	}
	var errs []error

	for _, serviceConfig := range cfg.Services {

		var (
			tmpHandler http.Handler
			err        error
			urlStr     string
		)

		switch typedServiceConfig := serviceConfig.(type) {

		case *roundrobin.Config:
			typedServiceConfig.Logger = logger
			tmpHandler, err = roundrobin.New(*typedServiceConfig)
			urlStr = typedServiceConfig.ServerName

		case *weightedroundrobin.Config:
			typedServiceConfig.Logger = logger
			tmpHandler, err = weightedroundrobin.New(*typedServiceConfig)
			urlStr = typedServiceConfig.ServerName

		default:
			errs = append(errs, errors.New("unexpected error: unexpected config type"))
			continue
		}

		if err != nil {
			errs = append(errs, err)
			continue
		}

		parsedUrl, err := url.Parse(urlStr)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed parse url [%s]: %w ", urlStr, err))
			continue
		}
		_, ok := sm.locator[parsedUrl.Host]
		if ok {
			errs = append(errs, fmt.Errorf("url [%s] was added yet (by different/current service)", parsedUrl.Host))
			continue
		}
		sm.locator[parsedUrl.Host] = tmpHandler

	}
	if len(sm.locator) == 0 {
		errs = append(errs, errors.New("no one service was set"))
	}
	return sm, errors.Join(errs...)
}

func (sm *ServiceManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service, ok := sm.locator[r.Host] // todo full path ?
	if !ok {
		sm.logger.Error("try to go on unregister url")                      // TODO ???
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable) // TODO not json?
		return
	}
	service.ServeHTTP(w, r)
}
