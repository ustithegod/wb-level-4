package worker

import (
	"context"
	"time"

	"github.com/ustithegod/wb-level-4/calendar/internal/logger"
)

type archiveService interface {
	ArchiveOldEvents(ctx context.Context) ([]any, error)
}

type ArchiveRunner struct {
	interval time.Duration
	run      func(ctx context.Context) (int, error)
	log      *logger.AsyncLogger
}

func NewArchiveRunner(interval time.Duration, log *logger.AsyncLogger, run func(ctx context.Context) (int, error)) *ArchiveRunner {
	return &ArchiveRunner{
		interval: interval,
		run:      run,
		log:      log,
	}
}

func (r *ArchiveRunner) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := r.run(ctx)
			if err != nil {
				r.log.Error(ctx, "archive_failed", "error", err.Error())
				continue
			}
			if count > 0 {
				r.log.Info(ctx, "archive_completed", "archived", count)
			}
		}
	}
}
