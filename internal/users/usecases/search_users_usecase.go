package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/yovannylopez/docsy-main/internal/users/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
	domainerrors "github.com/yovannylopez/docsy-main/internal/users/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/users/domain/ports"
	"github.com/yovannylopez/docsy-main/pkg/constants"
)

const (
	defaultSearchLimit = 10
	minSearchLimit     = 1
)

// SearchUsersUseCase holds the repositories needed for the user search use case
type SearchUsersUseCase struct {
	userRepo ports.UserProfileRepository
}

// NewSearchUsersUseCase creates a new instance of SearchUsersUseCase
func NewSearchUsersUseCase(userRepo ports.UserProfileRepository) *SearchUsersUseCase {
	return &SearchUsersUseCase{
		userRepo: userRepo,
	}
}

// Execute runs the user search use case
func (uc *SearchUsersUseCase) Execute(ctx context.Context, request *dtos.SearchUsersRequest) (*dtos.UsersListResponse, error) {
	// Validate query
	query := strings.TrimSpace(request.Q)
	if query == "" {
		return nil, domainerrors.ErrSearchQueryEmpty
	}

	// Validate and adjust limit
	limit := request.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	if limit > constants.MaxUsersLimit {
		limit = constants.MaxUsersLimit
	}
	if limit < minSearchLimit {
		limit = minSearchLimit
	}

	// Validate and adjust offset
	offset := request.Offset
	if offset < 0 {
		offset = 0
	}

	// Search users from the repository
	users, err := uc.userRepo.SearchUsers(ctx, query, request.Activo, limit, offset)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to search users from repository (query: %s, limit: %d, offset: %d): %w",
			query,
			limit,
			offset,
			err,
		)
	}

	// Get the total users matching the search
	total, err := uc.userRepo.CountSearchUsers(ctx, query, request.Activo)
	if err != nil {
		return nil, fmt.Errorf("failed to count search users: %w", err)
	}

	// Convert entities to response DTOs
	userResponses := uc.convertToUserResponses(users)

	// Build final response
	response := &dtos.UsersListResponse{
		Usuarios: userResponses,
		Total:    total,
		Limite:   limit,
		Offset:   offset,
	}

	return response, nil
}

// convertToUserResponses converts user entities to response DTOs
func (uc *SearchUsersUseCase) convertToUserResponses(users []entities.User) []dtos.UserListResponse {
	userResponses := make([]dtos.UserListResponse, 0, len(users))

	for _, user := range users {
		userResponse := dtos.UserListResponse{
			ID:                   user.ID,
			NombreUsuario:        user.Username,
			Email:                user.Email,
			PrimerNombre:         user.FirstName,
			SegundoNombre:        user.LastName,
			NumeroIdentificacion: user.IdentificationNumber,
			TipoIdentificacion:   user.IdentificationType,
			Telefono:             user.Phone,
			EstaActivo:           user.IsActive,
			EstaVerificado:       user.IsVerified,
			UltimoAcceso:         user.LastLoginAt,
			IntentosFallidos:     user.FailedLoginAttempts,
			UltimoIntentoFallido: user.LastFailedLoginAt,
			MfaHabilitado:        user.MFAEnabled,
			FechaCreacion:        user.CreatedAt,
			CreadoPor:            user.CreatedBy,
			ActualizadoPor:       user.UpdatedBy,
		}
		userResponses = append(userResponses, userResponse)
	}

	return userResponses
}
