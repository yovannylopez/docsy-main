# Test Utils - Stubs Centralizados

Este paquete proporciona stubs centralizados para todos los datos de prueba utilizados en el proyecto. Esto evita la duplicación de código y asegura consistencia en los tests.

## 📍 **Ubicación: `internal/shared/test_utils`**

El paquete está ubicado en la raíz de `shared` para ser accesible desde cualquier módulo del proyecto, siguiendo los principios de Clean Architecture.

## 🎯 **Beneficios**

- **Consistencia**: Todos los tests usan los mismos datos de prueba
- **Mantenibilidad**: Cambios en un solo lugar se reflejan en todos los tests
- **Legibilidad**: Tests más limpios y fáciles de entender
- **Reutilización**: No hay duplicación de código de configuración

## 📦 **Estructura de Stubs**

### **Configuraciones de Prueba**
```go
stubs := NewStubs()

// Obtener configuraciones específicas
validConfig := stubs.GetTestConfig("valid")
invalidConfig := stubs.GetTestConfig("invalid")
devConfig := stubs.GetTestConfig("development")
prodConfig := stubs.GetTestConfig("production")
minimalConfig := stubs.GetTestConfig("minimal")
```

### **Usuarios de Prueba**
```go
// Obtener usuarios específicos
adminUser := stubs.GetTestUser("admin")
regularUser := stubs.GetTestUser("regular")
multiUser := stubs.GetTestUser("multi")
emptyUser := stubs.GetTestUser("empty")
invalidUser := stubs.GetTestUser("invalid")
```

### **Contraseñas de Prueba**
```go
// Obtener contraseñas específicas
validPassword := stubs.GetTestPassword("valid")
complexPassword := stubs.GetTestPassword("complex")
emptyPassword := stubs.GetTestPassword("empty")
longPassword := stubs.GetTestPassword("long")
specialPassword := stubs.GetTestPassword("special")
unicodePassword := stubs.GetTestPassword("unicode")
```

### **Roles de Prueba**
```go
// Obtener roles específicos
adminRole := stubs.GetTestRole("admin")
userRole := stubs.GetTestRole("user")
guestRole := stubs.GetTestRole("guest")
emptyRole := stubs.GetTestRole("empty")
```

### **Contenedores de Prueba**
```go
// Obtener contenedores específicos
validContainer := stubs.GetTestContainer("valid")
mockContainer := stubs.GetTestContainer("mock")
emptyContainer := stubs.GetTestContainer("empty")
invalidContainer := stubs.GetTestContainer("invalid")
```

## 🚀 **Ejemplos de Uso**

### **Test de Configuración**
```go
func TestNewServer_WithStubs(t *testing.T) {
    stubs := NewStubs()
    
    tests := []struct {
        name        string
        configType  string
        expectError bool
    }{
        {
            name:        "valid config",
            configType:  "valid",
            expectError: false,
        },
        {
            name:        "invalid config",
            configType:  "invalid",
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config := stubs.GetTestConfig(tt.configType)
            server, err := NewServer(config)
            
            if tt.expectError {
                assert.Error(t, err)
                assert.Nil(t, server)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, server)
            }
        })
    }
}
```

### **Test de Password Hashing**
```go
func TestPasswordHashing_WithStubs(t *testing.T) {
    stubs := NewStubs()
    hasher := security.NewPasswordHasher()
    
    testCases := []struct {
        name        string
        password    string
        expectError bool
    }{
        {
            name:        "valid password",
            password:    stubs.GetTestPassword("valid"),
            expectError: false,
        },
        {
            name:        "complex password",
            password:    stubs.GetTestPassword("complex"),
            expectError: false,
        },
        {
            name:        "long password",
            password:    stubs.GetTestPassword("long"),
            expectError: true, // bcrypt limit
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            hash, err := hasher.Hash(tc.password)
            
            if tc.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotEmpty(t, hash)
                
                // Verify password
                result := hasher.Verify(tc.password, hash)
                assert.True(t, result)
            }
        })
    }
}
```

### **Test de Token Generation**
```go
func TestTokenGeneration_WithStubs(t *testing.T) {
    stubs := NewStubs()
    tokenGenerator := security.NewTokenGenerator("test-secret")
    
    testCases := []struct {
        name        string
        userType    string
        expectError bool
    }{
        {
            name:        "admin user",
            userType:    "admin",
            expectError: false,
        },
        {
            name:        "user with roles",
            userType:    "multi",
            expectError: false,
        },
        {
            name:        "empty user",
            userType:    "empty",
            expectError: false,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            user := stubs.GetTestUser(tc.userType)
            token, err := tokenGenerator.GenerateToken(user)
            
            if tc.expectError {
                assert.Error(t, err)
                assert.Nil(t, token)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, token)
                assert.NotEmpty(t, token.AccessToken)
                assert.NotEmpty(t, token.RefreshToken)
            }
        })
    }
}
```

### **Test de Usuarios**
```go
func TestUserValidation_WithStubs(t *testing.T) {
    stubs := NewStubs()
    
    testCases := []struct {
        name        string
        userType    string
        expectValid bool
    }{
        {
            name:        "admin user",
            userType:    "admin",
            expectValid: true,
        },
        {
            name:        "regular user",
            userType:    "regular",
            expectValid: true,
        },
        {
            name:        "invalid user",
            userType:    "invalid",
            expectValid: false,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            user := stubs.GetTestUser(tc.userType)
            
            if tc.expectValid {
                assert.NotEmpty(t, user.ID)
                assert.NotEmpty(t, user.Email)
                assert.NotEmpty(t, user.PasswordHash)
                assert.True(t, user.IsActive)
                assert.True(t, user.IsVerified)
            } else {
                assert.Empty(t, user.ID)
                assert.Empty(t, user.Email)
                assert.False(t, user.IsActive)
                assert.False(t, user.IsVerified)
            }
        })
    }
}
```

## 🔧 **Tipos de Datos Disponibles**

### **Configuraciones**
- `valid`: Configuración válida para tests normales
- `invalid`: Configuración inválida para tests de error
- `development`: Configuración para desarrollo
- `production`: Configuración para producción
- `minimal`: Configuración mínima para tests rápidos

### **Usuarios**
- `admin`: Usuario administrador con rol admin
- `regular`: Usuario regular con rol user
- `multi`: Usuario con múltiples roles
- `empty`: Usuario sin roles
- `invalid`: Usuario inválido para tests de error

### **Contraseñas**
- `valid`: Contraseña válida normal
- `complex`: Contraseña con caracteres especiales
- `empty`: Contraseña vacía
- `long`: Contraseña muy larga (excede límite bcrypt)
- `special`: Contraseña solo con caracteres especiales
- `unicode`: Contraseña con caracteres Unicode

### **Roles**
- `admin`: Rol de administrador
- `user`: Rol de usuario regular
- `guest`: Rol de invitado
- `empty`: Rol vacío

### **Contenedores**
- `valid`: Contenedor válido con todos los handlers
- `mock`: Contenedor mock para tests unitarios
- `empty`: Contenedor vacío para tests de error
- `invalid`: Contenedor inválido para tests de error

## 📝 **Mejores Prácticas**

### **1. Usar Stubs en Todos los Tests**
```go
// ✅ Correcto
func TestSomething(t *testing.T) {
    stubs := test_utils.NewStubs()
    config := stubs.GetTestConfig("valid")
    user := stubs.GetTestUser("admin")
    // ... test logic
}

// ❌ Incorrecto - Datos hardcodeados
func TestSomething(t *testing.T) {
    config := &CoreConfig{
        Server: ServerConfig{Port: "8080"},
        // ... más configuración hardcodeada
    }
    // ... test logic
}
```

### **2. Usar Nombres Descriptivos**
```go
// ✅ Correcto
testCases := []struct {
    name:        "admin user with valid config",
    userType:    "admin",
    configType:  "valid",
    expectError: false,
}

// ❌ Incorrecto
testCases := []struct {
    name:        "test 1",
    userType:    "user1",
    configType:  "config1",
    expectError: false,
}
```

### **3. Agrupar Tests Relacionados**
```go
func TestAuthentication_WithStubs(t *testing.T) {
    stubs := NewStubs()
    
    t.Run("valid login", func(t *testing.T) {
        user := stubs.GetTestUser("admin")
        password := stubs.GetTestPassword("valid")
        // ... test logic
    })
    
    t.Run("invalid login", func(t *testing.T) {
        user := stubs.GetTestUser("admin")
        password := stubs.GetTestPassword("empty")
        // ... test logic
    })
}
```

### **4. Usar Helpers para Casos Comunes**
```go
func setupTestEnvironment(t *testing.T) (*Stubs, *security.PasswordHasher, *security.TokenGenerator) {
    stubs := NewStubs()
    hasher := security.NewPasswordHasher()
    tokenGenerator := security.NewTokenGenerator("test-secret")
    
    return stubs, hasher, tokenGenerator
}

func TestAuthenticationFlow(t *testing.T) {
    stubs, hasher, tokenGenerator := setupTestEnvironment(t)
    // ... test logic using stubs
}
```

## 🔄 **Migración de Tests Existentes**

Para migrar tests existentes a usar stubs:

1. **Identificar datos duplicados**
2. **Reemplazar con stubs correspondientes**
3. **Actualizar assertions si es necesario**
4. **Ejecutar tests para verificar funcionamiento**

### **Ejemplo de Migración**

**Antes:**
```go
func TestUserRepository_Create(t *testing.T) {
    user := &entities.User{
        ID:        "test-id",
        Email:     "test@example.com",
        Password:  "hashed-password",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    // ... test logic
}
```

**Después:**
```go
func TestUserRepository_Create(t *testing.T) {
    stubs := NewStubs()
    user := stubs.GetTestUser("regular")
    // ... test logic
}
```

## 🎯 **Ventajas de Usar Stubs**

1. **Consistencia**: Todos los tests usan los mismos datos
2. **Mantenibilidad**: Cambios centralizados
3. **Legibilidad**: Tests más limpios
4. **Reutilización**: No hay duplicación
5. **Flexibilidad**: Fácil agregar nuevos tipos de datos
6. **Documentación**: Los stubs sirven como documentación de los datos esperados

## Pruebas HTTP con `CentralHTTPErrorHandler`

Para handlers que **propagan** `error` y delegan la respuesta al `HTTPErrorHandler` global (como en `cmd/composition`):

```go
import sharedtest "github.com/yovannylopez/docsy-main/internal/shared/test_utils"

e := sharedtest.NewEchoWithCentralHTTPErrorHandler()
e.GET("/api/v1/recurso", handler.Acción)
rec := sharedtest.ServeEcho(e, httptest.NewRequest(http.MethodGet, "/api/v1/recurso", nil))
sharedtest.AssertCentralJSONInternalServerError(t, rec, "unknown") // cadena vacía: no valida request_id
```

`DecodeAPIErrorEnvelope` permite leer `error` + `request_id` cuando el status o la forma del cuerpo no es el 500 genérico.

En tests de handlers, las aserciones sobre `rec.Code` suelen usar `http_status.OK.Code`, `http_status.BadRequest.Code`, etc., para alinearlas con `pkg/http_status` y con la API real.

## 📚 **Referencias**

- [Go Testing Best Practices](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Clean Architecture Testing](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
