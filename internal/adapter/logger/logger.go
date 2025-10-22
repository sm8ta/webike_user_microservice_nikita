package logger

import (
	"context"
	"log/slog"
	"os"
	"webike_services/webike_User-microservice_Nikita/internal/core/ports"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type LoggerAdapter struct {
	logger *slog.Logger
}

func NewLoggerAdapter(env string) ports.LoggerPort {
	var log *slog.Logger

	switch env {
	case envLocal, envDev:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	}

	return &LoggerAdapter{
		logger: log,
	}
}

func (l *LoggerAdapter) Info(msg string, fields map[string]interface{}) {
	if fields == nil {
		l.logger.Info(msg)
		return
	}
	l.logger.Info(msg, slog.Any("fields", fields))
}

func (l *LoggerAdapter) Error(msg string, fields map[string]interface{}) {
	if fields == nil {
		l.logger.Error(msg)
		return
	}
	l.logger.Error(msg, slog.Any("fields", fields))
}

func (l *LoggerAdapter) Debug(msg string, fields map[string]interface{}) {
	if fields == nil {
		l.logger.Debug(msg)
		return
	}
	l.logger.Debug(msg, slog.Any("fields", fields))
}

func (l *LoggerAdapter) Warn(msg string, fields map[string]interface{}) {
	if fields == nil {
		l.logger.Warn(msg)
		return
	}
	l.logger.Warn(msg, slog.Any("fields", fields))
}

func (l *LoggerAdapter) InfoGRPC(ctx context.Context, msg string, fields any) {
	if fields == nil {
		l.logger.InfoContext(ctx, msg)
		return
	}
	l.logger.InfoContext(ctx, msg, slog.Any("fields", fields))
}

func (l *LoggerAdapter) ErrorGRPC(ctx context.Context, msg string, fields any) {
	if fields == nil {
		l.logger.ErrorContext(ctx, msg)
		return
	}
	l.logger.ErrorContext(ctx, msg, slog.Any("fields", fields))
}

func (l *LoggerAdapter) DebugGRPC(ctx context.Context, msg string, fields any) {
	if fields == nil {
		l.logger.DebugContext(ctx, msg)
		return
	}
	l.logger.DebugContext(ctx, msg, slog.Any("fields", fields))
}

func (l *LoggerAdapter) WarnGRPC(ctx context.Context, msg string, fields any) {
	if fields == nil {
		l.logger.WarnContext(ctx, msg)
		return
	}
	l.logger.WarnContext(ctx, msg, slog.Any("fields", fields))
}
