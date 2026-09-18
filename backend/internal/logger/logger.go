// Package logger provides structured logging for the DefendCore platform.
//
// It wraps the standard library's log/slog with a configuration that
// adapts based on the runtime environment (development vs production).
package logger

import (
"context"
"log/slog"
"os"
"strings"
)

// Env represents the runtime environment.
type Env string

const (
EnvDevelopment Env = "development"
EnvStaging     Env = "staging"
EnvProduction  Env = "production"
)

// Log is the package-level logger. It is set by Init and should be used
// throughout the application via the exported helper functions.
var Log *slog.Logger = slog.Default()

// Init configures the global logger based on the provided environment.
//
// In development, it uses a human-readable text handler with debug level.
// In production, it uses JSON output with info level for log aggregation.
func Init(env Env, serviceName string) {
var handler slog.Handler

opts := &slog.HandlerOptions{
AddSource: env == EnvDevelopment,
Level:     levelFor(env),
}

if env == EnvDevelopment {
handler = slog.NewTextHandler(os.Stdout, opts)
} else {
handler = slog.NewJSONHandler(os.Stdout, opts)
}

Log = slog.New(handler).With(slog.String("service", serviceName))
}

func levelFor(env Env) slog.Level {
switch env {
case EnvDevelopment:
return slog.LevelDebug
case EnvStaging:
return slog.LevelInfo
case EnvProduction:
return slog.LevelInfo
default:
return slog.LevelInfo
}
}

// With returns a logger with additional attributes attached.
func With(args ...any) *slog.Logger {
return Log.With(args...)
}

// WithContext returns a logger with contextual attributes (request ID, user ID).
func WithContext(ctx context.Context) *slog.Logger {
if ctx == nil {
return Log
}

attrs := make([]any, 0, 4)
if v := ctx.Value("request_id"); v != nil {
attrs = append(attrs, slog.String("request_id", v.(string)))
}
if v := ctx.Value("user_id"); v != nil {
attrs = append(attrs, slog.String("user_id", v.(string)))
}
if v := ctx.Value("org_id"); v != nil {
attrs = append(attrs, slog.String("org_id", v.(string)))
}
return Log.With(attrs...)
}

// Debug logs at debug level with context.
func Debug(ctx context.Context, msg string, args ...any) {
WithContext(ctx).Debug(msg, args...)
}

// Info logs at info level with context.
func Info(ctx context.Context, msg string, args ...any) {
WithContext(ctx).Info(msg, args...)
}

// Warn logs at warn level with context.
func Warn(ctx context.Context, msg string, args ...any) {
WithContext(ctx).Warn(msg, args...)
}

// Error logs at error level with context.
func Error(ctx context.Context, msg string, args ...any) {
WithContext(ctx).Error(msg, args...)
}

// Fatal logs at error level and exits the program.
func Fatal(ctx context.Context, msg string, args ...any) {
WithContext(ctx).Error(msg, args...)
os.Exit(1)
}

// Err is a helper to attach an error to a log record.
func Err(err error) slog.Attr {
if err == nil {
return slog.String("error", "")
}
return slog.String("error", err.Error())
}

// Redact masks sensitive strings for logging (e.g., API keys).
func Redact(s string) string {
if len(s) <= 8 {
return strings.Repeat("*", len(s))
}
return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}
