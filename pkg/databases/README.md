# Librería Database - Control Eficiente de Conexiones PostgreSQL

Esta librería proporciona un control eficiente y optimizado de conexiones a PostgreSQL con configuraciones automáticas por ambiente, monitoreo de estadísticas y health checks integrados.

## 🎯 **Características Principales**

- ✅ **Pool de conexiones optimizado** con configuraciones por ambiente
- ✅ **Circuit Breaker integrado** para protección contra cascadas de fallos
- ✅ **Retry automático** con delays configurables
- ✅ **Timeouts configurables** para conexión y consultas
- ✅ **Monitoreo de estadísticas** del pool de conexiones y Circuit Breaker
- ✅ **Health checks integrados** con timeouts
- ✅ **Configuraciones predefinidas** para desarrollo y producción
- ✅ **Logging detallado** de conexiones y errores
- ✅ **Graceful shutdown** con estadísticas finales

## 📦 **Configuraciones Disponibles**

### **DevelopmentConfig()**
Configuración optimizada para desarrollo con recursos limitados:

```go
MaxOpenConns:    10
MaxIdleConns:    5
ConnMaxLifetime: 2 * time.Minute
ConnMaxIdleTime: 30 * time.Second
ConnectTimeout:  5 * time.Second
QueryTimeout:    15 * time.Second
MaxRetries:      2
RetryDelay:      500 * time.Millisecond
```

### **ProductionConfig()**
Configuración optimizada para producción con alta concurrencia:

```go
MaxOpenConns:    50
MaxIdleConns:    25
ConnMaxLifetime: 10 * time.Minute
ConnMaxIdleTime: 5 * time.Minute
ConnectTimeout:  15 * time.Second
QueryTimeout:    60 * time.Second
MaxRetries:      5
RetryDelay:      2 * time.Second
```

### **DefaultConfig()**
Configuración balanceada para uso general:

```go
MaxOpenConns:    25
MaxIdleConns:    10
ConnMaxLifetime: 5 * time.Minute
ConnMaxIdleTime: 1 * time.Minute
ConnectTimeout:  10 * time.Second
QueryTimeout:    30 * time.Second
MaxRetries:      3
RetryDelay:      1 * time.Second
```

## 🚀 **Uso Básico**

### **Configuración Automática por Ambiente**

```go
import "github.com/yovannylopez/docsy-main/pkg/databases"

// La configuración se detecta automáticamente según ENVIRONMENT
config := database.DevelopmentConfig() // o ProductionConfig()
config.Host = "localhost"
config.Port = "5432"
config.User = "postgres"
config.Password = "password"
config.DBName = "myapp"
config.SSLMode = "disable"

db, err := database.NewConnection(config)
if err != nil {
    log.Fatal(err)
}
defer database.Close(db)
```

### **Configuración Personalizada**

```go
config := database.Config{
    Host:             "prod-db.example.com",
    Port:             "5432",
    User:             "app_user",
    Password:         "secure_password",
    DBName:           "myapp_prod",
    SSLMode:          "require",
    MaxOpenConns:     100,
    MaxIdleConns:     50,
    ConnMaxLifetime:  15 * time.Minute,
    ConnMaxIdleTime:  10 * time.Minute,
    ConnectTimeout:   20 * time.Second,
    QueryTimeout:     120 * time.Second,
    MaxRetries:       10,
    RetryDelay:       5 * time.Second,
}

db, err := database.NewConnection(config)
```

## 📊 **Monitoreo y Estadísticas**

### **Obtener Estadísticas del Pool**

```go
stats := database.GetConnectionStats(db)
fmt.Printf("Conexiones abiertas: %d\n", stats["open_connections"])
fmt.Printf("Conexiones en uso: %d\n", stats["in_use"])
fmt.Printf("Conexiones idle: %d\n", stats["idle"])
```

### **Logging de Estadísticas**

```go
// Log automático de estadísticas
database.LogConnectionStats(db)
```

### **Health Check**

```go
// Verificar salud de la conexión
if err := database.HealthCheck(db, 5*time.Second); err != nil {
    log.Printf("Health check failed: %v", err)
} else {
    log.Println("Database is healthy")
}
```

## 🔧 **Integración en Core**

### **Configuración Automática**

```go
// En config/config.go
func getDBPoolConfig() DBPoolConfig {
    environment := config.GetEnv("ENVIRONMENT", "development")
    
    switch environment {
    case "production":
        prodConfig := database.ProductionConfig()
        return DBPoolConfig{
            MaxOpenConns:    prodConfig.MaxOpenConns,
            MaxIdleConns:    prodConfig.MaxIdleConns,
            ConnMaxLifetime: prodConfig.ConnMaxLifetime,
            ConnMaxIdleTime: prodConfig.ConnMaxIdleTime,
            ConnectTimeout:  prodConfig.ConnectTimeout,
            QueryTimeout:    prodConfig.QueryTimeout,
            MaxRetries:      prodConfig.MaxRetries,
            RetryDelay:      prodConfig.RetryDelay,
        }
    case "development":
        devConfig := database.DevelopmentConfig()
        return DBPoolConfig{
            MaxOpenConns:    devConfig.MaxOpenConns,
            MaxIdleConns:    devConfig.MaxIdleConns,
            ConnMaxLifetime: devConfig.ConnMaxLifetime,
            ConnMaxIdleTime: devConfig.ConnMaxIdleTime,
            ConnectTimeout:  devConfig.ConnectTimeout,
            QueryTimeout:    devConfig.QueryTimeout,
            MaxRetries:      devConfig.MaxRetries,
            RetryDelay:      devConfig.RetryDelay,
        }
    default:
        defaultConfig := database.DefaultConfig()
        return DBPoolConfig{
            MaxOpenConns:    defaultConfig.MaxOpenConns,
            MaxIdleConns:    defaultConfig.MaxIdleConns,
            ConnMaxLifetime: defaultConfig.ConnMaxLifetime,
            ConnMaxIdleTime: defaultConfig.ConnMaxIdleTime,
            ConnectTimeout:  defaultConfig.ConnectTimeout,
            QueryTimeout:    defaultConfig.QueryTimeout,
            MaxRetries:      defaultConfig.MaxRetries,
            RetryDelay:      defaultConfig.RetryDelay,
        }
    }
}
```

### **Container con Pool Optimizado**

```go
// En container/container.go
func NewContainer(cfg *config.CoreConfig) (*Container, error) {
    // Configurar conexión de base de datos con pool optimizado
    dbConfig := database.Config{
        Host:            cfg.Database.Host,
        Port:            cfg.Database.Port,
        User:            cfg.Database.User,
        Password:        cfg.Database.Password,
        DBName:          cfg.Database.DBName,
        SSLMode:         cfg.Database.SSLMode,
        MaxOpenConns:    cfg.DBPool.MaxOpenConns,
        MaxIdleConns:    cfg.DBPool.MaxIdleConns,
        ConnMaxLifetime: cfg.DBPool.ConnMaxLifetime,
        ConnMaxIdleTime: cfg.DBPool.ConnMaxIdleTime,
        ConnectTimeout:  cfg.DBPool.ConnectTimeout,
        QueryTimeout:    cfg.DBPool.QueryTimeout,
        MaxRetries:      cfg.DBPool.MaxRetries,
        RetryDelay:      cfg.DBPool.RetryDelay,
    }

    db, err := database.NewConnection(dbConfig)
    if err != nil {
        return nil, err
    }

    // ... resto de la configuración
}
```

## 🏥 **Health Checks Integrados**

### **Endpoint de Health Check**

```go
// En handlers/health_handler.go
func (h *HealthHandler) HealthCheck(c echo.Context) error {
    // Verificar salud de la base de datos
    dbHealth := "healthy"
    dbError := ""
    
    if err := database.HealthCheck(h.db, 5*time.Second); err != nil {
        dbHealth = "unhealthy"
        dbError = err.Error()
        logging.Error("Database health check failed", zap.Error(err))
    }

    // Obtener estadísticas del pool de conexiones
    dbStats := database.GetConnectionStats(h.db)

    response := map[string]any{
        "status":    "ok",
        "timestamp": time.Now().UTC(),
        "services": map[string]any{
            "database": map[string]any{
                "status": dbHealth,
                "error":  dbError,
                "stats":  dbStats,
            },
            "api": map[string]any{
                "status": "healthy",
            },
        },
    }

    statusCode := http.StatusOK
    if dbHealth == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }

    return c.JSON(statusCode, response)
}
```

## 📈 **Estadísticas Disponibles**

La función `GetConnectionStats()` retorna las siguientes métricas:

- **max_open_connections**: Número máximo de conexiones abiertas
- **open_connections**: Número actual de conexiones abiertas
- **in_use**: Conexiones actualmente en uso
- **idle**: Conexiones idle disponibles
- **wait_count**: Número de veces que se esperó por una conexión
- **wait_duration**: Tiempo total esperando por conexiones
- **max_idle_closed**: Conexiones idle cerradas por límite
- **max_lifetime_closed**: Conexiones cerradas por lifetime

## 🔄 **Retry y Resiliencia**

### **Configuración de Retry**

```go
config := database.Config{
    // ... configuración básica
    MaxRetries: 5,           // Número máximo de intentos
    RetryDelay: 2 * time.Second, // Delay entre intentos
}
```

### **Comportamiento de Retry**

1. **Intento inicial**: Conexión directa
2. **Intentos fallidos**: Log de warning con número de intento
3. **Delay progresivo**: Espera configurada entre intentos
4. **Timeout de conexión**: Timeout individual por intento
5. **Error final**: Log de error con todos los intentos

## 🎛️ **Configuración por Ambiente**

### **Variables de Entorno**

```bash
# Ambiente (development, production)
ENVIRONMENT=development

# Configuración de base de datos
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=myapp
DB_SSL_MODE=disable
```

### **Configuración Automática**

La librería detecta automáticamente el ambiente y aplica la configuración correspondiente:

- **development**: Recursos limitados, timeouts cortos
- **production**: Alta concurrencia, timeouts largos
- **default**: Configuración balanceada

## 🧪 **Testing**

### **Subpaquete `migrate` (integración con PostgreSQL)**

Hay un test de integración con [Testcontainers](https://golang.testcontainers.org/) que levanta `postgres:16-alpine`, aplica migraciones de fixture y comprueba `Up`, `Version`, `Down` e idempotencia. **Requiere Docker**.

```bash
cd pkg/databases
go test -tags=integration -cover ./migrate/...
```

Los archivos de migración de prueba están en `migrate/testdata/fixture_migrations/` (no deben confundirse con `migrations/core` del repositorio principal).

### **Health Check en Tests**

```go
func TestDatabaseHealth(t *testing.T) {
    config := database.DevelopmentConfig()
    config.Host = "localhost"
    config.Port = "5432"
    config.User = "test_user"
    config.Password = "test_password"
    config.DBName = "test_db"
    config.SSLMode = "disable"

    db, err := database.NewConnection(config)
    if err != nil {
        t.Fatalf("Failed to connect: %v", err)
    }
    defer database.Close(db)

    // Test health check
    if err := database.HealthCheck(db, 5*time.Second); err != nil {
        t.Errorf("Health check failed: %v", err)
    }

    // Test statistics
    stats := database.GetConnectionStats(db)
    if stats["open_connections"].(int) == 0 {
        t.Error("Expected open connections")
    }
}
```

## 📋 **Mejores Prácticas**

### **1. Configuración por Ambiente**
```go
// Usar configuraciones predefinidas
config := database.DevelopmentConfig() // o ProductionConfig()
// Personalizar solo lo necesario
config.Host = "my-db-host"
```

### **2. Monitoreo Regular**
```go
// Log estadísticas periódicamente
ticker := time.NewTicker(5 * time.Minute)
defer ticker.Stop()

for range ticker.C {
    database.LogConnectionStats(db)
}
```

### **3. Health Checks**
```go
// Verificar salud antes de operaciones críticas
if err := database.HealthCheck(db, 3*time.Second); err != nil {
    // Manejar error de salud
    return err
}
```

### **4. Graceful Shutdown**
```go
// Cerrar conexión con estadísticas
defer func() {
    if err := database.Close(db); err != nil {
        log.Printf("Error closing database: %v", err)
    }
}()
```

## 🚨 **Troubleshooting**

### **Problemas Comunes**

1. **Conexión lenta**: Aumentar `ConnectTimeout`
2. **Muchas conexiones**: Ajustar `MaxOpenConns` y `MaxIdleConns`
3. **Timeouts frecuentes**: Incrementar `QueryTimeout`
4. **Conexiones perdidas**: Verificar `ConnMaxLifetime`

### **Logs de Debug**

```go
// Habilitar logs detallados
logging.SetLevel("debug")

// Los logs incluyen:
// - Configuración de conexión
// - Intentos de retry
// - Estadísticas del pool
// - Errores de health check
```

## 🔌 **Circuit Breaker**

### **Protección contra Cascadas de Fallos**

El Circuit Breaker implementado protege la aplicación contra cascadas de fallos cuando la base de datos experimenta problemas. Ver documentación completa en [CIRCUIT_BREAKER.md](../../docs/CIRCUIT_BREAKER.md).

### **Uso Básico**

```go
// Crear conexión con Circuit Breaker
db, cbWrapper, err := databases.NewConnectionWithCircuitBreaker(ctx, databases.Config{
    Host:     "localhost",
    Port:     "5432",
    User:     "postgres",
    Password: "password",
    DBName:   "myapp",
    SSLMode:  "disable",
    CircuitBreaker: databases.ProductionCircuitBreakerConfig(),
})

// Ejecutar operaciones protegidas
err = cbWrapper.Execute(ctx, "query_users", func() error {
    return db.QueryContext(ctx, "SELECT * FROM users")
})

if err != nil {
    if errors.IsServiceUnavailableError(err) {
        // Circuit Breaker está abierto
        return c.JSON(http.StatusServiceUnavailable, map[string]any{
            "error": "Database service temporarily unavailable",
        })
    }
    return err
}
```

### **Configuraciones del Circuit Breaker**

#### **Desarrollo**
```go
CircuitBreaker: databases.DevelopmentCircuitBreakerConfig()
// MaxRequests: 2
// Interval: 30 segundos
// Timeout: 15 segundos
// ConsecutiveFailures: 3
```

#### **Producción**
```go
CircuitBreaker: databases.ProductionCircuitBreakerConfig()
// MaxRequests: 5
// Interval: 120 segundos
// Timeout: 60 segundos
// ConsecutiveFailures: 5
```

### **Monitoreo del Circuit Breaker**

```go
// Obtener estado actual
state := cbWrapper.GetState() // "closed", "open", "half-open", "disabled"

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

// Log de estadísticas
cbWrapper.LogCircuitBreakerStats()
```

## 📚 **Próximos Pasos**

- [ ] Agregar soporte para múltiples bases de datos
- [ ] Implementar métricas Prometheus
- [ ] Agregar configuración de SSL avanzada
- [ ] Implementar connection pooling distribuido
- [ ] Agregar soporte para read replicas
- [x] Implementar circuit breaker pattern ✅

## 🤝 **Contribución**

Para contribuir a esta librería:

1. Fork el repositorio
2. Crea una rama para tu feature
3. Implementa los cambios
4. Agrega tests
5. Envía un pull request

## 📄 **Licencia**

Este proyecto está bajo la licencia MIT. Ver el archivo LICENSE para más detalles. 