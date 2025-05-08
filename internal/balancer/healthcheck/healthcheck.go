package healthcheck

import (
	"context"
	"net/http"
	"time"
)

type ActiveHealthCheckConfig struct {
	//HealthHandler url.URL       `json:"health_handler"` // TODO thange format
	HealthHandler string        `json:"health_handler"`
	Method        string        `json:"method"`
	Interval      time.Duration `json:"interval"` // TODO thange format
	Timeout       time.Duration `json:"timeout"`
}

// TODO name difer from real smartCarousel
type SmartCarousel interface {
	GetAllUrls() []string
	MarkBad(url string)
}

// TODO think about stop
func RunActiveHealthCheck(cfg *ActiveHealthCheckConfig, carousel SmartCarousel) {
	if cfg == nil || carousel == nil {
		return
	}

	ctx := context.TODO()
	ticker := time.NewTicker(time.Duration(cfg.Interval)) // TODO check nonnegative

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				urls := carousel.GetAllUrls()
				for _, url := range urls {
					url := url
					go func() {
						ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Timeout))
						defer cancel()
						req, err := http.NewRequestWithContext(ctx, cfg.Method, url, nil)
						if err != nil {
							return // TODO ???
						}
						res, err := http.DefaultClient.Do(req)
						if err != nil || res.StatusCode >= 500 {
							carousel.MarkBad(url) // TODO log about it?
						}
					}()
				}
			}
		}
	}()

}
