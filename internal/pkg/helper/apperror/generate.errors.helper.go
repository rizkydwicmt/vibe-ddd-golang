package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

type AppException interface {
	error
	Code() int
	Message() string
	Unwrap() error
}

type baseException struct {
	message string
	code    int
	cause   error
}

func (e *baseException) Error() string {
	if e.cause != nil {
		return e.cause.Error()
	}
	return e.message
}

func (e *baseException) Code() int {
	return e.code
}

func (e *baseException) Message() string {
	return e.message
}

func (e *baseException) Unwrap() error {
	if e.cause == nil {
		return fmt.Errorf("%s", e.message)
	}
	return e.cause
}

type ValidationException struct {
	*baseException
}

func NewValidationError(cause error, message string) *ValidationException {
	if message == "" {
		message = "Failed to validate payload"
	}
	return &ValidationException{
		baseException: &baseException{
			message: message,
			code:    http.StatusBadRequest,
			cause:   cause,
		},
	}
}

type BadRequestException struct {
	*baseException
}

func NewBadRequestException(cause error, message string) *BadRequestException {
	return &BadRequestException{
		baseException: &baseException{
			message: message,
			code:    http.StatusBadRequest,
			cause:   cause,
		},
	}
}

type NotFoundException struct {
	*baseException
}

func NewNotFoundException(cause error, message string) *NotFoundException {
	return &NotFoundException{
		baseException: &baseException{
			message: message,
			code:    http.StatusNotFound,
			cause:   cause,
		},
	}
}

type NoRecordsException struct {
	*baseException
}

func NewNoRecordsException(cause error, message string) *NoRecordsException {
	return &NoRecordsException{
		baseException: &baseException{
			message: message,
			code:    http.StatusOK,
			cause:   cause,
		},
	}
}

type AuthorizationException struct {
	*baseException
}

func NewAuthorizationException(cause error, message string) *AuthorizationException {
	return &AuthorizationException{
		baseException: &baseException{
			message: message,
			code:    http.StatusUnauthorized,
			cause:   cause,
		},
	}
}

type ForbiddenException struct {
	*baseException
}

func NewForbiddenException(cause error, message string) *ForbiddenException {
	return &ForbiddenException{
		baseException: &baseException{
			message: message,
			code:    http.StatusForbidden,
			cause:   cause,
		},
	}
}

type InternalServerException struct {
	*baseException
}

func NewInternalServerException(cause error, message string) *InternalServerException {
	return &InternalServerException{
		baseException: &baseException{
			message: message,
			code:    http.StatusInternalServerError,
			cause:   cause,
		},
	}
}

func AsAppException(err error, appException *AppException) bool {
	return errors.As(err, appException)
}
