package handler

import (
	"net/http"
	"strconv"

	"vibe-ddd-golang/internal/application/user/dto"
	"vibe-ddd-golang/internal/application/user/service"
	"vibe-ddd-golang/internal/common/enum"
	types "vibe-ddd-golang/internal/common/type"
	"vibe-ddd-golang/internal/pkg/reqbind"
	"vibe-ddd-golang/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	service service.UserService
	logger  *zap.Logger
}

func NewUserHandler(service service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Render(c, response.NewBadRequestException("invalid id"))
		return 0, false
	}
	return uint(id), true
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User creation request"
// @Success 201 {object} types.ResponseAPI "Created user"
// @Failure 400 {object} types.ResponseAPI "Invalid request body"
// @Failure 409 {object} types.ResponseAPI "Email already exists"
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := reqbind.Bind(c, &req); err != nil {
		response.Render(c, response.NewBadRequestException("Invalid request body").WithCause(err))
		return
	}
	if err := response.Validate(c, &req); err != nil {
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusCreated, Code: enum.CodeCreated, Message: "User created.", Data: user})
}

// GetUser godoc
// @Summary Get a user by ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} types.ResponseAPI "User details"
// @Failure 404 {object} types.ResponseAPI "User not found"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusOK, Code: enum.CodeOK, Message: "OK", Data: user})
}

// GetUsers godoc
// @Summary Get all users
// @Tags users
// @Produce json
// @Param name query string false "Filter by name"
// @Param email query string false "Filter by email"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {object} types.ResponseAPI "List of users"
// @Router /users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	var filter dto.UserFilter
	if err := reqbind.BindQuery(c, &filter); err != nil {
		response.Render(c, response.NewBadRequestException("Invalid query parameters").WithCause(err))
		return
	}

	users, err := h.service.GetUsers(c.Request.Context(), &filter)
	if err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusOK, Code: enum.CodeOK, Message: "OK", Data: users})
}

// UpdateUser godoc
// @Summary Update a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserRequest true "User update request"
// @Success 200 {object} types.ResponseAPI "Updated user"
// @Failure 404 {object} types.ResponseAPI "User not found"
// @Failure 409 {object} types.ResponseAPI "Email already exists"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.UpdateUserRequest
	if err := reqbind.Bind(c, &req); err != nil {
		response.Render(c, response.NewBadRequestException("Invalid request body").WithCause(err))
		return
	}
	if err := response.Validate(c, &req); err != nil {
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusOK, Code: enum.CodeOK, Message: "User updated.", Data: user})
}

// UpdateUserPassword godoc
// @Summary Update user password
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param password body dto.UpdateUserPasswordRequest true "Password update request"
// @Success 200 {object} types.ResponseAPI "Password updated"
// @Failure 401 {object} types.ResponseAPI "Current password is incorrect"
// @Failure 404 {object} types.ResponseAPI "User not found"
// @Router /users/{id}/password [put]
func (h *UserHandler) UpdateUserPassword(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.UpdateUserPasswordRequest
	if err := reqbind.Bind(c, &req); err != nil {
		response.Render(c, response.NewBadRequestException("Invalid request body").WithCause(err))
		return
	}
	if err := response.Validate(c, &req); err != nil {
		return
	}

	if err := h.service.UpdateUserPassword(c.Request.Context(), id, &req); err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusOK, Code: enum.CodeOK, Message: "Password updated.", Data: nil})
}

// DeleteUser godoc
// @Summary Delete a user
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} types.ResponseAPI "User deleted"
// @Failure 404 {object} types.ResponseAPI "User not found"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
		response.Render(c, err)
		return
	}

	response.Send(c, &types.Response{HTTPStatus: http.StatusOK, Code: enum.CodeOK, Message: "User deleted.", Data: nil})
}

func (h *UserHandler) RegisterRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")
	{
		users.POST("", h.CreateUser)
		users.GET("", h.GetUsers)
		users.GET("/:id", h.GetUser)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
		users.PUT("/:id/password", h.UpdateUserPassword)
	}
}
