package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/handler"
	"vibe-ddd-golang/internal/application/user/repository"
	"vibe-ddd-golang/internal/application/user/service"
	"vibe-ddd-golang/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserIntegration(t *testing.T) (*gin.Engine, func()) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)

	logger := testutil.NewTestLogger(t)

	// Create real instances (no mocks)
	userRepo := repository.NewUserRepository(db, logger)
	userService := service.NewUserService(userRepo, logger)
	userHandler := handler.NewUserHandler(userService, logger)

	// Setup Gin router
	router := gin.New()
	api := router.Group("/api/v1")
	userHandler.RegisterRoutes(api)

	cleanup := func() {
		testutil.CleanDB(db)
	}

	return router, cleanup
}

func TestUserIntegration_CreateAndGetUser(t *testing.T) {
	router, cleanup := setupUserIntegration(t)
	defer cleanup()

	// Test data
	createReq := &dto.CreateUserRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	// Step 1: Create user
	reqBody, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	userID := int(data["id"].(float64))
	assert.Equal(t, createReq.Name, data["name"])
	assert.Equal(t, createReq.Email, data["email"])

	// Step 2: Get the created user
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/api/v1/users/"+string(rune(userID+'0')), nil)

	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var getResp map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &getResp)
	require.NoError(t, err)

	userData := getResp["data"].(map[string]interface{})
	assert.Equal(t, float64(userID), userData["id"])
	assert.Equal(t, createReq.Name, userData["name"])
	assert.Equal(t, createReq.Email, userData["email"])
}

func TestUserIntegration_CreateDuplicateEmail(t *testing.T) {
	router, cleanup := setupUserIntegration(t)
	defer cleanup()

	// Test data
	createReq := &dto.CreateUserRequest{
		Name:     "John Doe",
		Email:    "duplicate@example.com",
		Password: "password123",
	}

	// Step 1: Create first user
	reqBody, _ := json.Marshal(createReq)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBody))
	req1.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	// Step 2: Try to create user with same email
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBody))
	req2.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusConflict, w2.Code)

	var errorResp map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "email already exists")
}

func TestUserIntegration_GetUsers(t *testing.T) {
	router, cleanup := setupUserIntegration(t)
	defer cleanup()

	// Create multiple users
	users := []dto.CreateUserRequest{
		{Name: "User 1", Email: "user1@example.com", Password: "password1"},
		{Name: "User 2", Email: "user2@example.com", Password: "password2"},
		{Name: "User 3", Email: "user3@example.com", Password: "password3"},
	}

	// Create users
	for _, user := range users {
		reqBody, _ := json.Marshal(user)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// Get all users
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/users?page=1&page_size=10", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.UserListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Data, 3)
	assert.Equal(t, int64(3), response.TotalCount)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PageSize)
}

func TestUserIntegration_UpdateUser(t *testing.T) {
	router, cleanup := setupUserIntegration(t)
	defer cleanup()

	// Create user
	createReq := &dto.CreateUserRequest{
		Name:     "Original Name",
		Email:    "original@example.com",
		Password: "password123",
	}

	reqBody, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	userID := int(data["id"].(float64))

	// Update user
	updateReq := &dto.UpdateUserRequest{
		Name:  "Updated Name",
		Email: "updated@example.com",
	}

	updateBody, _ := json.Marshal(updateReq)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("PUT", "/api/v1/users/"+string(rune(userID+'0')), bytes.NewBuffer(updateBody))
	req2.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var updateResp map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &updateResp)
	require.NoError(t, err)

	updatedData := updateResp["data"].(map[string]interface{})
	assert.Equal(t, updateReq.Name, updatedData["name"])
	assert.Equal(t, updateReq.Email, updatedData["email"])
}

func TestUserIntegration_DeleteUser(t *testing.T) {
	router, cleanup := setupUserIntegration(t)
	defer cleanup()

	// Create user
	createReq := &dto.CreateUserRequest{
		Name:     "To Be Deleted",
		Email:    "delete@example.com",
		Password: "password123",
	}

	reqBody, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)

	data := createResp["data"].(map[string]interface{})
	userID := int(data["id"].(float64))

	// Delete user
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("DELETE", "/api/v1/users/"+string(rune(userID+'0')), nil)

	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var deleteResp map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &deleteResp)
	require.NoError(t, err)
	assert.Equal(t, "User deleted successfully", deleteResp["message"])

	// Try to get deleted user (should return 404)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/api/v1/users/"+string(rune(userID+'0')), nil)

	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusNotFound, w3.Code)
}
