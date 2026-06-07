package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/users/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/users/usecases"
	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
	"github.com/yovannylopez/docsy-main/pkg/responses"
)

// UsersHandler handles HTTP requests related to users
type UsersHandler struct {
	getUsersUseCase    *usecases.GetUsersUseCase
	createUsersUseCase *usecases.CreateUsersUseCase
	updateUserUseCase  *usecases.UpdateUserUseCase
	searchUsersUseCase *usecases.SearchUsersUseCase
	getUserByIDUseCase *usecases.GetUserByIDUseCase
}

// NewUsersHandler creates a new UsersHandler instance with injected use cases
func NewUsersHandler(
	getUsersUC *usecases.GetUsersUseCase,
	createUsersUC *usecases.CreateUsersUseCase,
	updateUserUC *usecases.UpdateUserUseCase,
	searchUsersUC *usecases.SearchUsersUseCase,
	getUserByIDUC *usecases.GetUserByIDUseCase,
) *UsersHandler {
	return &UsersHandler{
		getUsersUseCase:    getUsersUC,
		createUsersUseCase: createUsersUC,
		updateUserUseCase:  updateUserUC,
		searchUsersUseCase: searchUsersUC,
		getUserByIDUseCase: getUserByIDUC,
	}
}

// SearchUsers handles the user search request
func (h *UsersHandler) SearchUsers(c echo.Context) error {
	// Get search parameter
	query := strings.TrimSpace(c.QueryParam("q"))
	if query == "" {
		return responses.BadRequest(c, "the search parameter 'q' is required")
	}

	// Create pagination parser with custom configuration for users
	config := pagination.Config{
		DefaultLimit: 10, //nolint:mnd
		MaxLimit:     constants.MaxUsersLimit,
		MinLimit:     1,
	}
	parser := pagination.NewParser(config)

	// Parse pagination parameters
	params, err := parser.ParseFromQuery(
		c.QueryParam("limit"),
		c.QueryParam("offset"),
	)
	if err != nil {
		return responses.BadRequest(c, err.Error())
	}

	// Parse active parameter (optional)
	var activo *bool
	if activoParam := c.QueryParam("activo"); activoParam != "" {
		activoBool, err := strconv.ParseBool(activoParam)
		if err != nil {
			return responses.BadRequest(c, "the 'activo' parameter must be 'true' or 'false'")
		}
		activo = &activoBool
	}

	// Build request for the use case
	request := &dtos.SearchUsersRequest{
		Q:      query,
		Limit:  params.Limit,
		Offset: params.Offset,
		Activo: activo,
	}

	// Execute use case
	response, err := h.searchUsersUseCase.Execute(c.Request().Context(), request)
	if err != nil {
		return err
	}

	paginatedResponse := pagination.CreateResponse(response.Usuarios, params, response.Total)
	return responses.OKPaginated(c, "user search completed successfully", paginatedResponse)
}

// GetUsers handles the request to retrieve the list of users
func (h *UsersHandler) GetUsers(c echo.Context) error {
	// Create pagination parser with custom configuration for users
	config := pagination.Config{
		DefaultLimit: 10, //nolint:mnd
		MaxLimit:     constants.MaxUsersLimit,
		MinLimit:     1,
	}
	parser := pagination.NewParser(config)

	// Parse pagination parameters
	params, err := parser.ParseFromQuery(
		c.QueryParam("limit"),
		c.QueryParam("offset"),
	)
	if err != nil {
		return responses.BadRequest(c, err.Error())
	}

	// Build request for the use case
	request := &usecases.GetUsersRequest{
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	// Execute use case
	response, err := h.getUsersUseCase.Execute(c.Request().Context(), request)
	if err != nil {
		return err
	}

	paginatedResponse := pagination.CreateResponse(response.Usuarios, params, response.Total)
	return responses.OKPaginated(c, "users retrieved successfully", paginatedResponse)
}

// GetUserByID handles the request to retrieve a specific user
func (h *UsersHandler) GetUserByID(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return responses.BadRequest(c, "User ID is required")
	}

	// Retrieve user via use case
	user, err := h.getUserByIDUseCase.Execute(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, usecases.ErrUserNotFound) {
			return responses.NotFound(c, "User not found")
		}

		return err
	}

	return responses.OK(c, user, "User retrieved successfully")
}

// CreateUser handles the request to create users (individual or multiple)
func (h *UsersHandler) CreateUser(c echo.Context) error {
	// Get userID from X-User-ID header
	createdByUserID := c.Request().Header.Get("X-User-ID")

	// Read the request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return responses.BadRequest(c, "error reading request body")
	}

	// Try to parse as array of users first
	var request dtos.CreateUsersRequest
	err = json.Unmarshal(body, &request)
	if err != nil || len(request.Users) == 0 {
		// If it fails or has no users, try to parse as a single user
		var singleUser dtos.CreateUserRequest
		if err := json.Unmarshal(body, &singleUser); err != nil {
			return responses.BadRequest(c, "Invalid data format. Must be a single user or a users array.")
		}

		request = dtos.CreateUsersRequest{
			Users: []dtos.CreateUserRequest{singleUser},
		}
	}

	if len(request.Users) == 0 {
		return responses.BadRequest(c, "at least one user must be provided for creation")
	}

	// Execute use case with the creating user's ID
	response, err := h.createUsersUseCase.Execute(c.Request().Context(), &request, createdByUserID)
	if err != nil {
		return err
	}

	// Determine the response message
	var message string
	if len(request.Users) == 1 {
		if response.TotalCreated == 1 {
			message = "User created successfully"
		} else {
			message = "Error creating user"
		}
	} else {
		switch {
		case response.TotalCreated == len(request.Users):
			message = fmt.Sprintf("%d users created successfully", response.TotalCreated)
		case response.TotalCreated > 0:
			message = fmt.Sprintf("%d of %d users created successfully", response.TotalCreated, len(request.Users))
		default:
			message = "Error creating users"
		}
	}

	// Return successful response with 201 Created status code
	return responses.Created(c, response, message)
}

// PatchUser handles the partial update request for a user
func (h *UsersHandler) PatchUser(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return responses.BadRequest(c, "User ID is required")
	}

	// Get updatedBy from X-User-ID header
	updatedByUserID := c.Request().Header.Get("X-User-ID")

	// Parse request body
	var req dtos.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return responses.BadRequest(c, fmt.Sprintf("Invalid request body: %v", err))
	}

	// Execute use case
	response, err := h.updateUserUseCase.Execute(c.Request().Context(), userID, &req, updatedByUserID)
	if err != nil {
		return err
	}

	return responses.OK(c, response, "User updated successfully")
}

// UpdateUser handles the user update request (kept for compatibility)
func (h *UsersHandler) UpdateUser(c echo.Context) error {
	return h.PatchUser(c)
}

// GetUserProfile handles the request to retrieve the authenticated user's profile
func (h *UsersHandler) GetUserProfile(c echo.Context) error {
	// Get userID from X-User-ID header
	userID := c.Request().Header.Get("X-User-ID")
	if userID == "" {
		return responses.BadRequest(c, "X-User-ID header is required")
	}

	// Retrieve user via use case
	user, err := h.getUserByIDUseCase.Execute(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, usecases.ErrUserNotFound) {
			return responses.NotFound(c, "User not found")
		}

		return err
	}

	return responses.OK(c, user, "User profile retrieved successfully")
}

// UpdateUserProfile handles the request to update the authenticated user's profile
func (h *UsersHandler) UpdateUserProfile(c echo.Context) error {
	// TODO: Implement profile update
	return responses.NotImplemented(c, "Update user profile not implemented yet")
}

// ResetPassword handles the password reset request
func (h *UsersHandler) ResetPassword(c echo.Context) error {
	// TODO: Implement password reset
	return responses.NotImplemented(c, "Reset password not implemented yet")
}
