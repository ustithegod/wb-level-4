package service

import (
	"context"
	"errors"
	"time"

	"github.com/ustithegod/wb-level-4/calendar/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, event domain.Event) (domain.Event, error)
	Update(ctx context.Context, event domain.Event) (domain.Event, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (domain.Event, error)
	ListByRange(ctx context.Context, userID int64, from, to time.Time) ([]domain.Event, error)
	ArchiveBefore(ctx context.Context, before time.Time) ([]domain.Event, error)
}

type ReminderScheduler interface {
	Schedule(event domain.Event)
	Cancel(eventID int64)
}

type Service struct {
	repo      Repository
	scheduler ReminderScheduler
	now       func() time.Time
}

type CreateEventInput struct {
	UserID   int64
	Date     string
	Title    string
	RemindAt string
}

type UpdateEventInput struct {
	ID       int64
	UserID   int64
	Date     string
	Title    string
	RemindAt *string
}

func New(repo Repository, scheduler ReminderScheduler) *Service {
	return &Service{
		repo:      repo,
		scheduler: scheduler,
		now:       time.Now,
	}
}

func (s *Service) SetNow(fn func() time.Time) {
	s.now = fn
}

func (s *Service) CreateEvent(ctx context.Context, input CreateEventInput) (domain.Event, error) {
	date, remindAt, err := validatePayload(input.UserID, input.Date, input.Title, input.RemindAt)
	if err != nil {
		return domain.Event{}, err
	}

	now := s.now().UTC()
	event, err := s.repo.Create(ctx, domain.Event{
		UserID:   input.UserID,
		Date:     date,
		Title:    input.Title,
		RemindAt: remindAt,
		Created:  now,
		Updated:  now,
	})
	if err != nil {
		return domain.Event{}, err
	}

	if event.RemindAt != nil && s.scheduler != nil {
		s.scheduler.Schedule(event)
	}

	return event, nil
}

func (s *Service) UpdateEvent(ctx context.Context, input UpdateEventInput) (domain.Event, error) {
	if input.ID <= 0 {
		return domain.Event{}, domain.ErrEventNotFound
	}

	existing, err := s.repo.GetByID(ctx, input.ID)
	if err != nil {
		return domain.Event{}, err
	}

	date, remindAt, err := validatePayload(input.UserID, input.Date, input.Title, optionalString(input.RemindAt))
	if err != nil {
		return domain.Event{}, err
	}

	existing.UserID = input.UserID
	existing.Date = date
	existing.Title = input.Title
	existing.RemindAt = remindAt
	existing.Updated = s.now().UTC()

	event, err := s.repo.Update(ctx, existing)
	if err != nil {
		return domain.Event{}, err
	}

	if s.scheduler != nil {
		if event.RemindAt != nil {
			s.scheduler.Schedule(event)
		} else {
			s.scheduler.Cancel(event.ID)
		}
	}

	return event, nil
}

func (s *Service) DeleteEvent(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrEventNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if s.scheduler != nil {
		s.scheduler.Cancel(id)
	}

	return nil
}

func (s *Service) EventsForDay(ctx context.Context, userID int64, date string) ([]domain.Event, error) {
	start, err := parseDate(date)
	if err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, domain.ErrInvalidUserID
	}

	return s.repo.ListByRange(ctx, userID, start, start.AddDate(0, 0, 1))
}

func (s *Service) EventsForWeek(ctx context.Context, userID int64, date string) ([]domain.Event, error) {
	base, err := parseDate(date)
	if err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, domain.ErrInvalidUserID
	}

	weekday := int(base.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := base.AddDate(0, 0, -(weekday - 1))
	return s.repo.ListByRange(ctx, userID, start, start.AddDate(0, 0, 7))
}

func (s *Service) EventsForMonth(ctx context.Context, userID int64, date string) ([]domain.Event, error) {
	base, err := parseDate(date)
	if err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, domain.ErrInvalidUserID
	}

	start := time.Date(base.Year(), base.Month(), 1, 0, 0, 0, 0, time.UTC)
	return s.repo.ListByRange(ctx, userID, start, start.AddDate(0, 1, 0))
}

func (s *Service) ArchiveOldEvents(ctx context.Context) ([]domain.Event, error) {
	cutoff := beginningOfDay(s.now().UTC())
	events, err := s.repo.ArchiveBefore(ctx, cutoff)
	if err != nil {
		return nil, err
	}

	if s.scheduler != nil {
		for _, event := range events {
			s.scheduler.Cancel(event.ID)
		}
	}

	return events, nil
}

func parseDate(raw string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, domain.ErrInvalidDate
	}
	return date.UTC(), nil
}

func parseReminder(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}

	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04"} {
		ts, err := time.Parse(layout, raw)
		if err == nil {
			value := ts.UTC()
			return &value, nil
		}
	}

	return nil, errors.New("invalid remind_at")
}

func validatePayload(userID int64, dateRaw, title, remindRaw string) (time.Time, *time.Time, error) {
	if userID <= 0 {
		return time.Time{}, nil, domain.ErrInvalidUserID
	}
	if title == "" {
		return time.Time{}, nil, domain.ErrEmptyTitle
	}

	date, err := parseDate(dateRaw)
	if err != nil {
		return time.Time{}, nil, err
	}

	remindAt, err := parseReminder(remindRaw)
	if err != nil {
		return time.Time{}, nil, err
	}

	return date, remindAt, nil
}

func beginningOfDay(ts time.Time) time.Time {
	return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.UTC)
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
