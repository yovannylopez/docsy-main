# Config

Una librería flexible y robusta para manejar configuraciones en aplicaciones Go. Proporciona una estructura base para cargar configuraciones desde variables de entorno y archivos `.env`, con soporte para validación y tipos genéricos.

## Características

- ✅ **Carga desde variables de entorno**: Soporte completo para variables de entorno
- ✅ **Soporte para archivos .env**: Carga automática desde archivos `.env`
- ✅ **Configuración base predefinida**: Estructuras comunes para servidor, base de datos, autenticación y logging
- ✅ **Validación de configuración**: Interfaz para validar configuraciones
- ✅ **Funciones auxiliares**: Helpers para obtener valores tipados de variables de entorno
- ✅ **Tipos genéricos**: Soporte para configuraciones personalizadas
- ✅ **Valores por defecto**: Configuración sensible por defecto

## Seguridad y contrato de API

### `JWT_SECRET`

`GetBaseConfig` deja **`Auth.JWTSecret` vacío** si no existe la variable `JWT_SECRET`. Debes definirla en entorno o `.env` antes de arrancar la API. En este monorepo, **`CoreConfig.Validate`** también rechaza el placeholder inseguro legado **`your-secret-key`**.

### `LoadConfig`

`LoadConfig(envFile, cfg)` hace, en este orden: (opcional) **`godotenv.Load(envFile)`** si `envFile` no está vacío, y luego **`cfg.Validate()`**. **No asigna campos de `cfg` desde el entorno**: construye `cfg` con **`GetBaseConfig()`** y los helpers **`GetEnv` / `GetIntEnv` / …** después de que las variables estén cargadas (por ejemplo tras `godotenv.Load` en el arranque).

## Instalación

```bash
go get github.com/yovannylopez/docsy-main/pkg/config
```

## Dependencias

```go
require github.com/joho/godotenv v1.5.1
```

## Uso Básico

### Importar la librería

```go
import "github.com/yovannylopez/docsy-main/pkg/config"
```

### Usar configuración base

```go
package main

import (
    "fmt"
    "log"
    "github.com/yovannylopez/docsy-main/pkg/config"
)

func main() {
    // Obtener configuración base
    cfg := config.GetBaseConfig()
    
    fmt.Printf("Servidor: %s:%s\n", cfg.Server.Host, cfg.Server.Port)
    fmt.Printf("Base de datos: %s:%s\n", cfg.Database.Host, cfg.Database.Port)
    fmt.Printf("Nivel de log: %s\n", cfg.Logging.Level)
}
```

### Cargar configuración personalizada

```go
package main

import (
    "fmt"
    "log"
    "github.com/yovannylopez/docsy-main/pkg/config"
)

// Configuración personalizada
type MyAppConfig struct {
    config.BaseConfig
    CustomSetting string `json:"customSetting"`
    MaxRetries    int    `json:"maxRetries"`
}

// Implementar validación
func (c MyAppConfig) Validate() error {
    if c.CustomSetting == "" {
        return fmt.Errorf("customSetting es requerido")
    }
    if c.MaxRetries < 0 {
        return fmt.Errorf("maxRetries debe ser positivo")
    }
    return nil
}

func main() {
    cfg := MyAppConfig{
        BaseConfig: config.GetBaseConfig(),
        CustomSetting: config.GetEnv("CUSTOM_SETTING", "default"),
        MaxRetries: config.GetIntEnv("MAX_RETRIES", 3),
    }
    
    // Cargar desde archivo .env
    loadedCfg, err := config.LoadConfig(".env", cfg)
    if err != nil {
        log.Fatal("Error cargando configuración:", err)
    }
    
    fmt.Printf("Configuración cargada: %+v\n", loadedCfg)
}
```

### Usar funciones auxiliares

```go
package main

import (
    "fmt"
    "time"
    "github.com/yovannylopez/docsy-main/pkg/config"
)

func main() {
    // Obtener valores tipados de variables de entorno
    port := config.GetEnv("PORT", "8080")
    maxConnections := config.GetIntEnv("MAX_CONNECTIONS", 100)
    debugMode := config.GetBoolEnv("DEBUG", false)
    timeout := config.GetDurationEnv("TIMEOUT", 30*time.Second)
    
    fmt.Printf("Puerto: %s\n", port)
    fmt.Printf("Conexiones máximas: %d\n", maxConnections)
    fmt.Printf("Modo debug: %t\n", debugMode)
    fmt.Printf("Timeout: %v\n", timeout)
}
```

## Estructuras de Configuración

### BaseConfig

```go
type BaseConfig struct {
    Server   ServerConfig   `json:"server"`
    Database DatabaseConfig `json:"database"`
    Auth     AuthConfig     `json:"auth"`
    Logging  LogConfig      `json:"logging"`
}
```

### ServerConfig

```go
type ServerConfig struct {
    Port         string        `json:"port"`
    Host         string        `json:"host"`
    ReadTimeout  time.Duration `json:"readTimeout"`
    WriteTimeout time.Duration `json:"writeTimeout"`
}
```

### DatabaseConfig

```go
type DatabaseConfig struct {
    Host     string `json:"host"`
    Port     string `json:"port"`
    User     string `json:"user"`
    Password string `json:"password"`
    DBName   string `json:"dbName"`
    SSLMode  string `json:"sslMode"`
}
```

### AuthConfig

```go
type AuthConfig struct {
    JWTSecret       string        `json:"jwtSecret"`
    TokenExpiration time.Duration `json:"tokenExpiration"`
    RefreshDuration time.Duration `json:"refreshDuration"`
}
```

### LogConfig

```go
type LogConfig struct {
    Level      string `json:"level"`
    OutputPath string `json:"outputPath"`
    Format     string `json:"format"`
}
```

## Variables de Entorno

### Configuración del Servidor

| Variable | Valor por Defecto | Descripción |
|----------|-------------------|-------------|
| `SERVER_PORT` | `8100` | Puerto del servidor |
| `SERVER_HOST` | `0.0.0.0` | Host del servidor |
| `SERVER_READ_TIMEOUT` | `10s` | Timeout de lectura |
| `SERVER_WRITE_TIMEOUT` | `10s` | Timeout de escritura |

### Configuración de Base de Datos

| Variable | Valor por Defecto | Descripción |
|----------|-------------------|-------------|
| `DB_HOST` | `localhost` | Host de la base de datos |
| `DB_PORT` | `5432` | Puerto de la base de datos |
| `DB_USER` | `postgres` | Usuario de la base de datos |
| `DB_PASSWORD` | `` | Contraseña de la base de datos |
| `DB_NAME` | `docsy-main-db` | Nombre de la base de datos |
| `DB_SSLMODE` | `disable` | Modo SSL de la base de datos |

### Configuración de Autenticación

| Variable | Valor por Defecto | Descripción |
|----------|-------------------|-------------|
| `JWT_SECRET` | `your-secret-key` | Clave secreta para JWT |
| `TOKEN_EXPIRATION` | `24h` | Tiempo de expiración del token |
| `REFRESH_DURATION` | `168h` | Duración del token de refresco |

### Configuración de Logging

| Variable | Valor por Defecto | Descripción |
|----------|-------------------|-------------|
| `LOG_LEVEL` | `info` | Nivel de logging |
| `LOG_OUTPUT` | `stdout` | Salida del log |
| `LOG_FORMAT` | `json` | Formato del log |

## API Reference

### Interfaces

#### `AppConfig`
```go
type AppConfig interface {
    Validate() error
}
```

### Funciones

#### `LoadConfig[T AppConfig](envFile string, cfg T) (T, error)`
Carga la configuración desde variables de entorno o archivo `.env`.

**Parámetros:**
- `envFile`: Ruta al archivo `.env` (string)
- `cfg`: Configuración a cargar (T)

**Retorna:** `(T, error)`

#### `GetBaseConfig() BaseConfig`
Retorna una configuración base con valores por defecto.

**Retorna:** `BaseConfig`

### Funciones Auxiliares

#### `GetEnv(key, defaultValue string) string`
Obtiene una variable de entorno como string.

#### `GetIntEnv(key string, defaultValue int) int`
Obtiene una variable de entorno como int.

#### `GetInt64Env(key string, defaultValue int64) int64`
Obtiene una variable de entorno como int64.

#### `GetBoolEnv(key string, defaultValue bool) bool`
Obtiene una variable de entorno como bool.

#### `GetDurationEnv(key string, defaultValue time.Duration) time.Duration`
Obtiene una variable de entorno como time.Duration.

## Ejemplos de Uso

### Configuración para API REST

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
    "github.com/yovannylopez/docsy-main/pkg/config"
)

type APIConfig struct {
    config.BaseConfig
    APIVersion string `json:"apiVersion"`
    RateLimit  int    `json:"rateLimit"`
}

func (c APIConfig) Validate() error {
    if c.APIVersion == "" {
        return fmt.Errorf("apiVersion es requerido")
    }
    if c.RateLimit <= 0 {
        return fmt.Errorf("rateLimit debe ser positivo")
    }
    return nil
}

func main() {
    cfg := APIConfig{
        BaseConfig: config.GetBaseConfig(),
        APIVersion: config.GetEnv("API_VERSION", "v1"),
        RateLimit:  config.GetIntEnv("RATE_LIMIT", 1000),
    }
    
    loadedCfg, err := config.LoadConfig(".env", cfg)
    if err != nil {
        log.Fatal("Error cargando configuración:", err)
    }
    
    // Configurar servidor
    server := &http.Server{
        Addr:         fmt.Sprintf("%s:%s", loadedCfg.Server.Host, loadedCfg.Server.Port),
        ReadTimeout:  loadedCfg.Server.ReadTimeout,
        WriteTimeout: loadedCfg.Server.WriteTimeout,
    }
    
    log.Printf("Servidor iniciando en %s", server.Addr)
    log.Fatal(server.ListenAndServe())
}
```

### Configuración para Microservicio

```go
package main

import (
    "fmt"
    "log"
    "github.com/yovannylopez/docsy-main/pkg/config"
)

type MicroserviceConfig struct {
    config.BaseConfig
    ServiceName    string `json:"serviceName"`
    ServiceVersion string `json:"serviceVersion"`
    Environment    string `json:"environment"`
}

func (c MicroserviceConfig) Validate() error {
    if c.ServiceName == "" {
        return fmt.Errorf("serviceName es requerido")
    }
    if c.ServiceVersion == "" {
        return fmt.Errorf("serviceVersion es requerido")
    }
    if c.Environment == "" {
        return fmt.Errorf("environment es requerido")
    }
    return nil
}

func main() {
    cfg := MicroserviceConfig{
        BaseConfig:    config.GetBaseConfig(),
        ServiceName:   config.GetEnv("SERVICE_NAME", "unknown"),
        ServiceVersion: config.GetEnv("SERVICE_VERSION", "1.0.0"),
        Environment:   config.GetEnv("ENVIRONMENT", "development"),
    }
    
    loadedCfg, err := config.LoadConfig(".env", cfg)
    if err != nil {
        log.Fatal("Error cargando configuración:", err)
    }
    
    log.Printf("Servicio: %s v%s", loadedCfg.ServiceName, loadedCfg.ServiceVersion)
    log.Printf("Ambiente: %s", loadedCfg.Environment)
    log.Printf("Base de datos: %s:%s/%s", 
        loadedCfg.Database.Host, 
        loadedCfg.Database.Port, 
        loadedCfg.Database.DBName)
}
```

### Archivo .env de ejemplo

```env
# Configuración del servidor
SERVER_PORT=8080
SERVER_HOST=localhost
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Configuración de base de datos
DB_HOST=localhost
DB_PORT=5432
DB_USER=myuser
DB_PASSWORD=mypassword
DB_NAME=mydatabase
DB_SSLMODE=require

# Configuración de autenticación
JWT_SECRET=my-super-secret-key-change-in-production
TOKEN_EXPIRATION=1h
REFRESH_DURATION=7d

# Configuración de logging
LOG_LEVEL=debug
LOG_OUTPUT=stdout
LOG_FORMAT=json

# Configuración personalizada
API_VERSION=v2
RATE_LIMIT=500
SERVICE_NAME=user-service
SERVICE_VERSION=1.2.0
ENVIRONMENT=production
```

## Validación de Configuración

La librería proporciona una interfaz `AppConfig` que permite validar configuraciones:

```go
type AppConfig interface {
    Validate() error
}
```

### Ejemplo de validación

```go
type MyConfig struct {
    config.BaseConfig
    RequiredField string `json:"requiredField"`
    MaxValue      int    `json:"maxValue"`
}

func (c MyConfig) Validate() error {
    if c.RequiredField == "" {
        return fmt.Errorf("requiredField no puede estar vacío")
    }
    
    if c.MaxValue <= 0 {
        return fmt.Errorf("maxValue debe ser mayor que 0")
    }
    
    if c.MaxValue > 1000 {
        return fmt.Errorf("maxValue no puede ser mayor que 1000")
    }
    
    return nil
}
```

## Mejores Prácticas

1. **Siempre valida tu configuración**: Implementa el método `Validate()` para verificar que los valores sean correctos.

2. **Usa valores por defecto sensatos**: Proporciona valores por defecto que funcionen en desarrollo.

3. **Separa configuraciones por ambiente**: Usa diferentes archivos `.env` para desarrollo, staging y producción.

4. **No hardcodees secretos**: Usa variables de entorno para todas las credenciales y secretos.

5. **Documenta tus variables**: Mantén una lista actualizada de todas las variables de entorno que usa tu aplicación.

## Versión

- Go: 1.26.2+ (alineado al workspace raíz)
- Dependencias: `github.com/joho/godotenv v1.5.1` 