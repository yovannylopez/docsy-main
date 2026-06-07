# Mocks del Módulo Shared

Este directorio contiene los mocks compartidos utilizados por múltiples módulos del microservicio.

## Estructura

```
internal/shared/mocks/
└── README.md                       # Documentación de mocks compartidos
```

**Nota**: Actualmente no hay interfaces en el módulo shared que requieran mocks. Los componentes compartidos son principalmente implementaciones concretas.

## Generación de mocks

Actualmente no hay interfaces en el módulo shared que requieran mocks. Si en el futuro se agregan interfaces, se pueden generar mocks usando:

```bash
# Ejemplo de generación de mocks (cuando existan interfaces)
mockery --dir=internal/shared/domain/ports --output=internal/shared/mocks --outpkg=mocks --filename=interface_mock.go --name=InterfaceName
```

## Uso en pruebas

```go
package handlers_test

import (
    "testing"
    "net/http/httptest"
    
    "github.com/stretchr/testify/mock"
    "github.com/labstack/echo/v4"
    "github.com/yovannylopez/docsy-main/internal/shared/mocks"
    "github.com/yovannylopez/docsy-main/internal/shared/transport/handlers"
)

func TestHealthHandler(t *testing.T) {
    // Ejemplo de prueba sin mocks (usando implementaciones reales)
    db := database.NewConnection(config)
    logger := logging.NewLogger()
    
    // Crear handler con implementaciones reales
    handler := handlers.NewHealthHandler(db, logger)

    // Crear request de prueba
    e := echo.New()
    req := httptest.NewRequest("GET", "/health", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    // Ejecutar prueba
    err := handler.HealthCheck(c)
    // Assertions...
}
```

## Uso en otros módulos

Los mocks compartidos pueden ser utilizados por otros módulos:

```go
package auth_test

import (
    "testing"
    
    "github.com/stretchr/testify/mock"
    "github.com/yovannylopez/docsy-main/internal/auth/mocks"
    sharedMocks "github.com/yovannylopez/docsy-main/internal/shared/mocks"
    "github.com/yovannylopez/docsy-main/internal/auth/usecases"
)

func TestAuthUseCase(t *testing.T) {
    // Mocks específicos del módulo auth
    mockAuthService := authMocks.NewAuthenticationService(t)
    
    // Implementaciones reales para componentes compartidos
    db := database.NewConnection(config)
    logger := logging.NewLogger()

    // Configurar expectativas...
    useCase := usecases.NewAuthUseCase(mockAuthService, db, logger)
    // Ejecutar prueba...
}
```

## Notas importantes

- Los componentes compartidos son principalmente implementaciones concretas
- No hay interfaces en el módulo shared que requieran mocks
- Si se agregan interfaces en el futuro, se pueden generar mocks aquí
- Los componentes compartidos se pueden usar directamente en pruebas 