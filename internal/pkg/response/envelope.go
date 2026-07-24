package response

import (
	"errors"

	types "vibe-ddd-golang/internal/common/type"
	"vibe-ddd-golang/internal/pkg/validation"

	"vibe-ddd-golang/internal/common/enum"

	"github.com/gin-gonic/gin"
)

// AppError is the unified error type handlers return upward.
type AppError struct {
	HTTPStatus int
	Code       enum.ResultCode
	Message    string
	Data       any // e.g. types.FieldErrors for validation; provider detail otherwise
	cause      error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.cause != nil {
		return e.Message + ": " + e.cause.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.cause }

// WithCause attaches an underlying error for logs / dev debug.error.
func (e *AppError) WithCause(cause error) *AppError { e.cause = cause; return e }

// WithCode overrides the result code, keeping status and message.
func (e *AppError) WithCode(code enum.ResultCode) *AppError { e.Code = code; return e }

// WithData sets the response data payload (provider detail, hints, …).
func (e *AppError) WithData(data any) *AppError { e.Data = data; return e }

// WithMeta attaches a key/value into a map under Data. Convenience for ad-hoc
// operator detail when a typed Data payload is not warranted.
func (e *AppError) WithMeta(k string, v any) *AppError {
	m, ok := e.Data.(map[string]any)
	if !ok {
		m = map[string]any{}
		e.Data = m
	}
	m[k] = v
	return e
}

func newErr(code enum.ResultCode, msg string) *AppError {
	return &AppError{HTTPStatus: code.DefaultHTTPStatus(), Code: code, Message: msg}
}

// Constructors map onto standard result codes.

func NewBadRequestException(msg string) *AppError { return newErr(enum.CodeInvalidPayload, msg) }

func NewInvalidEncryptedRequestException(msg string) *AppError {
	return newErr(enum.CodeInvalidEncryptedRequest, msg)
}
func NewUnsupportedAPIVersionException(msg string) *AppError {
	return newErr(enum.CodeUnsupportedAPIVersion, msg)
}
func NewUnauthorizedException(msg string) *AppError { return newErr(enum.CodeUnauthorized, msg) }
func NewForbiddenException(msg string) *AppError    { return newErr(enum.CodeForbidden, msg) }
func NewNotFoundException(msg string) *AppError     { return newErr(enum.CodeNotFound, msg) }
func NewConflictException(msg string) *AppError     { return newErr(enum.CodeConflict, msg) }
func NewIdempotencyConflictException(msg string) *AppError {
	return newErr(enum.CodeIdempotencyConflict, msg)
}
func NewRateLimitedException(msg string) *AppError { return newErr(enum.CodeRateLimited, msg) }
func NewUnprocessableException(msg string) *AppError {
	return newErr(enum.CodeUnprocessableEntity, msg)
}
func NewProviderUnavailableException(msg string) *AppError {
	return newErr(enum.CodeProviderUnavailable, msg)
}
func NewUnavailableException(msg string) *AppError    { return newErr(enum.CodeServiceUnavailable, msg) }
func NewTimeoutException(msg string) *AppError        { return newErr(enum.CodeTimeout, msg) }
func NewInternalServerException(msg string) *AppError { return newErr(enum.CodeInternalError, msg) }

// As branches on AppError without reflecting on concrete types.
func As(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}

	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}

	return NewInternalServerException("internal error").WithCause(err), true
}

// FromError wraps an arbitrary error. Unknown errors collapse to 500.
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}
	if ae, ok := As(err); ok {
		return ae
	}
	return NewInternalServerException("internal error").WithCause(err)
}

// Render writes the standard envelope and aborts the request. debug.error
// surfaces the cause only in development (gin debug mode).
func Render(c *gin.Context, err error) {
	ae := FromError(err)
	if ae == nil {
		return
	}

	send := c.MustGet("send").(func(r *types.Response))
	send(&types.Response{
		HTTPStatus: ae.HTTPStatus,
		Code:       ae.Code,
		Message:    ae.Message,
		Data:       ae.Data,
		Err:        ae.cause,
	})
	return
}

func Send(c *gin.Context, resp *types.Response) {
	send := c.MustGet("send").(func(r *types.Response))
	send(&types.Response{
		HTTPStatus: resp.HTTPStatus,
		Code:       resp.Code,
		Message:    resp.Message,
		Data:       resp.Data,
		Err:        resp.Err,
	})
	return
}

func Validate(c *gin.Context, request any) error {
	if err := validation.Validate(request); err != nil {
		Render(c, NewBadRequestException("Failed to validate payload").WithCause(err))
		return err
	}
	return nil
}
