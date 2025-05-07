package tokenbucket

import (
	"encoding/json"
	"errors"
	"golang.org/x/time/rate"
	"log/slog"
	"net/http"
	"os"
)

type Config struct {
	Next               http.Handler
	ClientsLimits      map[string]float64
	Logger             *slog.Logger
	RequesterExtractor RequesterExtractor
}

type TokenBuket struct {
	nextHandler        http.Handler
	requesterExtractor RequesterExtractor
	limiterMap         *limiterMap
	logger             *slog.Logger
}

func New(cfg Config) (*TokenBuket, error) {

	if err := cfg.WithDefaults().Validate(); err != nil {
		return nil, err
	}

	tb := TokenBuket{
		nextHandler:        cfg.Next,
		limiterMap:         newLimiterMap(len(cfg.ClientsLimits)),
		requesterExtractor: cfg.RequesterExtractor,
		logger:             cfg.Logger,
	}

	for c, l := range cfg.ClientsLimits {
		tb.limiterMap.add(c, rate.NewLimiter(rate.Limit(l), 1))
	}

	return &tb, nil
}

func (tb *TokenBuket) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	logger := tb.logger

	requester := tb.requesterExtractor(r)

	logger = logger.With(
		slog.String("requester", requester))

	limiter, ok := tb.limiterMap.get(requester)
	if !ok { // TODO think about configurable behaviour
		logger.Info("unauthorized or empty requester")
		WriteError(w, "Unauthorized", 401, logger)
		return
	}
	if limiter == nil {
		logger.Error("nil limiter in initialized TokenBuket.limiterMap")
		WriteError(w, "Internal balancer error", 500, logger)
		return
	}

	if limiter.Allow() {
		tb.nextHandler.ServeHTTP(w, r)
	} else {
		logger.Info("not enough tokens")
		WriteError(w, "Too Many Requests", 429, logger) // TODO add Retry-After
	}

}

func (c Config) WithDefaults() Config {
	if c.Logger == nil {
		c.Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}
	if c.RequesterExtractor != nil {
		c.RequesterExtractor = IPPortRequesterExtractor
	}
	return c
}

var ErrUnsetHandlerAfter = errors.New("no handler specified after limiter") // TODO rename
var ErrEmptyRequesterList = errors.New("empty requester list")
var ErrEmptyRequestExtractor = errors.New("empty requester list")

func (c Config) Validate() error {
	var errs []error

	if c.Next == nil {
		errs = append(errs, ErrUnsetHandlerAfter)
	}
	if len(c.ClientsLimits) == 0 { // TODO switch everywhere to usual != ""
		errs = append(errs, ErrEmptyRequesterList)
	}
	if c.RequesterExtractor == nil {
		errs = append(errs, ErrEmptyRequestExtractor)
	}
	return errors.Join(errs...)
}

type RequesterExtractor func(r *http.Request) string // TODO think about interface - would has http error messager

var IPPortRequesterExtractor RequesterExtractor = func(r *http.Request) string {
	return r.RemoteAddr
}

func HeaderExtractor(headerName string) RequesterExtractor {
	return func(r *http.Request) string {
		hs := r.Header[http.CanonicalHeaderKey(headerName)]
		if len(hs) > 0 {
			return hs[0]
		}
		return ""
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteError(w http.ResponseWriter, msg string, httpErrCode int, logger *slog.Logger) {
	w.WriteHeader(httpErrCode)
	err := json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
	if err != nil {
		logger.Error("error while write response")
	}
}
