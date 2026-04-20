package worker

import (
	"context"
	"sync"
	"time"

	"github.com/ustithegod/wb-level-4/calendar/internal/domain"
	"github.com/ustithegod/wb-level-4/calendar/internal/logger"
)

type reminderCommand struct {
	event *domain.Event
	id    int64
}

type ReminderWorker struct {
	log    *logger.AsyncLogger
	cmdCh  chan reminderCommand
	timers map[int64]*time.Timer
	mu     sync.Mutex
}

func NewReminderWorker(log *logger.AsyncLogger, buffer int) *ReminderWorker {
	return &ReminderWorker{
		log:    log,
		cmdCh:  make(chan reminderCommand, buffer),
		timers: make(map[int64]*time.Timer),
	}
}

func (w *ReminderWorker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.stopAll()
			return
		case cmd := <-w.cmdCh:
			if cmd.event != nil {
				w.schedule(*cmd.event)
				continue
			}
			w.cancel(cmd.id)
		}
	}
}

func (w *ReminderWorker) Schedule(event domain.Event) {
	if event.RemindAt == nil {
		return
	}

	select {
	case w.cmdCh <- reminderCommand{event: &event}:
	default:
	}
}

func (w *ReminderWorker) Cancel(eventID int64) {
	select {
	case w.cmdCh <- reminderCommand{id: eventID}:
	default:
	}
}

func (w *ReminderWorker) schedule(event domain.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if timer, ok := w.timers[event.ID]; ok {
		timer.Stop()
		delete(w.timers, event.ID)
	}

	delay := time.Until(*event.RemindAt)
	if delay < 0 {
		delay = 0
	}

	w.timers[event.ID] = time.AfterFunc(delay, func() {
		w.log.Info(
			context.Background(),
			"reminder_sent",
			"event_id", event.ID,
			"user_id", event.UserID,
			"title", event.Title,
			"remind_at", event.RemindAt.Format(time.RFC3339),
		)

		w.mu.Lock()
		delete(w.timers, event.ID)
		w.mu.Unlock()
	})
}

func (w *ReminderWorker) cancel(eventID int64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	timer, ok := w.timers[eventID]
	if !ok {
		return
	}
	timer.Stop()
	delete(w.timers, eventID)
}

func (w *ReminderWorker) stopAll() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for id, timer := range w.timers {
		timer.Stop()
		delete(w.timers, id)
	}
}
