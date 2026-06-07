# Casos de Uso del Core

Este directorio contiene los casos de uso de la aplicación, siguiendo los principios de Clean Architecture.

## Casos de Uso Disponibles

### SignupUseCase
- **Archivo**: `signup_usecase.go`
- **Propósito**: Maneja el registro de nuevos usuarios
- **Complejidad**: Alta (múltiples validaciones y operaciones)
- **Estado**: ✅ Implementado

### LoginUseCase
- **Archivo**: `login_usecase.go`
- **Propósito**: Maneja la autenticación de usuarios
- **Complejidad**: Media (validación de credenciales y creación de sesión)
- **Estado**: ✅ Implementado

## Estructura de Casos de Uso

Cada caso de uso sigue el patrón:

```go
type [Name]UseCase struct {
    // Dependencias inyectadas
}

func (uc *[Name]UseCase) Execute(ctx context.Context, request *dtos.[Name]Request) (*dtos.[Name]Response, error) {
    // Lógica de negocio
}
```

## Principios Aplicados

1. **Separación de Responsabilidades**: Cada caso de uso maneja una operación específica
2. **Inyección de Dependencias**: Los repositorios y servicios se inyectan
3. **Testabilidad**: Fácil de probar de forma aislada
4. **Consistencia**: Mismo patrón para todos los casos de uso

## Beneficios de Usar Casos de Uso

- **Mantenibilidad**: Lógica de negocio centralizada
- **Testabilidad**: Fácil de probar de forma unitaria
- **Reutilización**: Pueden ser usados por diferentes handlers
- **Escalabilidad**: Fácil agregar nuevas funcionalidades
- **Consistencia**: Mismo patrón en toda la aplicación 