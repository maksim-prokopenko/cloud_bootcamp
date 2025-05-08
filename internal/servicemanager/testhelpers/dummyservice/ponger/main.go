package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/maximmihin/cb625/internal/servicemanager/testhelpers/dummyservice/ponger/app"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	_, err := app.Run(app.Config{
		ServerId: os.Getenv("SERVER_ID"),
		Port:     os.Getenv("SERVER_PORT"),
		Logger:   logger,
	})
	if err != nil {
		logger.Error(err.Error())
	}

	time.Sleep(1 * time.Hour) // TODO )
}
