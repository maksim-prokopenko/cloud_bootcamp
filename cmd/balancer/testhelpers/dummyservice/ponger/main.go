package main

import (
	"fmt"
	"github.com/Pallinder/sillyname-go"
	"log/slog"
	"net"
	"net/http"
	"os"
)

func main() {

	cfg := Config{
		ServerId: os.Getenv("SERVER_ID"),
		Port:     os.Getenv("SERVER_PORT"),
	}

	Run(cfg)
}

type Config struct {
	ServerId string
	Port     string
	Logger   *slog.Logger
}

func Run(cfg Config) {

	cfg = cfg.WithDefaults()

	logger := cfg.Logger.With(
		slog.String("server id", cfg.ServerId))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(fmt.Sprintf("request from %s", r.RemoteAddr))
		w.WriteHeader(200)
		_, _ = fmt.Fprint(w, "pong")
	})

	err := http.ListenAndServe(net.JoinHostPort("", cfg.Port), nil)
	if err != nil {
		logger.Error(err.Error())
	}
}

func (c Config) WithDefaults() Config {

	if len(c.ServerId) == 0 {
		c.ServerId = sillyname.GenerateStupidName()
	}

	if len(c.Port) == 0 {
		c.Port = "80"
	}

	if c.Logger == nil {
		c.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	return c
}
