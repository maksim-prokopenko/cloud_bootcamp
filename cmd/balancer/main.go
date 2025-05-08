package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/maximmihin/cb625/internal/servicemanager"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // todo get from env
	}))

	cfg, err := servicemanager.ParseFromJSON(os.Getenv("CONFIG_PATH"))
	if err != nil {
		logger.Error(err.Error())
		return
	}

	sm, err := servicemanager.NewServiceManager(cfg, logger)
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
