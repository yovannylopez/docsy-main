package dtos

import "time"

// UserListResponse represents the response for a user in the GET /api/v1/users endpoint
type UserListResponse struct {
	ID                   string     `json:"id"`
	NombreUsuario        *string    `json:"nombre_usuario"`
	Email                string     `json:"email"`
	PrimerNombre         string     `json:"primer_nombre"`
	SegundoNombre        string     `json:"segundo_nombre"`
	NumeroIdentificacion *string    `json:"numero_identificacion"`
	TipoIdentificacion   *string    `json:"tipo_identificacion"`
	Telefono             *string    `json:"telefono"`
	EstaActivo           bool       `json:"esta_activo"`
	EstaVerificado       bool       `json:"esta_verificado"`
	UltimoAcceso         *time.Time `json:"ultimo_acceso"`
	IntentosFallidos     int        `json:"intentos_fallidos"`
	UltimoIntentoFallido *time.Time `json:"ultimo_intento_fallido"`
	MfaHabilitado        bool       `json:"mfa_habilitado"`
	FechaCreacion        time.Time  `json:"fecha_creacion"`
	CreadoPor            *string    `json:"creado_por"`
	ActualizadoPor       *string    `json:"actualizado_por"`
}

// UsersListResponse represents the response for the GET /api/v1/users endpoint
type UsersListResponse struct {
	Usuarios []UserListResponse `json:"usuarios"`
	Total    int                `json:"total"`
	Limite   int                `json:"limite"`
	Offset   int                `json:"offset"`
}

// UserDetailResponse represents the detailed response for a user
type UserDetailResponse struct {
	ID                  string         `json:"id"`
	Email               string         `json:"email"`
	FirstName           string         `json:"first_name"`
	LastName            string         `json:"last_name"`
	Phone               *string        `json:"phone,omitempty"`
	IsActive            bool           `json:"is_active"`
	IsVerified          bool           `json:"is_verified"`
	LastLoginAt         *time.Time     `json:"last_login_at,omitempty"`
	FailedLoginAttempts int            `json:"failed_login_attempts"`
	LastFailedLoginAt   *time.Time     `json:"last_failed_login_at,omitempty"`
	LockedUntil         *time.Time     `json:"locked_until,omitempty"`
	MFAEnabled          bool           `json:"mfa_enabled"`
	PasswordChangedAt   time.Time      `json:"password_changed_at"`
	MustChangePassword  bool           `json:"must_change_password"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	Roles               []RoleResponse `json:"roles"`
}

// RoleResponse represents the role response
type RoleResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	IsSystemRole bool      `json:"is_system_role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateUserRequest represents the request to create an individual user
type CreateUserRequest struct {
	Username             *string `json:"username,omitempty"`
	Email                string  `json:"email" validate:"required,email"`
	Password             string  `json:"password" validate:"required,min=8"`
	FirstName            string  `json:"first_name" validate:"required,min=2"`
	LastName             string  `json:"last_name" validate:"required,min=2"`
	IdentificationNumber *string `json:"identification_number,omitempty"`
	IdentificationType   *string `json:"identification_type,omitempty"`
	Phone                *string `json:"phone,omitempty"`
	IsActive             *bool   `json:"is_active,omitempty"`
	IsVerified           *bool   `json:"is_verified,omitempty"`
	MFAEnabled           *bool   `json:"mfa_enabled,omitempty"`
	RoleName             string  `json:"role_name" validate:"required"`
}

// CreateUsersRequest represents the request to create multiple users
type CreateUsersRequest struct {
	Users []CreateUserRequest `json:"users" validate:"required,min=1"`
}

// CreateUserResponse represents the user creation response
type CreateUserResponse struct {
	ID                   string    `json:"id"`
	Username             *string   `json:"username,omitempty"`
	Email                string    `json:"email"`
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	IdentificationNumber *string   `json:"identification_number,omitempty"`
	IdentificationType   *string   `json:"identification_type,omitempty"`
	Phone                *string   `json:"phone,omitempty"`
	IsActive             bool      `json:"is_active"`
	IsVerified           bool      `json:"is_verified"`
	MFAEnabled           bool      `json:"mfa_enabled"`
	CreatedAt            time.Time `json:"created_at"`
	CreatedBy            *string   `json:"created_by,omitempty"`
	UpdatedBy            *string   `json:"updated_by,omitempty"`
}

// CreateUsersResponse represents the response for creating multiple users
type CreateUsersResponse struct {
	CreatedUsers []CreateUserResponse `json:"created_users"`
	TotalCreated int                  `json:"total_created"`
	Errors       []UserCreationError  `json:"errors,omitempty"`
}

// UserCreationError represents an error during the creation of a specific user
type UserCreationError struct {
	Index int    `json:"index"`
	Email string `json:"email"`
	Error string `json:"error"`
}

// UpdateUserRequest represents the user update request
type UpdateUserRequest struct {
	Email                *string `json:"email,omitempty" validate:"omitempty,email"`
	Username             *string `json:"username,omitempty"`
	FirstName            *string `json:"first_name,omitempty" validate:"omitempty,min=2"`
	LastName             *string `json:"last_name,omitempty" validate:"omitempty,min=2"`
	IdentificationNumber *string `json:"identification_number,omitempty"`
	IdentificationType   *string `json:"identification_type,omitempty"`
	Phone                *string `json:"phone,omitempty"`
	IsActive             *bool   `json:"is_active,omitempty"`
	IsVerified           *bool   `json:"is_verified,omitempty"`
	MFAEnabled           *bool   `json:"mfa_enabled,omitempty"`
}

// UpdateUserResponse represents the user update response
type UpdateUserResponse struct {
	ID                   string    `json:"id"`
	Email                string    `json:"email"`
	Username             *string   `json:"username,omitempty"`
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	IdentificationNumber *string   `json:"identification_number,omitempty"`
	IdentificationType   *string   `json:"identification_type,omitempty"`
	Phone                *string   `json:"phone,omitempty"`
	IsActive             bool      `json:"is_active"`
	IsVerified           bool      `json:"is_verified"`
	MFAEnabled           bool      `json:"mfa_enabled"`
	UpdatedAt            time.Time `json:"updated_at"`
	UpdatedBy            *string   `json:"updated_by,omitempty"`
}

// ResetPasswordRequest represents the password reset request
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}
