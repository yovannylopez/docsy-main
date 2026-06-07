# Circuit Breaker para PostgreSQL

## Descripción

El Circuit Breaker es un patrón de diseño que protege la aplicación contra cascadas de fallos cuando la base de datos PostgreSQL experimenta problemas. Implementa tres estados (closed, open, half-open) para gestionar automáticamente la disponibilidad del servicio.

## Características

### Estados del Circuit Breaker

1. **Closed (Cerrado)**: Estado normal, todas las peticiones pasan a la base de datos
2. **Open (Abierto)**: El circuito está abierto debido a fallos consecutivos, las peticiones son rechazadas inmediatamente
3. **Half-Open (Semi-abierto)**: Estado de prueba, permite un número limitado de peticiones para verificar si el servicio se ha recuperado

### Configuración

El Circuit Breaker se configura mediante la estructura `CircuitBreakerConfig`:

```go
type CircuitBreakerConfig struct {
    // MaxRequests: Número máximo de peticiones permitidas en estado half-open
    MaxRequests uint32
    
    // Interval: Período de tiempo para resetear el contador de fallos en estado closed
    Interval time.Duration
    
    // Timeout: Período de tiempo que el circuito permanece open antes de pasar a half-open
    Timeout time.Duration
    
    // ConsecutiveFailures: Número de fallos consecutivos antes de abrir el circuito
    ConsecutiveFailures uint32
    
    // Enabled: Indica si el Circuit Breaker está habilitado
    Enabled bool
}
```

### Configuraciones Predefinidas

#### Desarrollo
```go
cfg := databases.DevelopmentCircuitBreakerConfig()
// MaxRequests: 2
// Interval: 30 segundos
// Timeout: 15 segundos
// ConsecutiveFailures: 3
// Enabled: true
```

#### Por Defecto
```go
cfg := databases.DefaultCircuitBreakerConfig()
// MaxRequests: 3
// Interval: 60 segundos
// Timeout: 30 segundos
// ConsecutiveFailures: 5
// Enabled: true
```

#### Producción
```go
cfg := databases.ProductionCircuitBreakerConfig()
// MaxRequests: 5
// Interval: 120 segundos
// Timeout: 60 segundos
// ConsecutiveFailures: 5
// Enabled: true
```

## Uso

### Inicialización Automática

El Circuit Breaker se inicializa automáticamente cuando se crea una conexión a la base de datos:

```go
db, cbWrapper, err := databases.NewConnectionWithCircuitBreaker(ctx, databases.Config{
    DatabaseURL: "postgresql://...",
    CircuitBreaker: databases.ProductionCircuitBreakerConfig(),
})
```

### Ejecución de Operaciones

Para ejecutar operaciones protegidas por el Circuit Breaker:

```go
err := cbWrapper.Execute(ctx, "operation_name", func() error {
    // Tu operación de base de datos aquí
    return db.QueryContext(ctx, "SELECT * FROM users")
})

if err != nil {
    if errors.IsServiceUnavailableError(err) {
        // El Circuit Breaker está abierto
        // Retornar HTTP 503 Service Unavailable
        return c.JSON(http.StatusServiceUnavailable, map[string]any{
            "error": "Database service temporarily unavailable",
        })
    }
    // Otro tipo de error
    return err
}
```

### Verificación de Estado

```go
// Obtener el estado actual
state := cbWrapper.GetState() // "closed", "open", "half-open", o "disabled"

// Obtener estadísticas
stats := cbWrapper.GetCounts()
// stats contiene:
// - enabled: bool
// - state: string
// - requests: uint32
// - total_successes: uint32
// - total_failures: uint32
// - consecutive_successes: uint32
// - consecutive_failures: uint32
```

## Integración con Health Check

El Circuit Breaker se integra automáticamente con el endpoint de health check:

```bash
curl http://localhost:8080/health
```

Respuesta:
```json
{
  "status": "ok",
  "timestamp": "2026-03-14T21:00:00Z",
  "services": {
    "database": {
      "status": "healthy",
      "error": "",
      "stats": {
        "max_open_connections": 50,
        "open_connections": 5,
        "in_use": 2,
        "idle": 3
      },
      "circuit_breaker": {
        "enabled": true,
        "state": "closed",
        "requests": 1234,
        "total_successes": 1230,
        "total_failures": 4,
        "consecutive_successes": 100,
        "consecutive_failures": 0
      }
    }
  }
}
```

## Manejo de Errores

### Error de Servicio No Disponible

Cuando el Circuit Breaker está abierto, se retorna un error tipado `ServiceUnavailableError`:

```go
import "github.com/yovannylopez/docsy-main/pkg/errors"

if errors.IsServiceUnavailableError(err) {
    // El servicio no está disponible temporalmente
    // Retornar HTTP 503
    return c.JSON(http.StatusServiceUnavailable, map[string]any{
        "error": "Service temporarily unavailable",
        "message": "Please try again later",
    })
}
```

### Logging

El Circuit Breaker registra automáticamente los siguientes eventos:

1. **Inicialización**: Cuando se crea el Circuit Breaker
2. **Cambios de Estado**: Cuando el circuito cambia de closed → open → half-open → closed
3. **Fallos de Operación**: Cuando una operación de base de datos falla
4. **Circuito Abierto**: Cuando se rechaza una petición porque el circuito está abierto

Ejemplo de logs:
```
INFO  Circuit Breaker initialized  max_requests=5 interval=2m0s timeout=1m0s consecutive_failures_threshold=5
ERROR Circuit Breaker is opening due to consecutive failures  consecutive_failures=5 threshold=5 total_failures=5 total_requests=5 failure_ratio=1
INFO  Circuit Breaker state changed  circuit=PostgreSQL from=closed to=open
WARN  Circuit Breaker is open, rejecting database operation  operation=query_users state=open
INFO  Circuit Breaker state changed  circuit=PostgreSQL from=open to=half-open
INFO  Circuit Breaker state changed  circuit=PostgreSQL from=half-open to=closed
```

## Desactivación

Para desactivar el Circuit Breaker (no recomendado en producción):

```go
cfg := databases.Config{
    // ... otras configuraciones
    CircuitBreaker: databases.CircuitBreakerConfig{
        Enabled: false,
    },
}
```

## Monitoreo

### Métricas Clave

1. **Estado del Circuito**: Monitorear cambios frecuentes de estado
2. **Tasa de Fallos**: `total_failures / requests`
3. **Fallos Consecutivos**: Indica problemas persistentes
4. **Tiempo en Estado Open**: Duración de las interrupciones

### Alertas Recomendadas

1. **Circuito Abierto**: Alerta crítica cuando el circuito se abre
2. **Tasa de Fallos Alta**: Alerta cuando `failure_ratio > 0.1` (10%)
3. **Fallos Consecutivos**: Alerta cuando se acerca al umbral configurado

## Mejores Prácticas

1. **Configuración por Ambiente**:
   - Desarrollo: Umbrales más bajos para detectar problemas rápidamente
   - Producción: Umbrales más altos para evitar falsos positivos

2. **Timeouts Apropiados**:
   - `Timeout` debe ser suficiente para que la base de datos se recupere
   - Típicamente 30-60 segundos en producción

3. **Monitoreo Proactivo**:
   - Configurar alertas para cambios de estado
   - Revisar logs cuando el circuito se abre

4. **Manejo de Errores**:
   - Siempre verificar `IsServiceUnavailableError()`
   - Retornar HTTP 503 cuando el circuito está abierto
   - Proporcionar mensajes claros al usuario

5. **Testing**:
   - Probar el comportamiento con fallos simulados
   - Verificar que el circuito se abre correctamente
   - Validar la recuperación automática

## Troubleshooting

### El circuito se abre frecuentemente

**Causas posibles**:
- Base de datos sobrecargada
- Consultas lentas o mal optimizadas
- Problemas de red
- Configuración de timeouts muy baja

**Soluciones**:
- Revisar logs de la base de datos
- Optimizar consultas lentas
- Aumentar recursos de la base de datos
- Ajustar configuración del Circuit Breaker

### El circuito no se abre cuando debería

**Causas posibles**:
- Umbral de `ConsecutiveFailures` muy alto
- Errores no están siendo propagados correctamente

**Soluciones**:
- Reducir `ConsecutiveFailures`
- Verificar que los errores se propagan correctamente
- Revisar logs para confirmar que los fallos se están registrando

### Recuperación lenta después de que el circuito se abre

**Causas posibles**:
- `Timeout` muy largo
- `MaxRequests` muy bajo en estado half-open

**Soluciones**:
- Reducir `Timeout` para intentar reconectar más rápido
- Aumentar `MaxRequests` para permitir más peticiones de prueba

## Referencias

- [Circuit Breaker Pattern - Martin Fowler](https://martinfowler.com/bliki/CircuitBreaker.html)
- [github.com/sony/gobreaker](https://github.com/sony/gobreaker)
- [Resilience Patterns](https://docs.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker)

## Changelog

### v1.0.0 (2026-03-14)
- Implementación inicial del Circuit Breaker
- Integración con PostgreSQL
- Soporte para configuraciones por ambiente
- Integración con health check endpoint
- Tests unitarios completos
