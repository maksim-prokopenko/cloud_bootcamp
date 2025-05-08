package app

import (
	"encoding/json"
	"fmt"
	"github.com/maximmihin/cb625/internal/balancer/roundrobin"
	"log/slog"
	"os"

	"github.com/maximmihin/cb625/internal/healthcheck"
	"github.com/maximmihin/cb625/internal/limiter/tokenbucket"
)

type Config struct {
	Balancer          any                                  `json:"balancer"`
	Limiter           any                                  `json:"limiter"`
	ActiveHealthCheck *healthcheck.ActiveHealthCheckConfig `json:"active_health_check"`

	Logger *slog.Logger `json:"-"`
}

func (cfg *Config) UnmarshalJSON(bytes []byte) error {

	var errs []error

	var tmpConfig struct {
		Balancer          json.RawMessage                      `json:"balancer"`
		Limiter           json.RawMessage                      `json:"limiter"`
		ActiveHealthCheck *healthcheck.ActiveHealthCheckConfig `json:"active_health_check"`
	}

	err := json.Unmarshal(bytes, &tmpConfig)
	if err != nil {
		errs = append(errs, err)
	}

	cfg.Balancer, err = unmarshalBalancer(tmpConfig.Balancer)
	if err != nil {
		errs = append(errs, err)
	}

	cfg.Limiter, err = unmarshalLimiter(tmpConfig.Limiter)
	if err != nil {
		errs = append(errs, err)
	}

	cfg.ActiveHealthCheck = tmpConfig.ActiveHealthCheck

	return nil
}

func unmarshalBalancer(message json.RawMessage) (any, error) {
	var tmpBalancer struct {
		Type string `json:"type"`
	}
	err := json.Unmarshal(message, &tmpBalancer)
	if err != nil {
		return nil, err
	}
	switch tmpBalancer.Type {
	case "round_robin":
		rrCfg := new(roundrobin.Config) // todo
		err = json.Unmarshal(message, rrCfg)
		return rrCfg, err
	}
	return nil, fmt.Errorf("unexpected balancer type: %s", tmpBalancer.Type)
}

func unmarshalLimiter(message json.RawMessage) (any, error) {
	var tmpLimiter struct {
		Type string `json:"type"`
	}
	err := json.Unmarshal(message, &tmpLimiter)
	if err != nil {
		return nil, err
	}

	switch tmpLimiter.Type {
	case "token_buket":
		tbCfg := new(tokenbucket.Config)
		err = json.Unmarshal(message, tbCfg)
		return tbCfg, err
	}
	return nil, fmt.Errorf("unexpected limiter type: %s", tmpLimiter.Type)
}

func ParseFromJSON(path string) (*Config, error) {
	rawJSON, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := new(Config)
	err = json.Unmarshal(rawJSON, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
