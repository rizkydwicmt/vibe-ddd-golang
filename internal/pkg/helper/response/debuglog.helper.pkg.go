package responsehelper

import (
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"

	"vibe-ddd-golang/internal/common/enum"
	"vibe-ddd-golang/internal/pkg/logger"
	"vibe-ddd-golang/internal/pkg/reqctx"

	"github.com/gin-gonic/gin"
)

// maxLogPayload caps each logged payload so a large body can't blow up the log.
const maxLogPayload = 4 << 10 // 4KB

// logHeaderWhitelist is the set of request headers safe to log. Authorization
// and Cookie are intentionally excluded.
var logHeaderWhitelist = []string{
	"Content-Type", "Accept", "User-Agent",
}

// secretField matches obvious secret-bearing JSON keys for redaction.
var secretField = regexp.MustCompile(
	`(?i)"(password|pass|token|access_token|refresh_token|secret|authorization|otp|pin|api_key|apikey)"` +
		`\s*:\s*"[^"]*"`)

// LogResponseError emits a structured error-context log for any response with
// status >= 400 — nothing on success, so it stays cheap in production. Every
// >= 400 logs metadata + request payload + response payload + cause (when
// present, e.g. AppError.WithCause); 5xx additionally emits a filtered stack
// trace. Sensitive headers and secret JSON fields are redacted; payloads are
// truncated. Call from the response emission chokepoints (ResponseInit's send
// and response.Render) so both error paths are covered.
func LogResponseError(c *gin.Context, status int, code enum.ResultCode, msg string, data any, cause error) {
	if c == nil || status < 400 {
		return
	}

	reqID := reqctx.RequestID(c)
	attrs := []slog.Attr{
		slog.String("request_id", reqID),
		slog.Int("status", status),
		slog.Any("code", code),
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.Path),
		slog.String("msg", msg),
		slog.String("headers", safeHeaders(c)),
		slog.String("respBody", redactPayload(marshalData(data))),
	}
	if cause != nil {
		attrs = append(attrs, slog.String("cause", cause.Error()))
	}
	logger.ErrorCtx(c.Request.Context(), "response_error", attrs...)

	// Server faults: add a stack trace for root-causing. 4xx are client errors —
	// metadata + cause is enough, no stack.
	if status >= 500 {
		logger.ErrorWithStack("response_error stack requestId=%s path=%s code=%v", reqID, c.Request.URL.Path, code)
	}
}

func marshalData(data any) string {
	if data == nil {
		return ""
	}
	if b, err := json.Marshal(data); err == nil {
		return string(b)
	}
	return ""
}

func safeHeaders(c *gin.Context) string {
	parts := make([]string, 0, len(logHeaderWhitelist))
	for _, h := range logHeaderWhitelist {
		if v := c.GetHeader(h); v != "" {
			parts = append(parts, h+"="+v)
		}
	}
	return "{" + strings.Join(parts, " ") + "}"
}

// redactPayload masks secret fields and truncates. Returns "-" when empty.
func redactPayload(s string) string {
	if s == "" {
		return "-"
	}
	s = secretField.ReplaceAllString(s, `"$1":"***"`)
	if len(s) > maxLogPayload {
		s = s[:maxLogPayload] + "…(truncated)"
	}
	return s
}
