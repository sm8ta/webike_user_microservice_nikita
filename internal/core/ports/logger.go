package ports

import "context"

type LoggerPort interface {
	// Methods for all project
	Info(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Debug(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})

	// Methods for gRPC
	InfoGRPC(ctx context.Context, msg string, fields any)
	ErrorGRPC(ctx context.Context, msg string, fields any)
	DebugGRPC(ctx context.Context, msg string, fields any)
	WarnGRPC(ctx context.Context, msg string, fields any)
}
