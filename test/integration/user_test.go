package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"vibe-ddd-golang/internal/application/user/handler"
	"vibe-ddd-golang/internal/application/user/repository"
	"vibe-ddd-golang/internal/application/user/service"
	"vibe-ddd-golang/internal/common/params"
	types "vibe-ddd-golang/internal/common/type"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupUserIntegration wires real repo→service→handler over an in-memory SQLite DB,
// behind the standard request/response envelope middleware.
func setupUserIntegration(t *testing.T) *gin.Engine {
	db, err := testutil.SetupTestDatabase()
	require.NoError(t, err)

	log := testutil.NewSilentLogger()
	userRepo := repository.NewUserRepository(params.Params{MainDB: db}, log)
	userService := service.NewUserService(userRepo, log)
	userHandler := handler.NewUserHandler(userService, log)

	r := testutil.NewTestRouter()
	userHandler.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func request(t *testing.T, r *gin.Engine, method, path, body string) (*httptest.ResponseRecorder, types.ResponseAPI) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var env types.ResponseAPI
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	return w, env
}

func TestUserIntegration_CreateAndGetUser(t *testing.T) {
	r := setupUserIntegration(t)

	w, env := request(t, r, http.MethodPost, "/api/v1/users",
		`{"name":"John Doe","email":"john@example.com","password":"password123"}`)
	require.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "CREATED", string(env.Code))
	assert.NotEmpty(t, env.RequestID)

	data := env.Data.(map[string]interface{})
	userID := int(data["id"].(float64))
	assert.Equal(t, "john@example.com", data["email"])

	_, env2 := request(t, r, http.MethodGet, "/api/v1/users/"+strconv.Itoa(userID), "")
	assert.Equal(t, "OK", string(env2.Code))
	assert.Equal(t, "john@example.com", env2.Data.(map[string]interface{})["email"])
}

func TestUserIntegration_DuplicateEmailConflict(t *testing.T) {
	r := setupUserIntegration(t)
	body := `{"name":"A","email":"dup@example.com","password":"password123"}`

	w1, _ := request(t, r, http.MethodPost, "/api/v1/users", body)
	require.Equal(t, http.StatusCreated, w1.Code)

	w2, env := request(t, r, http.MethodPost, "/api/v1/users", body)
	assert.Equal(t, http.StatusConflict, w2.Code)
	assert.Equal(t, "CONFLICT", string(env.Code))
}

func TestUserIntegration_NotFound(t *testing.T) {
	r := setupUserIntegration(t)
	w, env := request(t, r, http.MethodGet, "/api/v1/users/999", "")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "NOT_FOUND", string(env.Code))
}
