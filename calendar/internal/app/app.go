package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/ustithegod/wb-level-4/calendar/internal/config"
	httphandler "github.com/ustithegod/wb-level-4/calendar/internal/http/handler"
	httpmiddleware "github.com/ustithegod/wb-level-4/calendar/internal/http/middleware"
	"github.com/ustithegod/wb-level-4/calendar/internal/logger"
	"github.com/ustithegod/wb-level-4/calendar/internal/repository/memory"
	"github.com/ustithegod/wb-level-4/calendar/internal/service"
	"github.com/ustithegod/wb-level-4/calendar/internal/worker"
)

type App struct {
	server  *http.Server
	async   *logger.AsyncLogger
	cancel  context.CancelFunc
	stopped chan struct{}
}

func New(cfg config.Config) *App {
	baseLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	asyncLogger := logger.NewAsync(baseLogger, cfg.LogBuffer)

	ctx, cancel := context.WithCancel(context.Background())

	repo := memory.New()
	reminders := worker.NewReminderWorker(asyncLogger, cfg.LogBuffer)
	svc := service.New(repo, reminders)

	handler := httphandler.New(svc)
	root := handler.Router()

	// Wrap the chi router with generic middlewares at the top level.
	serverHandler := http.Handler(root)
	serverHandler = chimiddleware.RequestID(serverHandler)
	serverHandler = chimiddleware.Recoverer(serverHandler)
	serverHandler = chimiddleware.Timeout(10 * time.Second)(serverHandler)
	serverHandler = httpmiddleware.RequestLogger(asyncLogger)(serverHandler)

	archiveRunner := worker.NewArchiveRunner(cfg.ArchiveInterval, asyncLogger, func(ctx context.Context) (int, error) {
		events, err := svc.ArchiveOldEvents(ctx)
		if err != nil {
			return 0, err
		}
		return len(events), nil
	})

	go asyncLogger.Run(ctx)
	go reminders.Run(ctx)
	go archiveRunner.Run(ctx)

	return &App{
		server: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: serverHandler,
		},
		async:   asyncLogger,
		cancel:  cancel,
		stopped: make(chan struct{}),
	}
}

func (a *App) Run() error {
	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	defer close(a.stopped)

	a.cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	a.async.Close()
	return nil
}
