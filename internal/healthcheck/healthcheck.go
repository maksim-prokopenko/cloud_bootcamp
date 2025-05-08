package healthcheck

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type ActiveHealthCheckConfig struct {
	HealthHandler string        `json:"handler"`
	Method        string        `json:"method"`
	Interval      time.Duration `json:"interval"`
	Timeout       time.Duration `json:"timeout"`

	Balancer Balancer     `json:"-"`
	Logger   *slog.Logger `json:"-"`
}

type Balancer interface {
	GetAllUrls() (goods, bads []string)
	MarkBad(url string)
	MarkAsGood(url string)
}

func RunActiveHealthCheck(ctx context.Context, cfg *ActiveHealthCheckConfig) {
	if cfg == nil || cfg.Balancer == nil {
		return
	}

	// TODO add overflow in cfg.Interval and cfg.Timeout

	cfg.Logger = cfg.Logger.With(
		slog.String("service", "active healthcheck"))

	cfg.Interval *= time.Second
	cfg.Timeout *= time.Millisecond
	cfg.Logger.Info(
		fmt.Sprintf("active health check will check urls each %s, with timeout %s",
			cfg.Interval, cfg.Timeout))

	ticker := time.NewTicker(cfg.Interval) // TODO check non negative

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				goods, bads := cfg.Balancer.GetAllUrls()

				for _, url := range goods {
					url := url
					go func() {
						ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
						defer cancel()
						req, err := http.NewRequestWithContext(ctx, cfg.Method, url, nil)
						if err != nil {
							cfg.Logger.Error(err.Error())
							return // TODO ???
						}
						res, err := http.DefaultClient.Do(req)
						if err != nil || res.StatusCode >= 500 {
							cfg.Balancer.MarkBad(url)
							cfg.Logger.Info(fmt.Sprintf("url [%s] mark as bad", url))
						} else {
							cfg.Logger.Info(fmt.Sprintf("url [%s] still good", url))
						}
					}()
				}

				for _, url := range bads {
					url := url
					go func() {
						ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
						defer cancel()
						req, err := http.NewRequestWithContext(ctx, cfg.Method, url, nil)
						if err != nil {
							cfg.Logger.Error(err.Error())
							return // TODO ???
						}
						res, err := http.DefaultClient.Do(req)
						if err != nil || res.StatusCode >= 500 {
							cfg.Logger.Info(fmt.Sprintf("url [%s] still bad", url))
						} else {
							cfg.Balancer.MarkAsGood(url)
							cfg.Logger.Info(fmt.Sprintf("url [%s] mark as good", url))
						}
					}()
				}

			}
		}
	}()

}
