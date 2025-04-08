package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	_ "test-task/wallet/docs"
	"test-task/wallet/internal/app"
	"time"
)

func main() {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	myapp, err := app.New()
	if err != nil {
		slog.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}

	go myapp.Run()

	<-stop
	slog.Info("received termination signal")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	myapp.Stop(shutdownCtx)
}
