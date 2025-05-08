package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/maximmihin/cb625/internal/app"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // todo get from env
	}))

	cfg, err := app.ParseFromJSON(os.Getenv("CONFIG_PATH"))
	if err != nil {
		logger.Error(err.Error())
		return
	}
	logger.Debug("config after parse",
		slog.Any("config", cfg))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg.Logger = logger
	sm, err := app.New(ctx, cfg)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	server := http.Server{
		Addr:    ":8080",
		Handler: sm,
	}

	logger.Info("Started at port: 8080")
	logger.Error(server.ListenAndServe().Error()) // todo separate listen and serve?
}
