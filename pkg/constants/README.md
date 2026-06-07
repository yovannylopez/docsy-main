# Librería de Constantes

Esta librería centraliza todas las constantes, mensajes estándar y configuraciones comunes utilizadas en los microservicios del proyecto.

## Frontera con otros `pkg/`

- **`pkg/http_status`**: los textos `SuccessMessage`, `CreatedMessage` e `InternalErrorMessage` coinciden con las descripciones de `OK`, `Created` e `InternalError` para respuestas JSON coherentes. Los **códigos HTTP numéricos** siguen viviendo en `http_status` / `net/http`, no aquí.
- **`pkg/responses`**: `responses.JSON` fija `Content-Type` con `constants.ContentTypeJSON`.
- **`pkg/pagination`**: `DefaultConfig` usa `DefaultPageSize` y `MaxPageSize` de este paquete.
- **`pkg/validators`**: mensajes de validación de campo (p. ej. `campo requerido`); aquí viven mensajes **de negocio** y códigos lógicos (`VALIDATION_ERROR`, etc.), no la gramática de cada regla de `StringValidator`.
- **`internal/shared/transport/middleware`**: `RequestIDMiddleware` y el error handler central usan `RequestIDHeader`.
- **Health / readiness**: mensajes largos en inglés (`ServiceReadyToReceiveTrafficMessage`, …) y slugs JSON (`HealthStatusHealthy`) para payloads de probes; el resto de mensajes de negocio suelen estar en español.

## Estructura

```
constants/
├── constants.go    # Constantes estáticas
├── messages.go     # Funciones de mensajes dinámicos (sin estado)
├── go.mod          # Dependencias del módulo
└── README.md       # Esta documentación
```

## Uso

### Constantes Estáticas

```go
import "github.com/yovannylopez/docsy-main/pkg/constants"

// Usar constantes de mensajes
if user == nil {
    return constants.UserNotFoundMessage
}

// Usar constantes de headers
c.Response().Header().Set(constants.RequestIDHeader, requestID)

// Usar constantes de roles
if user.Role == constants.RoleAdmin {
    // Lógica de administrador
}
```

### Mensajes dinámicos

Son **funciones de paquete** en `messages.go` (sin estado ni `New*`):

```go
import "github.com/yovannylopez/docsy-main/pkg/constants"

errorMsg := constants.UserNotFoundWithID("123")
// "Usuario con ID '123' no encontrado"

validationMsg := constants.FieldRequired("email")
// "El campo 'email' es requerido"

fileMsg := constants.FileTooLarge("document.pdf", 10)
// "El archivo 'document.pdf' excede el tamaño máximo de 10 MB"
```

## Tipos de Constantes

### Códigos de Respuesta
- `SuccessCode`
- `ErrorCode`
- `ValidationError`
- `NotFoundError`
- `ConflictError`
- `UnauthorizedError`
- `ForbiddenError`
- `InternalError`

### Mensajes de Éxito
- `SuccessMessage`
- `CreatedMessage`
- `UpdatedMessage`
- `DeletedMessage`

### Mensajes de Error
- `GenericErrorMessage`
- `ValidationErrorMessage`
- `NotFoundMessage`
- `ConflictMessage`
- `UnauthorizedMessage`
- `ForbiddenMessage`
- `InternalErrorMessage`

### Mensajes de Autenticación
- `InvalidCredentialsMessage`
- `TokenExpiredMessage`
- `TokenInvalidMessage`
- `LoginRequiredMessage`

### Mensajes Específicos por Dominio
- `UserNotFoundMessage`
- `UserAlreadyExistsMessage`
- `UserCreatedMessage`
- `EmailAlreadyExistsMessage`
- `DocumentNotFoundMessage`
- `DocumentCreatedMessage`
- `RoleNotFoundMessage`
- `PermissionDeniedMessage`

### Headers Estándar
- `RequestIDHeader`
- `AuthorizationHeader`
- `ContentTypeHeader`
- `UserAgentHeader`

### Valores de Headers
- `ContentTypeJSON`
- `BearerTokenType` (valor `token_type` en respuestas JWT/OAuth, sin espacio)
- `BearerPrefix` (prefijo del header `Authorization`, con espacio final)

### Límites y Configuraciones
- `MaxFileSizeMB`
- `MaxRequestSizeMB`
- `DefaultPageSize`
- `MaxPageSize`
- `MaxUsersLimit`, `MaxUsersBatchSize` (listados y creación masiva de usuarios)

### Autenticación (mensajes HTTP)
- `InvalidCredentialsMessage`, `LoginEmailPasswordRequiredMessage`, `LoginSuccessMessage`

### Health / readiness
- `HealthStatusHealthy`, `HealthStatusUnhealthy`
- `ServiceReadyToReceiveTrafficMessage`, `ServiceNotReadyToReceiveTrafficMessage`
- `ServiceHealthyMessage`, `ServiceReadyMessage`, `ServiceAliveMessage` (otros textos de probes)

### Formatos de Fecha
- `DateFormat`
- `DateTimeFormat`
- `TimeFormat`

### Estados
- `StatusActive`
- `StatusInactive`
- `StatusPending`
- `StatusDeleted`

### Tipos de Contenido
- `ContentTypeDocument`
- `ContentTypeImage`
- `ContentTypeVideo`
- `ContentTypeAudio`

### Roles del Sistema
- `RoleAdmin`
- `RoleUser`
- `RoleManager`
- `RoleViewer`

### Permisos del Sistema
- `PermissionRead`
- `PermissionWrite`
- `PermissionDelete`
- `PermissionAdmin`

## Funciones de mensajes dinámicos (`messages.go`)

### Usuarios
- `UserNotFoundWithID(userID string)`
- `UserNotFoundWithEmail(email string)`

### Documentos
- `DocumentNotFoundWithID(documentID string)`

### Validación
- `FieldRequired(fieldName string)`
- `FieldInvalid(fieldName, reason string)`
- `FieldTooLong(fieldName string, maxLength int)`
- `FieldTooShort(fieldName string, minLength int)`
- `ValidationFailedWithDetails(details []string)`

### Archivos
- `FileTooLarge(fileName string, maxSizeMB int)`
- `InvalidFileFormat(fileName string, allowedFormats []string)`

### Permisos
- `PermissionDeniedForResource(userID, resource, action string)`

### Recursos
- `ResourceCreatedWithID(resourceType, resourceID string)`
- `ResourceUpdatedWithID(resourceType, resourceID string)`
- `ResourceDeletedWithID(resourceType, resourceID string)`

### Servicios
- `RateLimitExceeded(limit, window string)`
- `ServiceUnavailable(serviceName string)`
- `DatabaseConnectionError(details string)`
- `ExternalServiceError(serviceName, details string)`

## Beneficios

1. **Consistencia**: Todos los servicios usan los mismos mensajes
2. **Mantenibilidad**: Cambios centralizados en un solo lugar
3. **Reutilización**: Una librería para múltiples servicios
4. **Internacionalización**: Fácil agregar soporte para múltiples idiomas
5. **Testing**: Fácil testing de mensajes
6. **Documentación**: Mensajes bien documentados y tipados

## Agregar Nuevas Constantes

Para agregar nuevas constantes:

1. Agregar la constante en `constants.go`
2. Si es un mensaje dinámico, agregar la función en `messages.go`
3. Documentar el uso en este README
4. Añadir o ampliar tests en este módulo (`constants_test.go`, `messages_test.go`) cuando cambie el contrato público

## Ejemplo de Uso en Handler

```go
package handlers

import (
    "github.com/labstack/echo/v4"

    "github.com/yovannylopez/docsy-main/pkg/constants"
    "github.com/yovannylopez/docsy-main/pkg/responses"
)

func (h *UserHandler) GetUser(c echo.Context) error {
    userID := c.Param("id")
    user, err := h.userRepo.FindByID(c.Request().Context(), userID)
    if err != nil {
        return responses.NotFound(c, constants.UserNotFoundWithID(userID))
    }
    return responses.OK(c, user, constants.SuccessMessage)
}
``` 