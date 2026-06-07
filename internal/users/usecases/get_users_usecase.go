package usecases

import (
	"context"
	"fmt"

	"github.com/yovannylopez/docsy-main/internal/users/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/users/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/users/domain/ports"
)

// GetUsersRequest represents the request to retrieve users
type GetUsersRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// GetUsersUseCase holds the repositories needed for the user retrieval use case
type GetUsersUseCase struct {
	userRepo ports.UserProfileRepository
}

// NewGetUsersUseCase creates a new instance of GetUsersUseCase
func NewGetUsersUseCase(userRepo ports.UserProfileRepository) *GetUsersUseCase {
	return &GetUsersUseCase{
		userRepo: userRepo,
	}
}

// Execute runs the user retrieval use case
func (uc *GetUsersUseCase) Execute(ctx context.Context, request *GetUsersRequest) (*dtos.UsersListResponse, error) {
	// Retrieve users from the repository
	users, err := uc.userRepo.GetAllUsers(ctx, request.Limit, request.Offset)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get users from repository (limit: %d, offset: %d): %w",
			request.Limit,
			request.Offset,
			err,
		)
	}

	// Get the total user count for pagination
	total, err := uc.userRepo.GetTotalUsersCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users count: %w", err)
	}

	// Convert entities to response DTOs
	userResponses := uc.convertToUserResponses(users)

	// Build final response
	response := &dtos.UsersListResponse{
		Usuarios: userResponses,
		Total:    total, // Use the actual total from the database
		Limite:   request.Limit,
		Offset:   request.Offset,
	}

	return response, nil
}

// convertToUserResponses converts user entities to response DTOs
func (uc *GetUsersUseCase) convertToUserResponses(users []entities.User) []dtos.UserListResponse {
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
