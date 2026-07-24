package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/handler"
	types "vibe-ddd-golang/internal/common/type"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setup(svc *testutil.MockUserService) *gin.Engine {
	r := testutil.NewTestRouter()
	h := handler.NewUserHandler(svc, testutil.NewSilentLogger())
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func do(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) types.ResponseAPI {
	t.Helper()
	var env types.ResponseAPI
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	return env
}

func TestCreateUser_Created(t *testing.T) {
	svc := new(testutil.MockUserService)
	svc.On("CreateUser", mock.Anything, mock.AnythingOfType("*dto.CreateUserRequest")).
		Return(&dto.UserResponse{ID: 1, Email: "john@example.com"}, nil)

	w := do(setup(svc), http.MethodPost, "/api/v1/users",
		`{"name":"John","email":"john@example.com","password":"password123"}`)

	assert.Equal(t, http.StatusCreated, w.Code)
	env := decode(t, w)
	assert.Equal(t, "CREATED", string(env.Code))
	assert.NotEmpty(t, env.RequestID)
	assert.Equal(t, "req_", env.RequestID[:4])
}

func TestCreateUser_InvalidPayload(t *testing.T) {
	svc := new(testutil.MockUserService) // service never called
	w := do(setup(svc), http.MethodPost, "/api/v1/users", `{"email":"not-an-email"}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "INVALID_PAYLOAD", string(decode(t, w).Code))
}

func TestGetUser_BadID(t *testing.T) {
	svc := new(testutil.MockUserService)
	w := do(setup(svc), http.MethodGet, "/api/v1/users/abc", "")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "INVALID_PAYLOAD", string(decode(t, w).Code))
}
