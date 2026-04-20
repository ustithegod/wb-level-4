package logger

import (
	"context"
	"log/slog"
)

type Record struct {
	Level slog.Level
	Msg   string
	Attrs []any
}

type AsyncLogger struct {
	base   *slog.Logger
	ch     chan Record
	closed chan struct{}
}

func NewAsync(base *slog.Logger, buffer int) *AsyncLogger {
	return &AsyncLogger{
		base:   base,
		ch:     make(chan Record, buffer),
		closed: make(chan struct{}),
	}
}

func (l *AsyncLogger) Run(ctx context.Context) {
	defer close(l.closed)

	for {
		select {
		case <-ctx.Done():
			l.drain()
			return
		case rec := <-l.ch:
			l.base.Log(context.Background(), rec.Level, rec.Msg, rec.Attrs...)
		}
	}
}

func (l *AsyncLogger) drain() {
	for {
		select {
		case rec := <-l.ch:
			l.base.Log(context.Background(), rec.Level, rec.Msg, rec.Attrs...)
		default:
			return
		}
	}
}

func (l *AsyncLogger) Close() {
	<-l.closed
}

func (l *AsyncLogger) Log(ctx context.Context, level slog.Level, msg string, attrs ...any) {
	rec := Record{Level: level, Msg: msg, Attrs: attrs}

	select {
	case l.ch <- rec:
	case <-ctx.Done():
	default:
		// Drop log records when the buffer is full to avoid blocking handlers.
	}
}

func (l *AsyncLogger) Info(ctx context.Context, msg string, attrs ...any) {
	l.Log(ctx, slog.LevelInfo, msg, attrs...)
}

func (l *AsyncLogger) Error(ctx context.Context, msg string, attrs ...any) {
	l.Log(ctx, slog.LevelError, msg, attrs...)
}
