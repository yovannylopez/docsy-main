# Librería de Paginación

Esta librería proporciona funcionalidad centralizada para manejar paginación en todos los endpoints del microservicio.

## Frontera con otros `pkg/`

- **`pkg/constants`**: `DefaultConfig` usa **`constants.DefaultPageSize`** y **`constants.MaxPageSize`** (no duplicar números mágicos en este paquete).
- **`pkg/responses`**: para listados con `limit`/`offset`, preferir **`responses.OKPaginated(c, mensaje, pagination.CreateResponse(...))`** (JSON con `status`, `message`, `data`, `pagination`). **Municipios/search** usa también `OKPaginated`; el query sigue admitiendo **`page`** además de **`limit`** (validado con `pkg/pagination`).

## Estructura

```
pagination/
├── pagination.go           # Implementación principal
├── pagination_test.go      # Tests unitarios
├── example/
│   ├── main.go             # Ejemplo ejecutable (package main)
│   └── refactor_example.go # Snippet con //go:build ignore (no compila en go test ./...)
├── go.mod
└── README.md
```

## Uso

### Configuración Básica

```go
import "github.com/yovannylopez/docsy-main/pkg/pagination"

// Usar configuración por defecto
parser := pagination.NewDefaultParser()

// O configuración personalizada
config := pagination.Config{
    DefaultLimit: 15,
    MaxLimit:     50,
    MinLimit:     1,
}
parser := pagination.NewParser(config)
```

### Parsing desde Query Parameters

```go
// En tu handler
func (h *MyHandler) List(c echo.Context) error {
    parser := pagination.NewDefaultParser()
    
    // Extraer parámetros de paginación
    params, err := parser.ParseFromQuery(
        c.QueryParam("limit"),
        c.QueryParam("offset"),
    )
    if err != nil {
        return responses.BadRequest(c, err.Error())
    }
    
    // Usar params.Limit y params.Offset en tu usecase
    data, total, err := h.usecase.Execute(ctx, params.Limit, params.Offset)
    if err != nil {
        return responses.InternalError(c, err.Error())
    }
    
    // Respuesta HTTP estándar del Core (metadata en clave "pagination" al nivel raíz)
    page := pagination.CreateResponse(data, params, total)
    return responses.OKPaginated(c, "datos obtenidos exitosamente", page)
}
```

### Crear Respuestas Paginadas

```go
// Datos obtenidos del repositorio
users := []User{...}
total := 150

// Crear respuesta con metadata de paginación
response := pagination.CreateResponse(users, params, total)

// Resultado:
// {
//   "data": [...],
//   "pagination": {
//     "total": 150,
//     "limit": 20,
//     "offset": 0,
//     "total_pages": 8,
//     "current_page": 1,
//     "has_next": true,
//     "has_previous": false
//   }
// }
```

### Utilidades de Conversión

```go
// Convertir página a offset
page := 3
limit := 20
offset := pagination.GetOffsetFromPage(page, limit) // 40

// Convertir offset a página
offset := 40
limit := 20
page := pagination.GetPageFromOffset(offset, limit) // 3
```

## Tipos Principales

### Params
```go
type Params struct {
    Limit  int `json:"limit"`
    Offset int `json:"offset"`
}
```

### Metadata
```go
type Metadata struct {
    Total       int  `json:"total"`
    Limit       int  `json:"limit"`
    Offset      int  `json:"offset"`
    TotalPages  int  `json:"total_pages"`
    CurrentPage int  `json:"current_page"`
    HasNext     bool `json:"has_next"`
    HasPrev     bool `json:"has_previous"`
}
```

### Response
```go
type Response struct {
    Data     any `json:"data"`
    Metadata Metadata    `json:"pagination"`
}
```

### Config
```go
type Config struct {
    DefaultLimit int
    MaxLimit     int
    MinLimit     int
}
```

## Configuración por Defecto

En código, `DefaultConfig` asigna `DefaultLimit` y `MaxLimit` desde **`pkg/constants`**; los valores efectivos son los de `DefaultPageSize` y `MaxPageSize` en ese paquete.

## Errores de validación

`Validate` y el parseo devuelven errores envueltos con **`errors.Is`** frente a **`pagination.ErrLimitOutOfRange`** y **`pagination.ErrNegativeOffset`** cuando aplica.

## Validaciones

- **Limit**: Debe estar entre `MinLimit` y `MaxLimit`
- **Offset**: No puede ser negativo
- **Query Parameters**: Deben ser números enteros válidos

## Beneficios

### ✅ Eliminación de Duplicación
- Un solo lugar para manejar lógica de paginación
- Código consistente en todos los endpoints
- Fácil mantenimiento y actualizaciones

### ✅ Consistencia
- Mismos valores por defecto en toda la aplicación
- Formato de respuesta estandarizado
- Validaciones uniformes

### ✅ Escalabilidad
- Fácil agregar nuevas funcionalidades de paginación
- Configuración centralizada
- Testing centralizado

### ✅ Flexibilidad
- Configuración personalizable por módulo
- Soporte para diferentes estrategias de paginación
- Fácil extensión para nuevas características

## Migración desde Código Actual

### Antes (Código Duplicado)
```go
// En cada handler
limit := 10
if limitParam := c.QueryParam("limit"); limitParam != "" {
    if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
        limit = parsedLimit
    }
}

offset := 0
if offsetParam := c.QueryParam("offset"); offsetParam != "" {
    if parsedOffset, err := strconv.Atoi(offsetParam); err == nil {
        offset = parsedOffset
    }
}
```

### Después (Código Centralizado)
```go
// En cada handler
parser := pagination.NewDefaultParser()
params, err := parser.ParseFromQuery(
    c.QueryParam("limit"),
    c.QueryParam("offset"),
)
if err != nil {
    return responses.BadRequest(c, err.Error())
}
```

## Testing

La librería incluye tests completos que cubren:
- Parsing de parámetros válidos e inválidos
- Validación de límites y offsets
- Cálculo de metadata de paginación
- Utilidades de conversión página/offset
- Casos edge (datos vacíos, límites extremos)

## Integración con Echo Framework

```go
// Middleware opcional para inyectar parser
func PaginationMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            parser := pagination.NewDefaultParser()
            c.Set("pagination_parser", parser)
            return next(c)
        }
    }
}

// En handler
func (h *MyHandler) List(c echo.Context) error {
    parser := c.Get("pagination_parser").(*pagination.Parser)
    params, err := parser.ParseFromQuery(
        c.QueryParam("limit"),
        c.QueryParam("offset"),
    )
    // ...
}
```

## Próximas Mejoras

- [ ] Soporte para cursor-based pagination
- [ ] Integración con OpenAPI/Swagger
- [ ] Middleware automático para Echo
- [ ] Métricas de paginación
- [ ] Cache de metadata de paginación
