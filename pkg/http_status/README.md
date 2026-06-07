# HTTP Status

Una librería simple y eficiente para manejar códigos de estado HTTP en aplicaciones Go. Proporciona constantes predefinidas para los códigos de estado más comunes, junto con métodos útiles para validar y trabajar con respuestas HTTP.

## Características

- ✅ **Códigos de estado predefinidos**: Constantes para todos los códigos HTTP comunes
- ✅ **Validación de tipos**: Métodos para verificar si un código es exitoso, error de cliente o error de servidor
- ✅ **Códigos personalizados**: Función para crear códigos de estado personalizados
- ✅ **Texto oficial HTTP**: Acceso al texto oficial de los códigos de estado
- ✅ **Dependencias mínimas en producción**: el código del paquete solo usa la biblioteca estándar (`net/http`); los tests del módulo usan `testify`

## Frontera con `pkg/responses`

Las respuestas HTTP del proyecto (`pkg/responses`) incluyen `Status *http_status.Status` para unificar código, descripción en JSON y helpers (`EchoJSON`, `EchoError`, etc.). Este paquete aporta el **valor semántico** (textos en español para documentación/OpenAPI y payloads); los códigos numéricos coinciden con `net/http`. No sustituye a `http.StatusText` para el texto en inglés oficial: para eso está `GetHTTPStatusText()`.

## Instalación

```bash
go get github.com/yovannylopez/docsy-main/pkg/http_status
```

## Uso Básico

### Importar la librería

```go
import "github.com/yovannylopez/docsy-main/pkg/http_status"
```

### Usar códigos de estado predefinidos

```go
// Respuestas exitosas
status := http_status.OK
fmt.Printf("Código: %d, Descripción: %s\n", status.Code, status.Description)
// Output: Código: 200, Descripción: Operación exitosa

// Errores del cliente
notFound := http_status.NotFound
fmt.Printf("Código: %d, Descripción: %s\n", notFound.Code, notFound.Description)
// Output: Código: 404, Descripción: Recurso no encontrado

// Errores del servidor
internalError := http_status.InternalError
fmt.Printf("Código: %d, Descripción: %s\n", internalError.Code, internalError.Description)
// Output: Código: 500, Descripción: Error interno del servidor
```

### Crear códigos personalizados

```go
// Crear un código de estado personalizado
customStatus := http_status.Custom(418, "Soy una tetera")
fmt.Printf("Código: %d, Descripción: %s\n", customStatus.Code, customStatus.Description)
// Output: Código: 418, Descripción: Soy una tetera
```

### Validar tipos de códigos

```go
status := http_status.OK

// Verificar si es exitoso (2xx)
if status.IsSuccess() {
    fmt.Println("La operación fue exitosa")
}

// Verificar si es error de cliente (4xx)
if status.IsClientError() {
    fmt.Println("Error del cliente")
}

// Verificar si es error de servidor (5xx)
if status.IsServerError() {
    fmt.Println("Error del servidor")
}
```

### Obtener texto oficial HTTP

```go
status := http_status.NotFound
officialText := status.GetHTTPStatusText()
fmt.Println(officialText) // Output: "Not Found"
```

### Obtener todos los códigos comunes

```go
commonCodes := http_status.CommonStatusCodes()
for name, status := range commonCodes {
    fmt.Printf("%s: %d - %s\n", name, status.Code, status.Description)
}
```

### Buscar por código numérico

```go
if s, ok := http_status.LookupByCode(404); ok {
    _ = s.Description // "Recurso no encontrado"
}
```

### Puntero a `Status` (p. ej. `*http_status.Status`)

```go
st := http_status.Ptr(http_status.OK) // copia en heap; distinto de `&http_status.OK` cuando necesitas identidad única
```

### Mensaje del envelope de error central (inglés)

La constante `http_status.EnvelopeInternalServerErrorMessageEN` es el texto del campo `message` para errores genéricos no mapeados en `CentralHTTPErrorHandler` (`internal/shared/transport/middleware`). No es la `Description` en español de `InternalError`.

## Códigos de Estado Disponibles

### Respuestas Exitosas (2xx)

| Constante | Código | Descripción |
|-----------|--------|-------------|
| `OK` | 200 | Operación exitosa |
| `Created` | 201 | Recurso creado exitosamente |
| `Accepted` | 202 | Solicitud aceptada |
| `NoContent` | 204 | Sin contenido |

### Errores del Cliente (4xx)

| Constante | Código | Descripción |
|-----------|--------|-------------|
| `BadRequest` | 400 | Solicitud inválida |
| `Unauthorized` | 401 | No autorizado |
| `Forbidden` | 403 | Prohibido |
| `NotFound` | 404 | Recurso no encontrado |
| `MethodNotAllowed` | 405 | Método no permitido |
| `Conflict` | 409 | Conflicto |
| `UnprocessableEntity` | 422 | Entidad no procesable |
| `TooManyRequests` | 429 | Demasiadas solicitudes |

### Errores del Servidor (5xx)

| Constante | Código | Descripción |
|-----------|--------|-------------|
| `InternalError` | 500 | Error interno del servidor |
| `NotImplemented` | 501 | No implementado |
| `BadGateway` | 502 | Puerta de enlace incorrecta |
| `ServiceUnavailable` | 503 | Servicio no disponible |
| `GatewayTimeout` | 504 | Tiempo de espera agotado |

## API Reference

### Estructura Status

```go
type Status struct {
    Code        int    `json:"code"`
    Description string `json:"description"`
}
```

### Funciones

#### `Custom(code int, description string) Status`
Crea un código de estado personalizado.

**Parámetros:**
- `code`: Código de estado HTTP (int)
- `description`: Descripción del código (string)

**Retorna:** `Status`

#### `CommonStatusCodes() map[string]Status`
Retorna un mapa con todos los códigos de estado comunes.

**Retorna:** `map[string]Status`

### Métodos

#### `(s Status) IsSuccess() bool`
Verifica si el código de estado es exitoso (2xx).

**Retorna:** `bool`

#### `(s Status) IsClientError() bool`
Verifica si el código de estado es un error del cliente (4xx).

**Retorna:** `bool`

#### `(s Status) IsServerError() bool`
Verifica si el código de estado es un error del servidor (5xx).

**Retorna:** `bool`

#### `(s Status) GetHTTPStatusText() string`
Retorna el texto oficial del código de estado HTTP.

**Retorna:** `string`

## Ejemplos de Uso

### En una API REST

```go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/yovannylopez/docsy-main/pkg/http_status"
)

func handleGetUser(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    
    if userID == "" {
        response := map[string]any{
            "status": http_status.BadRequest,
            "message": "ID de usuario requerido",
        }
        
        w.WriteHeader(http_status.BadRequest.Code)
        json.NewEncoder(w).Encode(response)
        return
    }
    
    // Simular búsqueda de usuario
    if userID == "123" {
        response := map[string]any{
            "status": http_status.OK,
            "data": map[string]string{
                "id": "123",
                "name": "Juan Pérez",
                "email": "juan@example.com",
            },
        }
        
        w.WriteHeader(http_status.OK.Code)
        json.NewEncoder(w).Encode(response)
        return
    }
    
    // Usuario no encontrado
    response := map[string]any{
        "status": http_status.NotFound,
        "message": "Usuario no encontrado",
    }
    
    w.WriteHeader(http_status.NotFound.Code)
    json.NewEncoder(w).Encode(response)
}
```

### Validación de respuestas

```go
func validateResponse(status http_status.Status) string {
    switch {
    case status.IsSuccess():
        return "Operación exitosa"
    case status.IsClientError():
        return "Error del cliente - revisar solicitud"
    case status.IsServerError():
        return "Error del servidor - contactar soporte"
    default:
        return "Código de estado desconocido"
    }
}
```
