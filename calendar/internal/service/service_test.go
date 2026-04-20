package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ustithegod/wb-level-4/calendar/internal/domain"
	"github.com/ustithegod/wb-level-4/calendar/internal/repository/memory"
	"github.com/ustithegod/wb-level-4/calendar/internal/service"
)

type schedulerStub struct {
	scheduled []domain.Event
	cancelled []int64
}

func (s *schedulerStub) Schedule(event domain.Event) {
	s.scheduled = append(s.scheduled, event)
}

func (s *schedulerStub) Cancel(eventID int64) {
	s.cancelled = append(s.cancelled, eventID)
}

func TestCreateEvent(t *testing.T) {
	repo := memory.New()
	scheduler := &schedulerStub{}
	svc := service.New(repo, scheduler)
	svc.SetNow(func() time.Time { return time.Date(2026, 4, 21, 9, 0, 0, 0, time.UTC) })

	event, err := svc.CreateEvent(context.Background(), service.CreateEventInput{
		UserID:   1,
		Date:     "2026-04-23",
		Title:    "doctor",
		RemindAt: "2026-04-22T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("CreateEvent returned error: %v", err)
	}

	if event.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if event.Title != "doctor" {
		t.Fatalf("unexpected title: %s", event.Title)
	}
	if len(scheduler.scheduled) != 1 {
		t.Fatalf("expected reminder to be scheduled once, got %d", len(scheduler.scheduled))
	}
}

func TestCreateEventInvalidDate(t *testing.T) {
	svc := service.New(memory.New(), nil)

	_, err := svc.CreateEvent(context.Background(), service.CreateEventInput{
		UserID: 1,
		Date:   "21-04-2026",
		Title:  "doctor",
	})
	if !errors.Is(err, domain.ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}
}

func TestDeleteMissingEvent(t *testing.T) {
	svc := service.New(memory.New(), nil)

	err := svc.DeleteEvent(context.Background(), 42)
	if !errors.Is(err, domain.ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestEventsForDayWeekMonth(t *testing.T) {
	svc := service.New(memory.New(), nil)

	inputs := []service.CreateEventInput{
		{UserID: 1, Date: "2026-04-20", Title: "monday"},
		{UserID: 1, Date: "2026-04-21", Title: "tuesday"},
		{UserID: 1, Date: "2026-04-30", Title: "month-end"},
		{UserID: 1, Date: "2026-05-01", Title: "next-month"},
		{UserID: 2, Date: "2026-04-21", Title: "other-user"},
	}
	for _, input := range inputs {
		if _, err := svc.CreateEvent(context.Background(), input); err != nil {
			t.Fatalf("CreateEvent returned error: %v", err)
		}
	}

	dayEvents, err := svc.EventsForDay(context.Background(), 1, "2026-04-21")
	if err != nil {
		t.Fatalf("EventsForDay returned error: %v", err)
	}
	if len(dayEvents) != 1 {
		t.Fatalf("expected 1 day event, got %d", len(dayEvents))
	}

	weekEvents, err := svc.EventsForWeek(context.Background(), 1, "2026-04-21")
	if err != nil {
		t.Fatalf("EventsForWeek returned error: %v", err)
	}
	if len(weekEvents) != 2 {
		t.Fatalf("expected 2 week events, got %d", len(weekEvents))
	}

	monthEvents, err := svc.EventsForMonth(context.Background(), 1, "2026-04-21")
	if err != nil {
		t.Fatalf("EventsForMonth returned error: %v", err)
	}
	if len(monthEvents) != 3 {
		t.Fatalf("expected 3 month events, got %d", len(monthEvents))
	}
}

func TestArchiveOldEvents(t *testing.T) {
	repo := memory.New()
	scheduler := &schedulerStub{}
	svc := service.New(repo, scheduler)
	svc.SetNow(func() time.Time { return time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC) })

	oldEvent, err := svc.CreateEvent(context.Background(), service.CreateEventInput{
		UserID:   1,
		Date:     "2026-04-20",
		Title:    "past",
		RemindAt: "2026-04-20T09:00:00Z",
	})
	if err != nil {
		t.Fatalf("CreateEvent returned error: %v", err)
	}
	if _, err := svc.CreateEvent(context.Background(), service.CreateEventInput{
		UserID: 1,
		Date:   "2026-04-21",
		Title:  "today",
	}); err != nil {
		t.Fatalf("CreateEvent returned error: %v", err)
	}

	archived, err := svc.ArchiveOldEvents(context.Background())
	if err != nil {
		t.Fatalf("ArchiveOldEvents returned error: %v", err)
	}

	if len(archived) != 1 {
		t.Fatalf("expected 1 archived event, got %d", len(archived))
	}
	if archived[0].ID != oldEvent.ID {
		t.Fatalf("unexpected archived event ID: %d", archived[0].ID)
	}
	if len(scheduler.cancelled) != 1 || scheduler.cancelled[0] != oldEvent.ID {
		t.Fatalf("expected archived reminder to be cancelled, got %#v", scheduler.cancelled)
	}
}
