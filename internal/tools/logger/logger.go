package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"PWZ1.0/internal/models/domainErrors"
)

var Logger *slog.Logger

func InitLogger() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Logger = slog.New(handler)
}

func LogErrorWithCode(ctx context.Context, err error, message string) {
	errCode := domainErrors.ErrorCodes[err]

	formatted := fmt.Sprintf("ERROR: %s: %s", errCode, message)
	Logger.ErrorContext(ctx, formatted)
}
