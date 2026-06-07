# Response

Una librería completa para manejar respuestas HTTP estandarizadas en aplicaciones Go. Proporciona estructuras de respuesta consistentes, manejo de errores tipados y integración con el framework Echo para crear APIs REST robustas y bien estructuradas.

## Características

- ✅ **Respuestas estandarizadas**: Estructura consistente para todas las respuestas HTTP
- ✅ **Integración con Echo**: Funciones helper específicas para el framework Echo
- ✅ **Errores tipados**: Sistema de errores con tipos predefinidos
- ✅ **Códigos de estado HTTP**: Integración con la librería http_status
- ✅ **Listados paginados**: `OKPaginated` + `pkg/pagination` para `data` + `pagination` en raíz
- ✅ **`BindJSON`**: lectura acotada (por defecto **1 MiB**); `errors.Is(..., responses.ErrJSONBodyTooLarge)` para cuerpos excesivos
- ⚠️ **`Validate`**: **deprecated** (no-op); usar validación en handlers / `validator`
- ✅ **`EchoAppError`** / **`BadRequestAppError`**: error estándar + `data` con el `AppError` completo
- ✅ **`ToHTTPAppError(err)`**: traduce errores de **`pkg/errors`** (cadena con `GetAppError` / `errors.As`) al `AppError` HTTP de este paquete, con mensajes genéricos en español para fallos de BD/internos (sin filtrar el `cause` al cliente).
- ✅ **Metadatos flexibles**: Soporte para información adicional en respuestas
- ✅ **Request ID tracking**: Seguimiento de requests para debugging

En **`cmd/composition`**, `echo.Echo.HTTPErrorHandler` está configurado con **`CentralHTTPErrorHandler`**, que aplica `*responses.AppError`, `ToHTTPAppError` y `*echo.HTTPError` de forma uniforme.

## Instalación

```bash
go get github.com/yovannylopez/docsy-main/pkg/responses
```

## Dependencias

```go
require (
    github.com/labstack/echo/v4 v4.13.4
    github.com/yovannylopez/docsy-main/pkg/http_status v0.0.0-00010101000000-000000000000
)
```

## Uso Básico

### Importar la librería

```go
import "github.com/yovannylopez/docsy-main/pkg/responses"
```

### Respuestas exitosas

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/yovannylopez/docsy-main/pkg/responses"
    "github.com/yovannylopez/docsy-main/pkg/http_status"
)

func handleGetUser(c echo.Context) error {
    user := User{
        ID:    "123",
        Name:  "Juan Pérez",
        Email: "juan@example.com",
    }
    
    // Respuesta exitosa con datos
    return responses.OK(c, user, "Usuario obtenido exitosamente")

    // O usando la función genérica
    // return responses.EchoJSON(c, &http_status.OK, user, "Usuario obtenido exitosamente")
}

func handleCreateUser(c echo.Context) error {
    user := User{
        ID:    "124",
        Name:  "María García",
        Email: "maria@example.com",
    }
    
    // Respuesta de creación
    return responses.Created(c, user, "Usuario creado exitosamente")
}
```

### Respuestas de error

```go
func handleGetUser(c echo.Context) error {
    userID := c.Param("id")
    
    if userID == "" {
        return responses.BadRequest(c, "ID de usuario requerido")
    }
    
    user, err := getUserByID(userID)
    if err != nil {
        return responses.NotFound(c, "Usuario no encontrado")
    }
    
    if user == nil {
        return responses.NotFound(c, "Usuario no encontrado")
    }
    
    return responses.OK(c, user, "Usuario obtenido exitosamente")
}
```

### Usar errores tipados

```go
func handleCreateUser(c echo.Context) error {
    var user User
    if err := c.Bind(&user); err != nil {
        appError := responses.NewBadRequestError("Datos de usuario inválidos")
        return responses.EchoError(c, &http_status.BadRequest, appError.Error())
    }
    
    // Validar email único
    if emailExists(user.Email) {
        appError := responses.NewDuplicateError("email")
        return responses.EchoError(c, &http_status.Conflict, appError.Error())
    }
    
    // Crear usuario
    createdUser, err := createUser(user)
    if err != nil {
        appError := responses.NewInternalError("Error creando usuario")
        return responses.EchoError(c, &http_status.InternalError, appError.Error())
    }
    
    return responses.Created(c, createdUser, "Usuario creado exitosamente")
}
```

## Estructuras de Respuesta

### Response

```go
type Response struct {
    Status  *http_status.Status `json:"status"`
    Message string              `json:"message,omitempty"`
    Data    any                 `json:"data,omitempty"`
    Error   string              `json:"error,omitempty"`
    Meta    map[string]any      `json:"meta,omitempty"`
}
```

### AppError

```go
type AppError struct {
    Type        ErrorType      `json:"type"`
    Code        int            `json:"code"`
    Message     string         `json:"message"`
    Description string         `json:"description,omitempty"`
    Details     map[string]any `json:"details,omitempty"`
    RequestID   string         `json:"request_id,omitempty"`
}
```

## Tipos de Error

### Errores de Validación
- `ValidationError`: Errores de validación de datos

### Errores de Autenticación
- `AuthenticationError`: Errores de autenticación
- `AuthorizationError`: Errores de autorización

### Errores de Recursos
- `NotFoundError`: Recurso no encontrado
- `ConflictError`: Conflicto de recursos
- `DuplicateError`: Recurso duplicado

### Errores del Sistema
- `InternalServerError`: Error interno del servidor
- `ServiceError`: Error de servicio
- `DatabaseError`: Error de base de datos

### Errores de Entrada
- `BadRequestError`: Solicitud inválida
- `UnprocessableError`: Entidad no procesable

## API Reference

### Funciones de Respuesta

#### `SuccessResponse(status *http_status.Status, data any, message string) *Response`
Crea una respuesta exitosa.

#### `ErrorResponse(status *http_status.Status, error string) *Response`
Crea una respuesta de error.

#### `JSON(w http.ResponseWriter, status int, data any) error`
Escribe JSON en `http.ResponseWriter`. Devuelve error si `json.Encode` falla.

#### `EchoJSON(c echo.Context, status *http_status.Status, data any, message string) error`
Envía una respuesta JSON usando Echo.

#### `EchoError(c echo.Context, status *http_status.Status, error string) error`
Envía una respuesta de error usando Echo.

#### `EchoAppError(c echo.Context, status *http_status.Status, appErr *AppError) error`
Igual que el envoltorio estándar, con `error` = `Message` y `data` = objeto `AppError` (tipo, detalles, etc.).

#### `BadRequestAppError(c echo.Context, appErr *AppError) error`
Atajo de `EchoAppError` con estado **400**.

#### `BindJSON(r *http.Request, v any) error`
Lee y decodifica JSON con límite **`DefaultMaxJSONBodyBytes`** (1 MiB). Errores envueltos; **`ErrJSONBodyTooLarge`** si el cuerpo supera el máximo.

#### `BindJSONWithLimit(r *http.Request, v any, maxBytes int64) error`
Igual que `BindJSON` con límite explícito (`maxBytes >= 1`). Cuerpo vacío devuelve **`io.EOF`** envuelto.

### Funciones Helper para Echo

#### `OK(c echo.Context, data any, message string) error`
Respuesta 200 OK.

#### `OKPaginated(c echo.Context, message string, body pagination.Response) error`
Respuesta 200 con `status`, `message`, `data` y `pagination` (metadata de `pkg/pagination`).

#### `Created(c echo.Context, data any, message string) error`
Respuesta 201 Created.

#### `BadRequest(c echo.Context, error string) error`
Respuesta 400 Bad Request.

#### `Unauthorized(c echo.Context, error string) error`
Respuesta 401 Unauthorized.

#### `Forbidden(c echo.Context, error string) error`
Respuesta 403 Forbidden.

#### `NotFound(c echo.Context, error string) error`
Respuesta 404 Not Found.

#### `Conflict(c echo.Context, error string) error`
Respuesta 409 Conflict.

#### `UnprocessableEntity(c echo.Context, error string) error`
Respuesta 422 Unprocessable Entity.

#### `InternalError(c echo.Context, error string) error`
Respuesta 500 Internal Server Error.

#### `NotImplemented(c echo.Context, error string) error`
Respuesta 501 Not Implemented.

### Funciones de Error

#### `NewAppError(errorType ErrorType, code int, message string) *AppError`
Crea un nuevo error de aplicación.

#### `NewValidationError(message string) *AppError`
Crea un error de validación.

#### `NewAuthenticationError(message string) *AppError`
Crea un error de autenticación.

#### `NewAuthorizationError(message string) *AppError`
Crea un error de autorización.

#### `NewNotFoundError(resource string) *AppError`
Crea un error de recurso no encontrado.

#### `NewConflictError(message string) *AppError`
Crea un error de conflicto.

#### `NewDuplicateError(field string) *AppError`
Crea un error de duplicado.

#### `NewInternalError(message string) *AppError`
Crea un error interno del servidor (HTTP 500); suele usarse con `constants.InternalErrorMessage`.

#### `NewBadRequestError(message string) *AppError`
Crea un error de solicitud inválida.

### Métodos de AppError

#### `WithDescription(description string) *AppError`
Agrega una descripción al error.

#### `WithDetails(details map[string]any) *AppError`
Agrega detalles adicionales al error.

#### `WithRequestID(requestID string) *AppError`
Agrega un ID de request al error.

## Ejemplos de Uso

### API REST Completa

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/yovannylopez/docsy-main/pkg/responses"
)

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    e := echo.New()
    
    // Rutas
    e.GET("/users/:id", getUser)
    e.POST("/users", createUser)
    e.PUT("/users/:id", updateUser)
    e.DELETE("/users/:id", deleteUser)
    
    e.Start(":8080")
}

func getUser(c echo.Context) error {
    userID := c.Param("id")
    
    if userID == "" {
        return responses.BadRequest(c, "ID de usuario requerido")
    }
    
    user, err := getUserByID(userID)
    if err != nil {
        return responses.InternalError(c, "Error obteniendo usuario")
    }
    
    if user == nil {
        return responses.NotFound(c, "Usuario no encontrado")
    }
    
    return responses.OK(c, user, "Usuario obtenido exitosamente")
}

func createUser(c echo.Context) error {
    var user User
    if err := c.Bind(&user); err != nil {
        return responses.BadRequest(c, "Datos de usuario inválidos")
    }
    
    // Validaciones
    if user.Name == "" {
        return responses.BadRequest(c, "Nombre es requerido")
    }
    
    if user.Email == "" {
        return responses.BadRequest(c, "Email es requerido")
    }
    
    // Verificar email único
    if emailExists(user.Email) {
        return responses.Conflict(c, "Email ya existe")
    }
    
    // Crear usuario
    createdUser, err := createUserInDB(user)
    if err != nil {
        return responses.InternalError(c, "Error creando usuario")
    }
    
    return responses.Created(c, createdUser, "Usuario creado exitosamente")
}

func updateUser(c echo.Context) error {
    userID := c.Param("id")
    
    var user User
    if err := c.Bind(&user); err != nil {
        return responses.BadRequest(c, "Datos de usuario inválidos")
    }
    
    user.ID = userID // Asegurar que el ID sea el correcto
    
    updatedUser, err := updateUserInDB(user)
    if err != nil {
        return responses.InternalError(c, "Error actualizando usuario")
    }
    
    return responses.OK(c, updatedUser, "Usuario actualizado exitosamente")
}

func deleteUser(c echo.Context) error {
    userID := c.Param("id")
    
    if err := deleteUserFromDB(userID); err != nil {
        return responses.InternalError(c, "Error eliminando usuario")
    }
    
    return responses.OK(c, nil, "Usuario eliminado exitosamente")
}
```

### Manejo de Errores Avanzado

```go
func handleComplexOperation(c echo.Context) error {
    requestID := c.Request().Header.Get("X-Request-ID")
    
    // Simular operación compleja
    result, err := performComplexOperation()
    if err != nil {
        appError := responses.NewInternalError("Error en operación compleja")
        appError.WithRequestID(requestID)
        appError.WithDetails(map[string]any{
            "operation": "complex_operation",
            "timestamp": time.Now().Unix(),
        })
        
        return responses.EchoError(c, &http_status.InternalError, appError.Error())
    }
    
    return responses.OK(c, result, "Operación completada exitosamente")
}
```

### Validación con Errores Tipados

```go
func validateUser(user User) *responses.AppError {
    if user.Name == "" {
        return responses.NewValidationError("Nombre es requerido")
    }
    
    if user.Email == "" {
        return responses.NewValidationError("Email es requerido")
    }
    
    if !isValidEmail(user.Email) {
        return responses.NewValidationError("Email inválido")
    }
    
    if user.Age < 0 || user.Age > 150 {
        return responses.NewValidationError("Edad inválida")
    }
    
    return nil
}

func createUserWithValidation(c echo.Context) error {
    var user User
    if err := c.Bind(&user); err != nil {
        return responses.BadRequest(c, "Datos inválidos")
    }

    if appError := validateUser(user); appError != nil {
        return responses.EchoError(c, &http_status.BadRequest, appError.Error())
    }

    createdUser, err := createUserInDB(user)
    if err != nil {
        return responses.InternalError(c, "Error creando usuario")
    }

    return responses.Created(c, createdUser, "Usuario creado exitosamente")
}
```

### Listado paginado (limit/offset) con `pkg/pagination`

```go
import (
    "github.com/labstack/echo/v4"
    "github.com/yovannylopez/docsy-main/pkg/pagination"
    "github.com/yovannylopez/docsy-main/pkg/responses"
)

func listItems(c echo.Context) error {
    parser := pagination.NewDefaultParser()
    params, err := parser.ParseFromQuery(c.QueryParam("limit"), c.QueryParam("offset"))
    if err != nil {
        return responses.BadRequest(c, err.Error())
    }
    rows, total, err := svc.List(c.Request().Context(), params.Limit, params.Offset)
    if err != nil {
        return responses.InternalError(c, err.Error())
    }
    page := pagination.CreateResponse(rows, params, total)
    return responses.OKPaginated(c, "elementos obtenidos", page)
}
```

### Respuesta con `meta` (casos puntuales)

Para datos extra sin bloque `pagination` (p. ej. telemetría), construye `*responses.Response` con `SuccessResponse`, asigna `Meta` y serializa con `c.JSON` usando el código de `http_status`.

## Formato de Respuestas

### Respuesta Exitosa
```json
{
  "status": {
    "code": 200,
    "description": "Operación exitosa"
  },
  "message": "Usuario obtenido exitosamente",
  "data": {
    "id": "123",
    "name": "Juan Pérez",
    "email": "juan@example.com"
  }
}
```

### Respuesta de Error
```json
{
  "status": {
    "code": 400,
    "description": "Solicitud inválida"
  },
  "error": "Email es requerido"
}
```

### Respuesta con Metadatos
```json
{
  "status": {
    "code": 200,
    "description": "Operación exitosa"
  },
  "message": "Usuarios obtenidos exitosamente",
  "data": [
    {
      "id": "123",
      "name": "Juan Pérez",
      "email": "juan@example.com"
    }
  ],
  "meta": {
    "page": "1",
    "limit": "10",
    "total": 50
  }
}
```

## Mejores Prácticas

1. **Usa funciones helper**: Utiliza `responses.OK`, `responses.Created`, `responses.OKPaginated` (listados con limit/offset), etc.

2. **Maneja errores apropiadamente**: Usa los tipos de error apropiados para cada situación.

3. **Incluye metadatos cuando sea útil**: Agrega información adicional como paginación, timestamps, etc.

4. **Valida datos de entrada**: Siempre valida los datos antes de procesarlos.

5. **Usa Request ID**: Incluye Request ID en errores para facilitar el debugging.

6. **Mantén consistencia**: Usa la misma estructura de respuesta en toda tu API.

## Versión

- Go: 1.26.2+ (alineado al workspace raíz)
- Dependencias:
  - `github.com/labstack/echo/v4 v4.13.4`
  - `github.com/yovannylopez/docsy-main/pkg/http_status`
  - `github.com/yovannylopez/docsy-main/pkg/pagination` (para `OKPaginated`)