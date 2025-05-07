package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/maximmihin/cb625/internal/balancer/roundrobin"
	"github.com/maximmihin/cb625/internal/limiter/tokenbucket"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // todo get from env
	}))

	logger.Info("hello world 1")

	cfg, err := ParseConfigFromEnv()
	if err != nil {
		logger.Error(err.Error())
	}
	logger.Debug("valid config",
		slog.Any("config", cfg))

	rr, err := roundrobin.New(roundrobin.Config{
		BackendUrls: cfg.BackendUrls,
		Logger:      logger,
	})
	if err != nil {
		logger.Error(err.Error())
		return
	}

	tb, err := tokenbucket.New(tokenbucket.Config{
		Next:               rr,
		ClientsLimits:      cfg.ClientsLimits,
		RequesterExtractor: tokenbucket.HeaderExtractor("ft-token"), // todo const header?
		Logger:             logger,
	})
	if err != nil {
		logger.Error(err.Error())
		return
	}

	server := http.Server{
		Addr:    ":8080",
		Handler: tb,
	}

	logger.Info("Started at port: 8080")
	logger.Error(server.ListenAndServe().Error()) // todo separate listen and serve?
}

type Config struct {
	BackendUrls   map[string]struct{}
	ClientsLimits map[string]float64
}

func ParseConfigFromEnv() (*Config, error) {

	var errs []error
	var cfg = &Config{
		BackendUrls:   map[string]struct{}{},
		ClientsLimits: map[string]float64{},
	}

	// TODO write more, port can be sat too
	// parse url's to redirect, expected format // BACKEND_URLS="1.1.1.127,1.1.1.128,1.1.1.129"
	{
		backendUrls := strings.Split(strings.Trim(os.Getenv("BACKEND_URLS"), "\""), ",")
		for _, url := range backendUrls {
			cfg.BackendUrls[url] = struct{}{}
		}
		if len(cfg.BackendUrls) != len(backendUrls) {
			errs = append(errs, errors.New("each backend must be unique"))
		}
		if len(cfg.BackendUrls) == 0 {
			errs = append(errs, errors.New("no one server was set"))
		}
	}

	// parse client limits, expected format
	// format CLIENT_LIMITS="<requesterUID>=<generateTokenPerSec>,..."
	// requesterUID (requester Unique ID - any string), generateTokenPerSec must be valid positive float64
	// example: CLIENT_LIMITS="vasya=60,petya=1.5"
	// example: CLIENT_LIMITS="1.1.1.127:11111=123,1.1.1.127:22222=123"
	{
		// TODO check case with unset client_limit
		pairs := strings.Split(strings.Trim(os.Getenv("CLIENT_LIMITS"), "\""), ",")
		for _, pair := range pairs {

			cl := strings.SplitN(pair, "=", 2)
			if len(cl) != 2 {
				errs = append(errs, errors.New("malformed config pair: "+pair))
				continue
			}

			clientIdStr, tokenDurStr := cl[0], cl[1]
			if len(clientIdStr) == 0 || len(tokenDurStr) == 0 {
				errs = append(errs, errors.New("empty client Id or token repair duration in pair: "+pair))
				continue
			}

			// TODO what todo with extremely big number?
			tokenDur, e := strconv.ParseFloat(tokenDurStr, 64)
			if e != nil {
				errs = append(errs, fmt.Errorf("invalid token duration %s: %w", tokenDurStr, e))
				continue
			}

			_, ok := cfg.ClientsLimits[clientIdStr]
			if ok {
				errs = append(errs, fmt.Errorf("each client in config must be unique; %s is duplicate", clientIdStr))
				continue
			}
			cfg.ClientsLimits[clientIdStr] = tokenDur
		}
		if len(cfg.ClientsLimits) == 0 {
			errs = append(errs, errors.New("no one valid client-limit pair was set"))
		}

	}

	return cfg, errors.Join(errs...)
}
