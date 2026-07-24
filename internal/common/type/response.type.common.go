// Package types holds cross-cutting transport DTOs shared by middleware and
// handlers. Domain DTOs live in their own application package, not here.
package types

import (
	"time"

	"vibe-ddd-golang/internal/common/enum"
)

// Response is the internal handler→middleware hand-off. A handler builds one and
// passes it to the `send` function installed by ResponseInit; the middleware
// renders the outward ResponseAPI.
type Response struct {
	HTTPStatus int
	Code       enum.ResultCode
	Message    string
	Data       any
	Err        error // logged + sanitized into debug.error; never serialized raw
}

// ResponseAPIDebug is runtime metadata. version/timing are always safe to return;
// `error` MUST be sanitized in production (no stack traces, SQL, secrets, PII).
type ResponseAPIDebug struct {
	Version   string    `json:"version"`
	Error     *string   `json:"error"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	RuntimeMs int64     `json:"runtimeMs"`
}

// ResponseAPI is the standard plaintext API response. Clients branch on `Code`, display
// `Message`, and read business payload from `Data`.
//
//	{
//	  "requestId": "...",
//	  "code": "OK",
//	  "message": "...",
//	  "debug": { "version": "v1.0.0", "error": null, "startTime": ..., "endTime": ..., "runtimeMs": 12 },
//	  "data": {}
//	}
type ResponseAPI struct {
	RequestID string            `json:"requestId"`
	Code      enum.ResultCode   `json:"code"`
	Message   string            `json:"message"`
	Debug     *ResponseAPIDebug `json:"debug,omitempty"`
	Data      any               `json:"data"`
}
