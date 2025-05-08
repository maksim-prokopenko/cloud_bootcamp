package app

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
)

type Config struct {
	ServerId string
	Port     string
	Logger   *slog.Logger
}

func Run(cfg Config) (string, error) {

	cfg = cfg.WithDefaults()

	logger := cfg.Logger.With(
		slog.String("server id", cfg.ServerId))

	logger.Info("config",
		slog.Any("config", cfg))

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(fmt.Sprintf("request from %s", r.RemoteAddr))
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, "pong from "+cfg.ServerId)
	})

	server := &http.Server{
		Addr:    net.JoinHostPort("0.0.0.0", cfg.Port),
		Handler: mux,
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		logger.Error(err.Error())
		return "", err
	}

	go func() {
		logger.Error(server.Serve(ln).Error())
		defer ln.Close()
	}()

	return ln.Addr().String(), err

}

func (c Config) WithDefaults() Config {

	if len(c.ServerId) == 0 {
		c.ServerId = GenerateStupidName()
	}

	if len(c.Port) == 0 {
		c.Port = "80"
	}

	if c.Logger == nil {
		c.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	return c
}
