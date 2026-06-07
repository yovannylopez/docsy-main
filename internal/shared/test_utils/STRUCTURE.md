# Estructura del Paquete TestUtils

## 🎯 **Ubicación Correcta: `internal/shared/test_utils`**

### **¿Por qué en la raíz de `shared` y no en `infrastructure`?**

El paquete `testutils` está ubicado en `internal/shared/testutils` (raíz de shared) en lugar de `internal/shared/infrastructure/testutils` por las siguientes razones:

## 📋 **Razones Arquitectónicas**

### **1. Utilidad Transversal**
```go
// ✅ Correcto - Accesible desde cualquier módulo
import "github.com/yovannylopez/docsy-main/internal/shared/test_utils"

// ❌ Incorrecto - Limitado solo a infrastructure
import "github.com/yovannylopez/docsy-main/internal/shared/infrastructure/test_utils"
```

### **2. Principio de Responsabilidad Única**
- **`infrastructure/`**: Contiene implementaciones de infraestructura (DB, HTTP, etc.)
- **`testutils/`**: Contiene utilidades para testing (stubs, mocks, helpers)

### **3. Reutilización en Todas las Capas**
```go
// Puede ser usado en:
// - internal/auth/domain/entities/ (para tests de entidades)
// - internal/auth/usecases/ (para tests de casos de uso)
// - internal/auth/transport/ (para tests de handlers)
// - internal/shared/infrastructure/ (para tests de infraestructura)
// - pkg/* (para tests de paquetes públicos)
```

## 🏗️ **Estructura de Clean Architecture**

```
internal/
├── auth/                          # Módulo de autenticación
│   ├── domain/                    # Entidades y reglas de negocio
│   ├── usecases/                  # Casos de uso
│   ├── infrastructure/            # Implementaciones de infraestructura
│   └── transport/                 # Handlers HTTP
├── shared/                        # Utilidades compartidas
│   ├── domain/                    # Entidades compartidas
│   ├── infrastructure/            # Infraestructura compartida
│   ├── transport/                 # Transporte compartido
│   └── test_utils/                 # 🎯 Utilidades de testing
└── other-modules/                 # Otros módulos
```

## 🔄 **Flujo de Uso**

### **Desde Domain Layer**
```go
// internal/auth/domain/entities/user_test.go
package entities

import (
    "testing"
    "github.com/yovannylopez/docsy-main/internal/shared/test_utils"
)

func TestUser_Validation(t *testing.T) {
    stubs := test_utils.NewStubs()
    user := stubs.GetTestUser("admin")
    // ... test logic
}
```

### **Desde Use Cases**
```go
// internal/auth/usecases/login_usecase_test.go
package usecases

import (
    "testing"
    "github.com/yovannylopez/docsy-main/internal/shared/test_utils"
)

func TestLoginUseCase(t *testing.T) {
    stubs := test_utils.NewStubs()
    user := stubs.GetTestUser("admin")
    password := stubs.GetTestPassword("valid")
    // ... test logic
}
```

### **Desde Infrastructure**
```go
// internal/auth/infrastructure/repositories/user_repository_test.go
package repositories

import (
    "testing"
    "github.com/yovannylopez/docsy-main/internal/shared/test_utils"
)

func TestUserRepository_Create(t *testing.T) {
    stubs := test_utils.NewStubs()
    user := stubs.GetTestUser("regular")
    // ... test logic
}
```

### **Desde Transport**
```go
// internal/auth/transport/handlers/auth_handler_test.go
package handlers

import (
    "testing"
    "github.com/yovannylopez/docsy-main/internal/shared/test_utils"
)

func TestAuthHandler_Login(t *testing.T) {
    stubs := test_utils.NewStubs()
    config := stubs.GetTestConfig("valid")
    // ... test logic
}
```

## 📦 **Beneficios de la Ubicación Correcta**

### **1. Accesibilidad Universal**
- ✅ Accesible desde cualquier módulo del proyecto
- ✅ No requiere importaciones complejas
- ✅ Sigue el principio de dependencias hacia adentro

### **2. Separación de Responsabilidades**
- ✅ `infrastructure/`: Implementaciones concretas
- ✅ `testutils/`: Utilidades de testing
- ✅ `domain/`: Lógica de negocio
- ✅ `transport/`: Capa de presentación

### **3. Mantenibilidad**
- ✅ Cambios centralizados
- ✅ Fácil de encontrar
- ✅ Documentación clara
- ✅ Reutilización máxima

## 🎯 **Principios Aplicados**

### **Clean Architecture**
```
┌─────────────────────────────────────┐
│           TestUtils                 │ ← Utilidad transversal
├─────────────────────────────────────┤
│        Transport Layer              │ ← Handlers, Routes
├─────────────────────────────────────┤
│        Use Cases Layer              │ ← Business Logic
├─────────────────────────────────────┤
│        Domain Layer                 │ ← Entities, Rules
├─────────────────────────────────────┤
│     Infrastructure Layer            │ ← DB, External APIs
└─────────────────────────────────────┘
```

### **Dependency Inversion**
- Los stubs no dependen de implementaciones específicas
- Proporcionan abstracciones para testing
- Siguen el principio de inversión de dependencias

### **Single Responsibility**
- `testutils/`: Solo responsabilidad de testing
- `infrastructure/`: Solo responsabilidad de infraestructura
- `domain/`: Solo responsabilidad de negocio

## 📝 **Convenciones de Naming**

### **Estructura de Archivos**
```
internal/shared/testutils/
├── stubs.go                    # Stubs centralizados
├── example_usage_test.go       # Ejemplos de uso
├── README.md                   # Documentación principal
└── STRUCTURE.md               # Este archivo
```

### **Naming de Funciones**
```go
// ✅ Correcto
stubs := testutils.NewStubs()
user := stubs.GetTestUser("admin")
config := stubs.GetTestConfig("valid")

// ❌ Incorrecto
stubs := testutils.CreateStubs()
user := stubs.GetUser("admin")
config := stubs.GetConfig("valid")
```

## 🔄 **Migración de Tests Existentes**

### **Antes (Ubicación Incorrecta)**
```go
import "github.com/yovannylopez/docsy-main/internal/shared/infrastructure/test_utils"
```

### **Después (Ubicación Correcta)**
```go
import "github.com/yovannylopez/docsy-main/internal/shared/test_utils"
```

## 🎯 **Ventajas de la Ubicación Actual**

1. **Accesibilidad**: Cualquier módulo puede usar los stubs
2. **Simplicidad**: Importaciones más cortas y claras
3. **Flexibilidad**: Fácil agregar nuevos tipos de stubs
4. **Mantenibilidad**: Cambios centralizados
5. **Escalabilidad**: Crece con el proyecto
6. **Documentación**: Ubicación intuitiva

## 📚 **Referencias**

- [Clean Architecture - Uncle Bob](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Testing Best Practices](https://github.com/golang/go/wiki/TableDrivenTests)
