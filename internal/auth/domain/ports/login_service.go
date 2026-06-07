package ports

import (
	"context"

	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
)

// LoginService define la interfaz para el servicio de login
type LoginService interface {
	Login(ctx context.Context, request *dtos.LoginRequest, userAgent, ipAddress string) (*dtos.LoginResponse, error)
}
