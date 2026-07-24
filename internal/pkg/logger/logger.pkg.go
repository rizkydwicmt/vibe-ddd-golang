package logger

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"vibe-ddd-golang/internal/pkg/reqctx"
)

var (
	Debug      *log.Logger
	Info       *log.Logger
	Warning    *log.Logger
	Error      *log.Logger
	HTTP       *log.Logger
	errorTrace *log.Logger

	slogger *slog.Logger
)

// Setup initializes the package-level loggers. Call once at process start
// (registered via fx.Invoke(infrastructure.InitializeLogger)).
func Setup() {
	HTTP = log.New(os.Stdout, "[HTTP]\t", log.Ldate|log.Ltime)
	Info = log.New(os.Stdout, "[INFO]\t", log.Ldate|log.Ltime)
	Warning = log.New(os.Stdout, "[WARNING]\t", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(os.Stdout, "[DEBUG]\t", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(os.Stdout, "[ERROR]\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorTrace = log.New(os.Stdout, "[ERROR]\t", log.Ldate|log.Ltime)

	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slogger = slog.New(&contextHandler{inner: base})
}

// Slog returns the slog.Logger, lazily initializing a default for test code
// that never calls Setup.
func Slog() *slog.Logger {
	if slogger == nil {
		slogger = slog.New(&contextHandler{
			inner: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		})
	}
	return slogger
}

func InfoCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	Slog().LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}
func WarnCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	Slog().LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}
func ErrorCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	Slog().LogAttrs(ctx, slog.LevelError, msg, attrs...)
}
func DebugCtx(ctx context.Context, msg string, attrs ...slog.Attr) {
	Slog().LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// contextHandler decorates each record with attributes pulled from the
// context. Today: request_id only. Add new low-cardinality fields here so
// call sites never have to remember to attach them.
type contextHandler struct{ inner slog.Handler }

func (h *contextHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return h.inner.Enabled(ctx, lvl)
}
func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := reqctx.RequestIDFromContext(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.inner.Handle(ctx, r)
}
func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{inner: h.inner.WithAttrs(attrs)}
}
func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{inner: h.inner.WithGroup(name)}
}

// ErrorWithStack logs a message followed by a filtered goroutine stack.
func ErrorWithStack(format string, v ...any) {
	if errorTrace == nil {
		Setup()
	}
	errorTrace.Printf("%s\n%s", fmt.Sprintf(format, v...), getStackTrace())
}

func getStackTrace() string {
	buf := make([]byte, 64*1024)
	buf = buf[:runtime.Stack(buf, false)]
	lines := strings.Split(string(buf), "\n")

	out := make([]string, 0, len(lines))
	skip := 0
	for i, line := range lines {
		if strings.Contains(line, "logger.go") ||
			strings.Contains(line, "runtime/") ||
			strings.Contains(line, "getStackTrace") ||
			strings.Contains(line, "ErrorWithStack") {
			skip = i + 1
			continue
		}
		if i <= skip {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}
