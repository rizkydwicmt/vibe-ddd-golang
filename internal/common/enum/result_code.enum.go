package enum

import "net/http"

// ResultCode is the stable, machine-readable result code clients branch on
type ResultCode string

const (
	CodeOK                    ResultCode = "OK"
	CodeCreated               ResultCode = "CREATED"
	CodeAccepted              ResultCode = "ACCEPTED"
	CodeInvalidPayload        ResultCode = "INVALID_PAYLOAD"
	CodeUnsupportedAPIVersion ResultCode = "UNSUPPORTED_API_VERSION"
	CodeBadRequest            ResultCode = "BAD_REQUEST"
	CodeUnauthorized          ResultCode = "UNAUTHORIZED"
	CodeForbidden             ResultCode = "FORBIDDEN"
	CodeNotFound              ResultCode = "NOT_FOUND"
	CodeConflict              ResultCode = "CONFLICT"
	CodeIdempotencyConflict   ResultCode = "IDEMPOTENCY_CONFLICT"
	CodeRateLimited           ResultCode = "RATE_LIMITED"
	CodeUnprocessableEntity   ResultCode = "UNPROCESSABLE_ENTITY"
	CodeProviderUnavailable   ResultCode = "PROVIDER_UNAVAILABLE"
	CodeServiceUnavailable    ResultCode = "SERVICE_UNAVAILABLE"
	CodeTimeout               ResultCode = "TIMEOUT"
	CodeInternalError         ResultCode = "INTERNAL_ERROR"
)

// String returns the wire form.
func (c ResultCode) ToString() string { return string(c) }

// DefaultHTTPStatus maps a result code to its conventional HTTP status
func (c ResultCode) DefaultHTTPStatus() int {
	switch c {
	case CodeOK:
		return http.StatusOK
	case CodeCreated:
		return http.StatusCreated
	case CodeAccepted:
		return http.StatusAccepted
	case CodeInvalidPayload, CodeUnsupportedAPIVersion, CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict, CodeIdempotencyConflict:
		return http.StatusConflict
	case CodeUnprocessableEntity:
		return http.StatusUnprocessableEntity
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeProviderUnavailable:
		return http.StatusBadGateway
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case CodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

func DefaultResultCode(code int) ResultCode {
	switch code {
	case http.StatusOK:
		return CodeOK
	case http.StatusCreated:
		return CodeCreated
	case http.StatusAccepted:
		return CodeAccepted
	case http.StatusBadRequest:
		return CodeBadRequest
	case http.StatusUnauthorized:
		return CodeUnauthorized
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusConflict:
		return CodeConflict
	case http.StatusTooManyRequests:
		return CodeRateLimited
	case http.StatusBadGateway:
		return CodeProviderUnavailable
	case http.StatusServiceUnavailable:
		return CodeServiceUnavailable
	case http.StatusGatewayTimeout:
		return CodeTimeout
	default:
		return CodeInternalError
	}
}
