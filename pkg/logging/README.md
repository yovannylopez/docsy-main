# Logging

Una librería simple y eficiente para logging estructurado en aplicaciones Go. Basada en [Zap](https://github.com/uber-go/zap), proporciona una interfaz fácil de usar para logging con soporte para campos estructurados y diferentes niveles de log.

## Frontera con otros `pkg/`

- **`pkg/constants`**: el header HTTP de correlación es `constants.RequestIDHeader` (`X-Request-ID`). En logs estructurados la clave estable es **`logging.FieldKeyRequestID`** (`request_id`); úsala vía **`WithRequestID`**.
- **Stdlib `log`**: evitar `log.Print` / `log.Printf` en handlers y servicios del Core; usar **`logging.Info` / `Warn` / `Error` / `Debug`** para salida coherente con `Init` (JSON en producción, consola en desarrollo).
- **`Init`**: debe ejecutarse en el arranque (`cmd`, `bootstrap`) antes de workers o goroutines que logueen; hasta entonces el logger es no-op (`zap.NewNop()`).

## Características

- ✅ **Logging estructurado**: Soporte completo para campos estructurados
- ✅ **Múltiples niveles**: Debug, Info, Warn, Error
- ✅ **Configuración automática**: Modo desarrollo y producción
- ✅ **Campos personalizados**: Funciones helper para campos comunes
- ✅ **Request ID**: Soporte para tracking de requests
- ✅ **Basado en Zap**: Alto rendimiento y funcionalidad completa
- ✅ **Fácil integración**: API simple y directa

## Instalación

```bash
go get github.com/yovannylopez/docsy-main/pkg/logging
```

## Dependencias

```go
require go.uber.org/zap v1.27.0
```

## Uso Básico

### Importar la librería

```go
import "github.com/yovannylopez/docsy-main/pkg/logging"
```

### Comportamiento antes de `Init`

Hasta la primera llamada exitosa a `Init`, el logger global es un **`zap.NewNop()`**: las llamadas a `Info`, `Error`, etc. **no hacen panic** (no emiten salida). Tras `Init`, el logger real sustituye al no-op.

### Inicializar el logger

```go
package main

import (
    "log"
    "github.com/yovannylopez/docsy-main/pkg/logging"
)

func main() {
    // Inicializar en modo desarrollo
    if err := logging.Init(false); err != nil {
        log.Fatal("Error inicializando logger:", err)
    }
    
    // Inicializar en modo producción
    if err := logging.Init(true); err != nil {
        log.Fatal("Error inicializando logger:", err)
    }
}
```

### Logging básico

```go
package main

import (
    "github.com/yovannylopez/docsy-main/pkg/logging"
)

func main() {
    // Inicializar logger
    logging.Init(false)
    
    // Diferentes niveles de log
    logging.Debug("Mensaje de debug")
    logging.Info("Mensaje informativo")
    logging.Warn("Mensaje de advertencia")
    logging.Error("Mensaje de error")
}
```

### Logging con campos estructurados

```go
package main

import (
    "github.com/yovannylopez/docsy-main/pkg/logging"
    "go.uber.org/zap"
)

func main() {
    logging.Init(false)
    
    // Logging con campos personalizados
    logging.Info("Usuario autenticado",
        zap.String("user_id", "123"),
        zap.String("email", "user@example.com"),
        zap.String("ip", "192.168.1.1"),
    )
    
    // Logging con error
    err := someFunction()
    if err != nil {
        logging.Error("Error en función",
            zap.Error(err),
            zap.String("function", "someFunction"),
        )
    }
}
```

### Usar campos helper

```go
package main

import (
    "errors"
    "github.com/yovannylopez/docsy-main/pkg/logging"
    "go.uber.org/zap"
)

func main() {
    logging.Init(false)
    
    // Usar campos helper
    logging.Info("Operación completada",
        logging.WithRequestID("req-123"),
        zap.String("operation", "user_create"),
        zap.Int("duration_ms", 150),
    )
    
    // Logging con error usando helper
    err := errors.New("archivo no encontrado")
    logging.Error("Error procesando archivo",
        logging.WithError(err),
        zap.String("file", "config.json"),
    )
}
```

## API Reference

### Funciones de Inicialización

#### `Init(production bool) error`
Inicializa el logger con la configuración especificada.

**Parámetros:**
- `production`: Si es true, usa configuración de producción; si es false, usa configuración de desarrollo

**Retorna:** `error`

#### `Sync() error`
Vacía buffers del logger (p. ej. salida a stderr). Llamar **una vez al apagar** el proceso; es habitual ignorar el error: `_ = logging.Sync()`.

**Uso típico en `main`:**

```go
if err := logging.Init(isProd); err != nil {
    log.Fatal(err)
}
defer func() { _ = logging.Sync() }()
```

### Funciones de Logging

#### `Info(msg string, fields ...zap.Field)`
Registra un mensaje con nivel INFO.

#### `Error(msg string, fields ...zap.Field)`
Registra un mensaje con nivel ERROR.

#### `Warn(msg string, fields ...zap.Field)`
Registra un mensaje con nivel WARN.

#### `Debug(msg string, fields ...zap.Field)`
Registra un mensaje con nivel DEBUG.

### Funciones Helper

#### `FieldKeyRequestID`
Constante `"request_id"`: clave JSON en logs (no confundir con el nombre del header HTTP).

#### `WithRequestID(requestID string) zap.Field`
Crea un campo para el ID de request usando `FieldKeyRequestID`.

**Parámetros:**
- `requestID`: ID del request (string)

**Retorna:** `zap.Field`

#### `Logger() *zap.Logger`
Retorna la instancia del logger de Zap.

**Retorna:** `*zap.Logger`

#### `WithField(key string, value any) zap.Field`
Crea un campo estructurado genérico (equivalente práctico a `zap.Any`); combinable con `logging.Info` / `logging.Error` / etc.

#### `WithError(err error) zap.Field`
Crea un campo `error` con el error completo para Zap. Si `err` es **nil**, devuelve **`zap.Skip()`** (no se añade campo al log).

## Ejemplos de Uso

### En una API REST

```go
package main

import (
    "net/http"
    "time"
    "github.com/yovannylopez/docsy-main/pkg/logging"
    "go.uber.org/zap"
)

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Generar request ID
        requestID := generateRequestID()
        
        // Log request
        logging.Info("Request iniciado",
            logging.WithRequestID(requestID),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.String("user_agent", r.UserAgent()),
        )
        
        // Procesar request
        next.ServeHTTP(w, r)
        
        // Log response
        duration := time.Since(start)
        logging.Info("Request completado",
            logging.WithRequestID(requestID),
            zap.Duration("duration", duration),
            zap.String("status", "200"),
        )
    })
}

func handleUserCreate(w http.ResponseWriter, r *http.Request) {
    userID := "123"
    
    logging.Info("Usuario creado",
        logging.WithRequestID(r.Header.Get("X-Request-ID")),
        zap.String("user_id", userID),
        zap.String("email", "user@example.com"),
    )
    
    w.WriteHeader(http.StatusCreated)
}
```

### En un servicio de base de datos

```go
package main

import (
    "database/sql"
    "github.com/yovannylopez/docsy-main/pkg/logging"
    "go.uber.org/zap"
)

type UserService struct {
    db *sql.DB
}

func (s *UserService) CreateUser(user User) error {
    logging.Info("Creando usuario",
        zap.String("email", user.Email),
        zap.String("operation", "user_create"),
    )
    
    // Simular operación de base de datos
    if err := s.db.Ping(); err != nil {
        logging.Error("Error conectando a base de datos",
            logging.WithError(err),
            zap.String("operation", "user_create"),
        )
        return err
    }
    
    logging.Info("Usuario creado exitosamente",
        zap.String("user_id", "123"),
        zap.String("email", user.Email),
    )
    
    return nil
}
```

### En un microservicio

```go
package main

import (
    "context"
    "time"
    "github.com/yovannylopez/docsy-main/pkg/logging"
    "go.uber.org/zap"
)

type OrderService struct{}

func (s *OrderService) ProcessOrder(ctx context.Context, order Order) error {
    requestID := ctx.Value("request_id").(string)
    
    logging.Info("Procesando orden",
        logging.WithRequestID(requestID),
        zap.String("order_id", order.ID),
        zap.Float64("total", order.Total),
    )
    
    // Simular procesamiento
    time.Sleep(100 * time.Millisecond)
    
    if order.Total > 1000 {
        logging.Warn("Orden con valor alto",
            logging.WithRequestID(requestID),
            zap.String("order_id", order.ID),
            zap.Float64("total", order.Total),
        )
    }
    
    logging.Info("Orden procesada exitosamente",
        logging.WithRequestID(requestID),
        zap.String("order_id", order.ID),
        zap.Duration("processing_time", 100*time.Millisecond),
    )
    
    return nil
}
```

### Configuración en diferentes ambientes

```go
package main

import (
    "os"
    "github.com/yovannylopez/docsy-main/pkg/logging"
)

func initLogger() {
    // Determinar ambiente
    env := os.Getenv("ENVIRONMENT")
    isProduction := env == "production"
    
    // Inicializar logger
    if err := logging.Init(isProduction); err != nil {
        panic("Error inicializando logger: " + err.Error())
    }
    
    logging.Info("Logger inicializado",
        logging.WithField("environment", env),
        logging.WithField("production", isProduction),
    )
}

func main() {
    initLogger()
    
    logging.Info("Aplicación iniciada")
    // ... resto de la aplicación
}
```

## Configuración de Zap

La librería utiliza Zap como backend. En modo desarrollo, Zap proporciona logs legibles por humanos. En modo producción, proporciona logs estructurados en JSON para mejor rendimiento.

### Modo Desarrollo
```json
{
  "level": "info",
  "msg": "Usuario autenticado",
  "user_id": "123",
  "email": "user@example.com"
}
```

### Modo Producción
```json
{
  "level": "info",
  "ts": 1640995200.123,
  "msg": "Usuario autenticado",
  "user_id": "123",
  "email": "user@example.com"
}
```

## Mejores Prácticas

1. **Inicializa el logger al inicio**: Siempre inicializa el logger al comienzo de tu aplicación.

2. **Usa campos estructurados**: En lugar de concatenar strings, usa campos estructurados para mejor búsqueda y análisis.

3. **Incluye contexto relevante**: Agrega campos como `request_id`, `user_id`, `operation`, etc.

4. **Maneja errores apropiadamente**: Usa `logging.WithError()` para errores y `logging.Error()` para el nivel.

5. **Usa niveles apropiados**:
   - `Debug`: Información detallada para debugging
   - `Info`: Información general de la aplicación
   - `Warn`: Situaciones que requieren atención pero no son errores
   - `Error`: Errores que afectan la funcionalidad

6. **No loggees información sensible**: Evita loggear contraseñas, tokens, datos personales, etc.

7. **Sync al salir**: Usa `defer` con `logging.Sync()` en `main` para volcar buffers al terminar (ver API Reference).

## Integración con Middleware

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Agregar request ID al contexto
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        r = r.WithContext(ctx)
        
        // Log request
        logging.Info("HTTP Request",
            logging.WithRequestID(requestID),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.String("remote_addr", r.RemoteAddr),
        )
        
        // Procesar request
        next.ServeHTTP(w, r)
        
        // Log response
        duration := time.Since(start)
        logging.Info("HTTP Response",
            logging.WithRequestID(requestID),
            zap.Duration("duration", duration),
        )
    })
}
```

## Versión

- Go: 1.26.2+ (alineado al workspace raíz)
- Dependencias: `go.uber.org/zap v1.27.0`