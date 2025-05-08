package app

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/maximmihin/cb625/internal/app/testhelpers/dummyservice/ponger/app"
)

func TestE2E(t *testing.T) {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	//pongers := runPongers(t, logger, 4)

	cfg := new(Config)
	err := json.Unmarshal([]byte(jsonConfig), cfg)
	require.NoError(t, err)

	cfg.Logger = logger
	sm, err := New(t.Context(), cfg)

	//sm, err := New(t.Context(), &Config{
	//	Balancer: &roundrobin.Config{
	//		BackendUrls: []roundrobin.ServerConfig{
	//			{URL: pongers[2]},
	//			{URL: pongers[3]},
	//		},
	//	},
	//	ActiveHealthCheck: &healthcheck.ActiveHealthCheckConfig{
	//		HealthHandler: "/ping",
	//		Method:        "GET",
	//		Interval:      10,  // 10 s
	//		Timeout:       100, // 100 ms
	//	},
	//	Limiter: &tokenbucket.Config{
	//		ClientsLimits: map[string]float64{
	//			"petya": 1.5,
	//			"vasya": 60,
	//		},
	//	},
	//	Logger: logger,
	//})
	require.NoError(t, err)

	runHttpServer(t, sm, 8080)

	t.Run("round robin", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "http://localhost:8080", nil)
		req.Header.Add("ft-header", "petya")

		for i := 0; i < 30; i++ {
			res, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			resStr, err := responseToString(res)
			assert.NoError(t, err)
			t.Log(resStr)
			time.Sleep(time.Second - (400 * time.Millisecond))
		}
	})

	// TODO check case when all backend dead
}

func runPongers(t *testing.T, logger *slog.Logger, num int) []string {
	pingers := make([]string, num)
	for i := 0; i < num; i++ {
		runed, err := app.Run(app.Config{
			ServerId: fmt.Sprintf("ponger № %d", i+1),
			Port:     "0",
			Logger:   logger,
		})
		require.NoError(t, err)
		t.Log(fmt.Sprintf("ponger %d runned on %s", i, runed))
		pingers[i] = "http://" + runed
	}
	return pingers
}

func runHttpServer(t *testing.T, handler http.Handler, port int) {
	portStr := strconv.Itoa(port)
	server := http.Server{
		Addr:    net.JoinHostPort("", portStr),
		Handler: handler,
	}

	ln, err := net.Listen("tcp", server.Addr)
	require.NoError(t, err)

	t.Log("Started at port: " + portStr)
	go func() {
		t.Log(server.Serve(ln).Error())
	}()
}

func responseToString(resp *http.Response) (string, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return string(bodyBytes), nil
}

const jsonConfig = `
{
  "balancer": {
    "type": "round_robin",
    "servers": [
      { "url": "http://localhost:8081", "weight": 3 },
      { "url": "http://localhost:8082", "weight": 2 },
      { "url": "http://localhost:8083", "weight": 1 }
    ]
  },
  "limiter": {
    "type": "token_buket",
    "users": [
      { "token": "vasya", "limit":  1.5 },
      { "token": "petya", "limit":  10 }
    ]
  },
  "active_health_check": {
    "handler": "/ping",
    "method": "GET",
    "interval": 10,
    "timeout": 100
  }
}`
