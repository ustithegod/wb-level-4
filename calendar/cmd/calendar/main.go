package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ustithegod/wb-level-4/calendar/internal/app"
	"github.com/ustithegod/wb-level-4/calendar/internal/config"
)

func main() {
	cfg := config.MustLoad()
	application := app.New(cfg)

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Run()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Printf("received signal %s, shutting down", sig)
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
