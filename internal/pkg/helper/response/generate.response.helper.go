package responsehelper

import (
	"time"

	"vibe-ddd-golang/internal/common/enum"
	types "vibe-ddd-golang/internal/common/type"
	"vibe-ddd-golang/internal/pkg/reqctx"

	"github.com/gin-gonic/gin"
)

// defaultVersion is used in debug.version when the API-version middleware has
// not selected one (e.g. probes, pre-version-middleware failures).
const defaultVersion = "v1.0.0"

func BuildResponse(c *gin.Context, r *types.Response, exposeRawError bool) *types.ResponseAPI {
	if r.Code == "" {
		r.Code = enum.DefaultResultCode(r.HTTPStatus)
	}
	if r.Message == "" {
		r.Message = DefaultMessage(r.Code)
	}

	version := reqctx.SelectedVersion(c)
	if version == "" {
		version = defaultVersion
	}

	var errStr *string
	if exposeRawError && r.Err != nil {
		s := r.Err.Error()
		errStr = &s
	}

	start := reqctx.StartTime(c)
	end := time.Now()

	return &types.ResponseAPI{
		RequestID: reqctx.RequestID(c),
		Code:      r.Code,
		Message:   r.Message,
		Debug: &types.ResponseAPIDebug{
			Version:   version,
			Error:     errStr,
			StartTime: start,
			EndTime:   end,
			RuntimeMs: end.Sub(start).Milliseconds(),
		},
		Data: r.Data,
	}
}

// DefaultMessage returns a safe human message for a result code.
func DefaultMessage(code enum.ResultCode) string {
	switch code {
	case enum.CodeOK:
		return "Request completed successfully."
	case enum.CodeCreated:
		return "Resource created."
	case enum.CodeAccepted:
		return "Request accepted for processing."
	case enum.CodeInvalidPayload:
		return "Payload validation failed."
	case enum.CodeInvalidEncryptedRequest:
		return "Invalid encrypted request."
	case enum.CodeUnsupportedAPIVersion:
		return "Unsupported API version."
	case enum.CodeUnauthorized:
		return "Authentication required."
	case enum.CodeBadRequest:
		return "Bad request."
	case enum.CodeForbidden:
		return "Not allowed."
	case enum.CodeNotFound:
		return "Resource not found."
	case enum.CodeConflict:
		return "State conflict."
	case enum.CodeIdempotencyConflict:
		return "Idempotency key reused with a different payload."
	case enum.CodeRateLimited:
		return "Rate limit exceeded."
	case enum.CodeUnprocessableEntity:
		return "Request could not be processed."
	case enum.CodeProviderUnavailable:
		return "Provider unavailable."
	case enum.CodeServiceUnavailable:
		return "Service temporarily unavailable."
	case enum.CodeTimeout:
		return "Operation timed out."
	default:
		return "Internal server error."
	}
}
