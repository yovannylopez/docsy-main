# Paquete de Errores - Docsy

Este paquete proporciona un sistema de manejo de errores personalizado para la aplicación docsy-main, combinando las mejores prácticas de `github.com/pkg/errors` con errores específicos del dominio.

## Relación con `pkg/responses`

En el monorepo existe otro tipo **`AppError`** en [`pkg/responses`](../responses/error_types.go), orientado a **respuestas HTTP** (código de estado, serialización JSON para la API, integración con Echo). **No es el mismo tipo** que el `AppError` de este paquete.

| Paquete | Uso recomendado |
|---------|-----------------|
| **`pkg/errors` (este)** | Errores en capas de infraestructura / servicios compartidos (DB, circuit breaker, repositorios) donde conviene `ErrorType`, `Resource`, `Operation` y cadena con `Wrap`. |
| **`pkg/responses`** | Errores que el **middleware HTTP** o los handlers traducen directamente a JSON de cliente (códigos HTTP, mensajes de API). |

Importar este paquete con **alias** (p. ej. `apperrors`) para no colisionar con `errors` de la biblioteca estándar.

## Comportamiento destacado (v1.1)

- **`Wrap`**: si el error ya es `*AppError`, devuelve un **nuevo** valor con el mensaje prefijado y `Unwrap()` apuntando al original; **no muta** el puntero recibido.
- **`Cause`**: delega en `github.com/pkg/errors.Cause` **sin** añadir `fmt.Errorf` extra (preserva la raíz para `errors.Is` / `errors.As`).
- **`IsValidationError`, `IsNotFoundError`, …**: usan **`As` en la cadena** de errores, no solo type assertion directa.
- **`DatabaseError`**: el campo `Message` visible en `Error()` usa el prefijo en español `fallo en base de datos: <operation>`; `operation` debe ser un identificador estable (snake_case) para logs y métricas.

## 🎯 **Características**

### **✅ Errores Tipados**
- Errores específicos por tipo de problema
- Códigos de error estandarizados
- Mensajes descriptivos

### **✅ Contexto Enriquecido**
- Información adicional (operación, recurso, detalles)
- Mensajes para usuarios finales
- Stack trace preservado

### **✅ Compatibilidad**
- Compatible con `github.com/pkg/errors`
- Implementa interfaces estándar de Go
- Fácil migración desde errores básicos

## 📋 **Tipos de Error Disponibles**

| Tipo | Descripción | Uso |
|------|-------------|-----|
| `ValidationError` | Errores de validación | Datos de entrada inválidos |
| `NotFoundError` | Recurso no encontrado | Entidades que no existen |
| `UnauthorizedError` | No autorizado | Autenticación requerida |
| `ForbiddenError` | Acceso prohibido | Permisos insuficientes |
| `ConflictError` | Conflicto de recursos | Recursos duplicados |
| `DatabaseError` | Error de base de datos | Operaciones DB fallidas |
| `ExternalServiceError` | Error de servicio externo | APIs externas |
| `InternalError` | Error interno | Errores del sistema |

## 🚀 **Uso Básico**

### **Crear Errores**
```go
import "github.com/yovannylopez/docsy-main/pkg/errors"

// Error de validación
err := errors.ValidationError("INVALID_EMAIL", "Email inválido")

// Error de recurso no encontrado
err := errors.NotFoundError("user", "user123")

// Error de base de datos
err := errors.DatabaseError("create_user", dbErr)
```

### **Envolver Errores**
```go
// Envolver con contexto
err := errors.Wrap(dbErr, "failed to create user")

// Envolver con formato
err := errors.Wrapf(dbErr, "failed to create user with email %s", email)
```

### **Verificar Tipos**
```go
if errors.IsValidationError(err) {
    // Manejar error de validación
}

if errors.IsNotFoundError(err) {
    // Manejar error de no encontrado
}

if errors.IsDatabaseError(err) {
    // Manejar error de base de datos
}
```

## 🔧 **Funciones Principales**

### **Creación de Errores**
- `New(errorType, code, message)` - Error básico
- `ValidationError(code, message)` - Error de validación
- `NotFoundError(resource, identifier)` - Recurso no encontrado
- `UnauthorizedError(message)` - No autorizado
- `ForbiddenError(message)` - Acceso prohibido
- `ConflictError(resource, identifier)` - Conflicto
- `DatabaseError(operation, err)` - Error de DB
- `ExternalServiceError(service, operation, err)` - Servicio externo
- `InternalError(message, err)` - Error interno

### **Manejo de Errores**
- `Wrap(err, message)` - Envolver con contexto
- `Wrapf(err, format, args...)` - Envolver con formato
- `WithStack(err)` - Agregar stack trace
- `Cause(err)` - Obtener error original
- `Is(err, target)` - Verificar tipo
- `As(err, target)` - Extraer tipo específico

### **Verificación de Tipos**
- `IsValidationError(err)` - ¿Es error de validación?
- `IsNotFoundError(err)` - ¿Es error de no encontrado?
- `IsUnauthorizedError(err)` / `IsForbiddenError(err)` - Autorización
- `IsConflictError(err)` - Conflicto
- `IsDatabaseError(err)` - ¿Es error de base de datos?
- `IsInternalError(err)` / `IsExternalServiceError(err)` - Interno / externo
- `IsServiceUnavailableError(err)` - Servicio no disponible
- `GetAppError(err)` - Extraer AppError

### **Métodos de AppError**
- `WithDetails(details)` - Agregar detalles
- `WithUserMessage(message)` - Mensaje para usuario
- `WithOperation(operation)` - Agregar operación
- `WithResource(resource)` - Agregar recurso

## 📝 **Ejemplos de Uso**

### **Error de Validación**
```go
func ValidateEmail(email string) error {
    if !isValidEmail(email) {
        return errors.ValidationError("INVALID_EMAIL", "Email inválido")
            .WithDetails("Formato incorrecto")
            .WithUserMessage("Por favor ingresa un email válido")
    }
    return nil
}
```

### **Error de Base de Datos**
```go
func CreateUser(user *User) error {
    if err := db.Create(user); err != nil {
        return errors.DatabaseError("create_user", err)
            .WithOperation("INSERT")
            .WithResource("users")
    }
    return nil
}
```

### **Error de Recurso No Encontrado**
```go
func GetUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        return nil, errors.Wrap(err, "failed to find user")
    }
    if user == nil {
        return nil, errors.NotFoundError("user", id)
    }
    return user, nil
}
```

### **Manejo de Errores**
```go
func HandleUserRequest(userID string) error {
    user, err := GetUser(userID)
    if err != nil {
        if errors.IsNotFoundError(err) {
            return errors.Wrap(err, "user not found in request")
        }
        if errors.IsDatabaseError(err) {
            return errors.Wrap(err, "database error in user request")
        }
        return errors.Wrap(err, "unexpected error in user request")
    }
    // Procesar usuario...
    return nil
}
```

## 🔄 **Migración desde Errores Básicos**

### **Antes**
```go
return fmt.Errorf("user not found: %s", userID)
```

### **Después**
```go
return errors.NotFoundError("user", userID)
```

### **Antes**
```go
return fmt.Errorf("database error: %v", err)
```

### **Después**
```go
return errors.DatabaseError("find_user", err)
```

## 🎯 **Beneficios**

1. **✅ Errores Tipados**: Fácil identificación del tipo de problema
2. **✅ Contexto Rico**: Información adicional para debugging
3. **✅ Consistencia**: Formato estandarizado en toda la aplicación
4. **✅ Mantenibilidad**: Fácil de extender y modificar
5. **✅ Debugging Mejorado**: Stack trace y contexto preservados
6. **✅ APIs Limpias**: Mensajes apropiados para usuarios finales

## 🔗 **Integración con Otros Paquetes**

Este paquete se integra perfectamente con:
- `github.com/pkg/errors` - Funcionalidades avanzadas
- `pkg/responses` - Respuestas HTTP consistentes
- `pkg/logging` - Logging estructurado
- Middleware de manejo de errores

## 🧪 **Tests**

```bash
cd pkg/errors && go test . -cover
```

Cubren `Wrap` (inmutabilidad), `Cause`, clasificación con cadena `%w` y fábricas básicas.

## 📚 **Próximos pasos (opcionales)**

1. Aumentar cobertura (constructores `New`, `Format`, helpers restantes).
2. Roadmap: reducir dependencia de `github.com/pkg/errors` a favor de `fmt.Errorf` con `%w` donde no haga falta stack.
3. Guía de contribución global: cuándo usar este paquete vs `pkg/responses` vs errores de dominio en `internal/*/domain/errors`. 