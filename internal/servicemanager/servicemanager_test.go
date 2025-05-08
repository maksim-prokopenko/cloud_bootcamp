package servicemanager

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/maximmihin/cb625/internal/balancer/roundrobin"
	weightedroundrobin "github.com/maximmihin/cb625/internal/balancer/roundrobin/weighted"
	"github.com/maximmihin/cb625/internal/servicemanager/testhelpers/dummyservice/ponger/app"
)

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

func TestE2E(t *testing.T) {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // todo get from env
	}))

	pongers := runPongers(t, logger, 4)

	{

		sm, err := NewServiceManager(&ServiceManagerConfig{
			Services: map[string]any{
				"web_app": &weightedroundrobin.Config{
					ServerName: "http://web1.local",
					BackendUrls: []weightedroundrobin.ServerConfig{
						{URL: pongers[0], Weight: 3},
						{URL: pongers[1], Weight: 2},
					},
				},
				"offline_app": &roundrobin.Config{
					ServerName: "http://offline1.local",
					BackendUrls: []roundrobin.ServerConfig{
						{URL: pongers[2]},
						{URL: pongers[3]},
					},
				},
				//"auth_service": &roundrobin.Config{
				//	ActiveHealthCheck: &healthcheck.ActiveHealthCheckConfig{
				//		HealthHandler: "/ping",
				//		Method:        "GET",
				//		Interval:      10,
				//		Timeout:       1,
				//	},
				//	BackendUrls: []roundrobin.ServerConfig{
				//		{URL: "http://auth1:9090"},
				//		{URL: "http://auth2:9090"},
				//	},
				//},
			},
		}, logger)
		require.NoError(t, err)

		runHttpServer(t, sm, 8080)
	}

	http.DefaultClient.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(network, "127.0.0.1:8080")
		},
	}

	t.Run("round robin", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://web1.local", nil)
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			_ = res
		}
	})

	t.Run("weighted round robin", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://offline1.local", nil)
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			_ = res
		}
	})

}
