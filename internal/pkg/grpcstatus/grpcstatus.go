package grpcstatus

import (
	"vibe-ddd-golang/internal/common/enum"
	"vibe-ddd-golang/internal/pkg/response"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FromError(err error) error {
	if err == nil {
		return nil
	}

	appErr, _ := response.As(err)
	code := codes.Internal
	switch appErr.Code {
	case enum.CodeNotFound:
		code = codes.NotFound
	case enum.CodeInvalidPayload, enum.CodeBadRequest:
		code = codes.InvalidArgument
	case enum.CodeConflict:
		code = codes.AlreadyExists
	case enum.CodeUnauthorized:
		code = codes.Unauthenticated
	case enum.CodeForbidden:
		code = codes.PermissionDenied
	}

	return status.Error(code, appErr.Message)
}
