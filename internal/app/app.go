package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/maximmihin/cb625/internal/balancer/roundrobin"
	"github.com/maximmihin/cb625/internal/healthcheck"
	"github.com/maximmihin/cb625/internal/limiter/tokenbucket"
)

func New(ctx context.Context, cfg *Config) (http.Handler, error) {
	var finalHandler http.Handler
	var errs []error

	// initialize balancer
	{
		var err error
		switch BalancerConfig := cfg.Balancer.(type) {
		case *roundrobin.Config:
			BalancerConfig.Logger = cfg.Logger
			finalHandler, err = roundrobin.New(BalancerConfig)
		default:
			errs = append(errs, errors.New("unexpected balancer type"))
		}
		if err != nil {
			errs = append(errs, err)
		}
	}

	// initialize health check
	if cfg.ActiveHealthCheck != nil {
		h, ok := finalHandler.(healthcheck.Balancer) // it is important to do this immediately after initializing the balancer
		if !ok {
			errs = append(errs, fmt.Errorf("balancer doesn't provide healthcech.Balancer interface"))
		} else {
			cfg.ActiveHealthCheck.Balancer = h
			cfg.ActiveHealthCheck.Logger = cfg.Logger
			healthcheck.RunActiveHealthCheck(ctx, cfg.ActiveHealthCheck) // TODO how to stop it?
		}
	}

	// initialize limiter
	if cfg.Limiter != nil {
		var err error
		switch LimiterConfig := cfg.Limiter.(type) {
		case *tokenbucket.Config:
			LimiterConfig.Logger = cfg.Logger
			LimiterConfig.NextHandler = finalHandler
			LimiterConfig.RequesterExtractor = tokenbucket.HeaderExtractor("ft-header")
			finalHandler, err = tokenbucket.New(LimiterConfig)
		default:
			errs = append(errs, errors.New("unexpected limiter type"))
		}
		if err != nil {
			errs = append(errs, err)
		}
	}

	return finalHandler, errors.Join(errs...)
}
