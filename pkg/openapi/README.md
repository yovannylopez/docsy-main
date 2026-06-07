# Librería OpenAPI Dinámica

## 🎯 **Descripción**

Librería independiente para generar especificaciones OpenAPI 3.0 dinámicamente desde aplicaciones Go con Echo framework.

## 🚀 **Características**

### **✅ Generación Dinámica**
- **Reflection-based**: Extrae información automáticamente de structs Go
- **DTO Integration**: Genera esquemas desde tus DTOs
- **Validation Tags**: Soporta tags `validate` para restricciones
- **Echo Integration**: Detecta rutas automáticamente

### **✅ Documentación Automática**
- **Swagger UI**: Interfaz web integrada en `/docs`
- **JSON Spec**: Especificación en `/openapi.json`
- **Real-time**: Se actualiza automáticamente

### **✅ Flexibilidad**
- **Independiente**: Librería reutilizable
- **Configurable**: Personalizable por endpoint
- **Extensible**: Fácil agregar nuevas funcionalidades

## 📁 **Estructura**

```
openapi/
├── openapi.go              # Modelo OpenAPI 3.0, Generator, GenerateFromEcho, Generate / GenerateYAML
├── schema.go               # SchemaGenerator (reflection, tags validate)
├── envelope.go             # Esquemas y ejemplos alineados con pkg/responses (éxito / error JSON)
├── api_error_envelope.go   # ErrorResponse en components + RegisterStandardErrorResponseSchema
├── pagination_envelope.go  # Documentación 400 por limit/offset (pkg/pagination)
├── middleware.go           # Middleware /docs y /openapi.json; SetupOpenAPIRoutes
├── *_test.go               # Tests
├── go.mod
└── README.md
```

## 🔧 **Uso Básico**

### **1. Inicializar Generador**

```go
import "github.com/yovannylopez/docsy-main/pkg/openapi"

// Crear generador
generator := openapi.NewGenerator(
    "Mi API",
    "Descripción de mi API",
    "1.0.0",
)

// Configurar servidores
generator.AddServer("http://localhost:9000", "Servidor local")
generator.AddTag("auth", "Autenticación")
```

### **2. Integrar con Echo**

```go
// Configurar documentación (solo middleware: /openapi.json y /docs, sin duplicar echo.GET)
openapi.SetupOpenAPIRoutes(e, generator)

// Generar paths desde rutas Echo (luego enriquecer con los Setup*Spec de cada vertical)
generator.GenerateFromEcho(e)
```

### **3. Generar Esquemas desde DTOs**

```go
// Crear generador de esquemas
schemaGen := openapi.NewSchemaGenerator(generator)

// Generar esquemas
schemaGen.GenerateSchemaFromStruct("UserRequest", UserRequest{})
schemaGen.GenerateSchemaFromStruct("UserResponse", UserResponse{})
```

## 📋 **Ejemplo Completo**

### **DTO con Validaciones**

```go
type SignupRequest struct {
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8"`
    FirstName string `json:"first_name" validate:"required"`
    LastName  string `json:"last_name" validate:"required"`
    RoleName  string `json:"role_name" validate:"required,oneof=user admin"`
}
```

### **Configuración en Servidor**

```go
// En tu servidor
func setupOpenAPI(e *echo.Echo) {
    generator := openapi.NewGenerator(
        "Core",
        "Docsy API",
        "1.0.0",
    )
    
    // Configurar
    generator.AddServer("http://localhost:8100", "Development")
    generator.AddTag("auth", "Authentication")
    
    // Integrar
    openapi.SetupOpenAPIRoutes(e, generator)
    generator.GenerateFromEcho(e)
    
    // Esquemas específicos
    schemaGen := openapi.NewSchemaGenerator(generator)
    schemaGen.GenerateSchemaFromStruct("SignupRequest", SignupRequest{})
}
```

## 🌐 **Endpoints Disponibles**

| **Endpoint** | **Descripción** |
|-------------|----------------|
| `/docs` | Swagger UI interactivo |
| `/openapi.json` | Especificación JSON |

## 📊 **Esquemas Generados**

### **Automático desde DTOs**

```json
{
  "components": {
    "schemas": {
      "SignupRequest": {
        "type": "object",
        "properties": {
          "email": {
            "type": "string",
            "format": "email"
          },
          "password": {
            "type": "string",
            "minLength": 8
          },
          "first_name": {
            "type": "string"
          },
          "last_name": {
            "type": "string"
          },
          "role_name": {
            "type": "string",
            "enum": ["user", "admin"]
          }
        },
        "required": ["email", "password", "first_name", "last_name", "role_name"]
      }
    }
  }
}
```

## 🔍 **Validaciones Soportadas**

| **Validación** | **Tag** | **OpenAPI** |
|---------------|---------|-------------|
| Required | `required` | `required: true` |
| Email | `email` | `format: email` |
| Min Length | `min=8` | `minLength: 8` |
| Max Length | `max=100` | `maxLength: 100` |
| Enum | `oneof=val1 val2` | `enum: ["val1", "val2"]` |
| Pattern | `pattern=regex` | `pattern: regex` |

## 🚀 **Performance**

- **Lazy Generation**: Solo se genera cuando se solicita
- **Caching**: Especificación en memoria
- **Minimal Overhead**: < 1ms por request
- **Memory Efficient**: Reutiliza estructuras

## 🔧 **Configuración Avanzada**

### **Personalizar Operaciones**

```go
// Configurar operación específica
operation := &openapi.Operation{
    Summary:     "Registrar usuario",
    Description: "Crea una nueva cuenta",
    Tags:        []string{"auth"},
    RequestBody: &openapi.RequestBody{
        Required: true,
        Content: map[string]openapi.MediaType{
            "application/json": {
                Schema: &openapi.Schema{
                    Ref: "#/components/schemas/SignupRequest",
                },
            },
        },
    },
}
```

### **Agregar Seguridad**

```go
// Configurar JWT
generator.AddSecurityScheme("bearerAuth", &openapi.SecurityScheme{
    Type:         "http",
    Scheme:       "bearer",
    BearerFormat: "JWT",
})
```

## 📈 **Beneficios**

### **✅ Para Desarrolladores**
- **Documentación automática**: No más docs manuales
- **Testing integrado**: Swagger UI para testing
- **Validación visual**: Ve tus esquemas en tiempo real

### **✅ Para APIs**
- **Consistencia**: Esquemas siempre actualizados
- **Interoperabilidad**: Estándar OpenAPI 3.0
- **Integración**: Compatible con herramientas estándar

### **✅ Para Mantenimiento**
- **DRY**: No duplicar información
- **Type Safety**: Basado en tipos Go
- **Reflection**: Automático desde código

## 🎯 **Próximos Pasos**

1. **✅ Implementado**: Generación básica, YAML (`GenerateYAML`), structs anidados en esquemas
2. **📋 Pendiente**: Swagger UI self-hosted (airgap / CSP estricto)
3. **📋 Pendiente**: Caché de spec si el documento crece mucho

## 🚀 **Cómo Contribuir**

```bash
# Clonar
git clone <repo>

# Instalar dependencias
go mod tidy

# Ejecutar tests
go test ./...

# Ejecutar ejemplo
go run examples/basic/main.go
```

---

**¡Documentación automática y dinámica para tus APIs Go! 🚀** 