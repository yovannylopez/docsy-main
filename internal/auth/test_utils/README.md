# Auth Test Utils (Object Mother)

Este paquete implementa el patrón **Object Mother** para el bounded context `auth`. Es la **fuente obligatoria** de entidades, DTOs y escenarios en tests bajo `internal/auth/` (ver `.cursor/skills/go-testing/SKILL.md` y `docs/ARCHITECTURE.md`). Usa `NewAuthStubs()`, `CloneUser`, `CloneLoginRequest`, etc., para variar solo el delta del caso.

Incluye entidades, DTOs, casos de uso y escenarios de prueba predefinidos.

### Clones (`clone.go`)

Para no mutar entidades compartidas del mother (p. ej. UUIDs generados en `NewAuthStubs()`), usa **`CloneUser`**, **`CloneRole`**, **`CloneSession`**, **`CloneAuthToken`**, **`CloneLoginRequest`**, **`CloneSignupRequest`**, **`CloneAuditLog`** antes de cambiar campos en un test. `GetTestUsers` / `GetTestRoles` ya devuelven copias independientes.

**`StringPtr(s string) *string`** (`ptr.go`) es el helper compartido para punteros a string en tests; otros bounded contexts (p. ej. `users`) lo reutilizan vía `authtest.StringPtr`.

## Estructura

### AuthTestEntities
Contiene entidades de prueba para todos los tipos del módulo auth:

- **Users**: `ValidUser`, `AdminUser`, `InactiveUser`, `LockedUser`, `UnverifiedUser`, `UserWithMFA`, `UserWithPhone`, `EmptyUser`, `InvalidUser`
- **Roles**: `ValidRole`, `AdminRole`, `UserRole`, `GuestRole`, `EmptyRole`, `InvalidRole`
- **Sessions**: `ValidSession`, `ExpiredSession`, `RevokedSession`, `EmptySession`, `InvalidSession`
- **AuditLogs**: `ValidAuditLog`, `FailedAuditLog`, `EmptyAuditLog`, `InvalidAuditLog`
- **AuthTokens**: `ValidAuthToken`, `ExpiredAuthToken`, `EmptyAuthToken`, `InvalidAuthToken`

### AuthTestDTOs
Contiene DTOs de prueba para requests y responses:

- **Login**: `ValidLoginRequest`, `InvalidLoginRequest`, `EmptyLoginRequest`, `ValidLoginResponse`, `EmptyLoginResponse`
- **Signup**: `ValidSignupRequest`, `InvalidSignupRequest`, `EmptySignupRequest`, `ValidSignupResponse`, `EmptySignupResponse`
- **Users**: `ValidUserResponse`, `AdminUserResponse`, `EmptyUserResponse`, `InvalidUserResponse`
- **Roles**: `ValidRoleResponse`, `AdminRoleResponse`, `EmptyRoleResponse`
- **Sessions**: `ValidSessionResponse`, `ExpiredSessionResponse`, `EmptySessionResponse`
- **Tokens**: `ValidTokenResponse`, `ExpiredTokenResponse`, `EmptyTokenResponse`
- **AuditLogs**: `ValidAuditLogResponse`, `FailedAuditLogResponse`, `EmptyAuditLogResponse`

### AuthTestUseCases
Contiene casos de uso de prueba:

- **GetUsers**: `ValidGetUsersRequest`, `InvalidGetUsersRequest`, `EmptyGetUsersRequest`
- **Responses**: `ValidUsersListResponse`, `EmptyUsersListResponse`, `InvalidUsersListResponse`

### AuthTestScenarios
Contiene escenarios de prueba completos:

- **Login Scenarios**: `SuccessfulLogin`, `FailedLogin`, `AccountLocked`, `AccountInactive`, `InvalidCredentials`
- **Signup Scenarios**: `SuccessfulSignup`, `FailedSignup`, `DuplicateEmail`, `InvalidData`
- **GetUsers Scenarios**: `SuccessfulGetUsers`, `FailedGetUsers`, `EmptyUsers`, `PaginationError`

## Uso

### Inicialización

```go
import (
    "testing"
    "github.com/yovannylopez/docsy-main/internal/auth/test_utils"
)

func TestAuthFunction(t *testing.T) {
    // Crear instancia de stubs
    stubs := test_utils.NewAuthStubs()
    
    // Usar stubs en tests
    user := stubs.Entities.ValidUser
    loginRequest := stubs.DTOs.ValidLoginRequest
    // ...
}
```

### Obtener Entidades Específicas

```go
// Obtener usuario válido
user := stubs.GetTestEntity("valid_user").(*entities.User)

// Obtener usuario admin
adminUser := stubs.GetTestEntity("admin_user").(*entities.User)

// Obtener usuario bloqueado
lockedUser := stubs.GetTestEntity("locked_user").(*entities.User)
```

### Obtener DTOs Específicos

```go
// Obtener request de login válido
loginRequest := stubs.GetTestDTO("valid_login_request").(*dtos.LoginRequest)

// Obtener request de signup válido
signupRequest := stubs.GetTestDTO("valid_signup_request").(*dtos.SignupRequest)

// Obtener response de login válido
loginResponse := stubs.GetTestDTO("valid_login_response").(*dtos.LoginResponse)
```

### Obtener Casos de Uso Específicos

```go
// Obtener request de obtención de usuarios válido
getUsersRequest := stubs.GetTestUseCase("valid_get_users_request").(*usecases.GetUsersRequest)

// Obtener response de lista de usuarios válido
usersListResponse := stubs.GetTestUseCase("valid_users_list_response").(*dtos.UsersListResponse)
```

### Obtener Escenarios Específicos

```go
// Obtener escenario de login exitoso
successfulLogin := stubs.GetTestScenario("successful_login").(AuthLoginScenario)

// Obtener escenario de signup exitoso
successfulSignup := stubs.GetTestScenario("successful_signup").(AuthSignupScenario)

// Obtener escenario de obtención de usuarios exitoso
successfulGetUsers := stubs.GetTestScenario("successful_get_users").(AuthGetUsersScenario)
```

### Crear Datos Personalizados

```go
// Crear usuario mock personalizado
customUser := stubs.CreateMockUser("custom@example.com", "Custom", "User", true, true)

// Crear request de login mock personalizado
customLoginRequest := stubs.CreateMockLoginRequest("test@example.com", "password123")

// Crear request de signup mock personalizado
customSignupRequest := stubs.CreateMockSignupRequest("new@example.com", "password123", "New", "User", "user")

// Crear request de obtención de usuarios mock personalizado
customGetUsersRequest := stubs.CreateMockGetUsersRequest(20, 10)
```

### Obtener Listas de Datos

```go
// Obtener lista de usuarios de prueba
users := stubs.GetTestUsers(5) // Retorna 5 usuarios diferentes

// Obtener lista de roles de prueba
roles := stubs.GetTestRoles(3) // Retorna 3 roles diferentes
```

## Ejemplos de Uso en Tests

### Test de Login Exitoso

```go
func TestLoginUseCase_Execute_Success(t *testing.T) {
    stubs := test_utils.NewAuthStubs()
    scenario := stubs.Scenarios.SuccessfulLogin
    
    // Configurar mocks
    mockRepo := mocks.NewUserRepository(t)
    mockRepo.On("FindByEmail", mock.Anything, scenario.Request.Email).Return(scenario.User, nil)
    
    // Ejecutar test
    useCase := NewLoginUseCase(mockRepo, ...)
    response, err := useCase.Execute(context.Background(), scenario.Request, scenario.UserAgent, scenario.IPAddress)
    
    // Verificar resultados
    assert.NoError(t, err)
    assert.NotNil(t, response)
    // ... más verificaciones
}
```

### Test de Login Fallido

```go
func TestLoginUseCase_Execute_InvalidCredentials(t *testing.T) {
    stubs := test_utils.NewAuthStubs()
    scenario := stubs.Scenarios.InvalidCredentials
    
    // Configurar mocks
    mockRepo := mocks.NewUserRepository(t)
    mockRepo.On("FindByEmail", mock.Anything, scenario.Request.Email).Return(scenario.User, nil)
    
    // Ejecutar test
    useCase := NewLoginUseCase(mockRepo, ...)
    response, err := useCase.Execute(context.Background(), scenario.Request, scenario.UserAgent, scenario.IPAddress)
    
    // Verificar resultados
    assert.Error(t, err)
    assert.Nil(t, response)
    assert.Equal(t, scenario.ExpectedError, err.Error())
}
```

### Test de Obtención de Usuarios

```go
func TestGetUsersUseCase_Execute_Success(t *testing.T) {
    stubs := test_utils.NewAuthStubs()
    scenario := stubs.Scenarios.SuccessfulGetUsers
    
    // Configurar mocks
    mockRepo := mocks.NewUserRepository(t)
    mockRepo.On("GetAllUsers", mock.Anything, scenario.Request.Limit, scenario.Request.Offset).Return(scenario.Users, nil)
    
    // Ejecutar test
    useCase := NewGetUsersUseCase(mockRepo)
    response, err := useCase.Execute(context.Background(), scenario.Request)
    
    // Verificar resultados
    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.Equal(t, scenario.ExpectedResponse.Total, response.Total)
    // ... más verificaciones
}
```

## Tipos de Entidades Disponibles

### Users
- `valid_user`: Usuario válido con datos completos
- `admin_user`: Usuario administrador con MFA habilitado
- `inactive_user`: Usuario inactivo
- `locked_user`: Usuario bloqueado por intentos fallidos
- `unverified_user`: Usuario no verificado
- `user_with_mfa`: Usuario con MFA habilitado
- `user_with_phone`: Usuario con número de teléfono
- `empty_user`: Usuario con datos vacíos
- `invalid_user`: Usuario con datos inválidos

### Roles
- `valid_role`: Rol válido
- `admin_role`: Rol de administrador
- `user_role`: Rol de usuario regular
- `guest_role`: Rol de invitado
- `empty_role`: Rol con datos vacíos
- `invalid_role`: Rol con datos inválidos

### Sessions
- `valid_session`: Sesión válida y activa
- `expired_session`: Sesión expirada
- `revoked_session`: Sesión revocada
- `empty_session`: Sesión con datos vacíos
- `invalid_session`: Sesión con datos inválidos

### AuthTokens
- `valid_auth_token`: Token válido
- `expired_auth_token`: Token expirado
- `empty_auth_token`: Token con datos vacíos
- `invalid_auth_token`: Token con datos inválidos

## Tipos de DTOs Disponibles

### Login
- `valid_login_request`: Request de login válido
- `invalid_login_request`: Request de login inválido
- `empty_login_request`: Request de login vacío
- `valid_login_response`: Response de login válido
- `empty_login_response`: Response de login vacío

### Signup
- `valid_signup_request`: Request de signup válido
- `invalid_signup_request`: Request de signup inválido
- `empty_signup_request`: Request de signup vacío
- `valid_signup_response`: Response de signup válido
- `empty_signup_response`: Response de signup vacío

### Users
- `valid_user_response`: Response de usuario válido
- `admin_user_response`: Response de usuario administrador
- `empty_user_response`: Response de usuario vacío
- `invalid_user_response`: Response de usuario inválido

## Escenarios Disponibles

### Login Scenarios
- `successful_login`: Login exitoso con datos completos
- `failed_login`: Login fallido por credenciales incorrectas
- `account_locked`: Login fallido por cuenta bloqueada
- `account_inactive`: Login fallido por cuenta inactiva
- `invalid_credentials`: Login fallido por credenciales inválidas

### Signup Scenarios
- `successful_signup`: Signup exitoso
- `failed_signup`: Signup fallido por validación
- `duplicate_email`: Signup fallido por email duplicado
- `invalid_data`: Signup fallido por datos inválidos

### GetUsers Scenarios
- `successful_get_users`: Obtención exitosa de usuarios
- `failed_get_users`: Obtención fallida por error de base de datos
- `empty_users`: Obtención exitosa sin usuarios
- `pagination_error`: Error en parámetros de paginación

## Beneficios

1. **Consistencia**: Todos los tests usan los mismos datos de prueba
2. **Mantenibilidad**: Cambios en los datos se hacen en un solo lugar
3. **Completitud**: Cubre todos los casos de uso y escenarios
4. **Reutilización**: Los stubs se pueden usar en múltiples tests
5. **Legibilidad**: Los tests son más claros y fáciles de entender
6. **Productividad**: Reduce el tiempo de escritura de tests

## Convenciones

- Los nombres de los stubs siguen el patrón `{Tipo}{Estado}` (ej: `ValidUser`, `InvalidLoginRequest`)
- Los escenarios incluyen todos los datos necesarios para el test completo
- Los helper methods permiten obtener datos específicos por tipo
- Los métodos de creación permiten personalizar datos según necesidades específicas
