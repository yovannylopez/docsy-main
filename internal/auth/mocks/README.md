# Mocks del Módulo Auth

Este directorio contiene los mocks específicos del módulo de autenticación y autorización.

## Estructura

```
internal/auth/mocks/
├── user_repository_mock.go         # Mock para UserRepository del módulo auth
├── authentication_service_mock.go   # Mock para AuthenticationService
├── authorization_service_mock.go    # Mock para AuthorizationService
├── login_service_mock.go           # Mock para LoginService
├── signup_service_mock.go          # Mock para SignupService
├── password_hasher_mock.go         # Mock para PasswordHasher
├── token_generator_mock.go         # Mock para TokenGenerator
├── session_repository_mock.go      # Mock para SessionRepository
├── audit_repository_mock.go        # Mock para AuditRepository
└── system_config_repository_mock.go # Mock para SystemConfigRepository
```

## Generación de mocks

Los mocks se generan automáticamente usando `mockery`. Para generar los mocks del módulo auth:

```bash
# Generar mocks del módulo auth
mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=user_repository_mock.go --name=UserRepository
mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=authentication_service_mock.go --name=AuthenticationService
mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=authorization_service_mock.go --name=AuthorizationService
mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=login_service_mock.go --name=LoginService
mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=signup_service_mock.go --name=SignupService
mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=password_hasher_mock.go --name=PasswordHasher
mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=token_generator_mock.go --name=TokenGenerator
mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=session_repository_mock.go --name=SessionRepository
mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=audit_repository_mock.go --name=AuditRepository
mockery --dir=internal/auth/domain/ports --output=internal/auth/mocks --outpkg=mocks --filename=system_config_repository_mock.go --name=SystemConfigRepository
```

## Uso en pruebas

```go
package usecases_test

import (
    "testing"
    "context"
    
    "github.com/yovannylopez/docsy-main/internal/auth/mocks"
    "github.com/yovannylopez/docsy-main/internal/auth/usecases"
    "github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
)

func TestSignupUseCase(t *testing.T) {
    mockUserRepo := mocks.NewUserRepository(t)
    mockPasswordHasher := mocks.NewPasswordHasher(t)
    mockTokenGenerator := mocks.NewTokenGenerator(t)

    // Configurar expectativas
    mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, nil)
    mockPasswordHasher.On("Hash", "password123").Return("hashed_password", nil)

    // Crear caso de uso con mocks
    useCase := usecases.NewSignupUseCase(mockUserRepo, mockPasswordHasher, mockTokenGenerator)

    // Ejecutar prueba...
    request := &dtos.SignupRequest{
        Email:    "test@example.com",
        Password: "password123",
        Name:     "Test User",
    }
    
    result, err := useCase.Execute(context.Background(), request)
    // Assertions...
}
```

## Notas importantes

- Los mocks están dentro del módulo auth para mantener la independencia
- Se regeneran automáticamente cuando cambias las interfaces del módulo
- No edites manualmente los archivos de mock, se sobrescribirán
- Cada módulo tiene sus propios mocks específicos 