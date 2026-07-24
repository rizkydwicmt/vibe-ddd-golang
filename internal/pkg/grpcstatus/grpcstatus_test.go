package grpcstatus_test

import (
	"errors"
	"testing"

	"vibe-ddd-golang/internal/pkg/grpcstatus"
	"vibe-ddd-golang/internal/pkg/response"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFromError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "not found", err: response.NewNotFoundException("missing"), code: codes.NotFound},
		{name: "bad request", err: response.NewBadRequestException("invalid"), code: codes.InvalidArgument},
		{name: "conflict", err: response.NewConflictException("exists"), code: codes.AlreadyExists},
		{name: "unauthorized", err: response.NewUnauthorizedException("login"), code: codes.Unauthenticated},
		{name: "forbidden", err: response.NewForbiddenException("denied"), code: codes.PermissionDenied},
		{name: "unknown", err: errors.New("boom"), code: codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, status.Code(grpcstatus.FromError(tt.err)))
		})
	}
}
