package memory

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/ustithegod/wb-level-4/calendar/internal/domain"
)

type Repository struct {
	mu     sync.RWMutex
	nextID int64
	events map[int64]domain.Event
}

func New() *Repository {
	return &Repository{
		nextID: 1,
		events: make(map[int64]domain.Event),
	}
}

func (r *Repository) Create(_ context.Context, event domain.Event) (domain.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	event.ID = r.nextID
	r.nextID++
	r.events[event.ID] = event

	return event, nil
}

func (r *Repository) Update(_ context.Context, event domain.Event) (domain.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.events[event.ID]; !ok {
		return domain.Event{}, domain.ErrEventNotFound
	}

	r.events[event.ID] = event
	return event, nil
}

func (r *Repository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.events[id]; !ok {
		return domain.ErrEventNotFound
	}

	delete(r.events, id)
	return nil
}

func (r *Repository) GetByID(_ context.Context, id int64) (domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	event, ok := r.events[id]
	if !ok {
		return domain.Event{}, domain.ErrEventNotFound
	}

	return event, nil
}

func (r *Repository) ListByRange(_ context.Context, userID int64, from, to time.Time) ([]domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]domain.Event, 0)
	for _, event := range r.events {
		if event.UserID != userID {
			continue
		}
		if !event.Date.Before(from) && event.Date.Before(to) {
			events = append(events, event)
		}
	}

	slices.SortFunc(events, func(a, b domain.Event) int {
		if a.Date.Before(b.Date) {
			return -1
		}
		if a.Date.After(b.Date) {
			return 1
		}
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})

	return events, nil
}

func (r *Repository) ArchiveBefore(_ context.Context, before time.Time) ([]domain.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	archived := make([]domain.Event, 0)
	for id, event := range r.events {
		if event.Date.Before(before) {
			archived = append(archived, event)
			delete(r.events, id)
		}
	}

	slices.SortFunc(archived, func(a, b domain.Event) int {
		if a.Date.Before(b.Date) {
			return -1
		}
		if a.Date.After(b.Date) {
			return 1
		}
		return 0
	})

	return archived, nil
}
