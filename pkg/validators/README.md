# Librería de Validación Centralizada

Esta librería proporciona validaciones reutilizables y centralizadas para todo el proyecto. El objetivo es evitar duplicación de código de validación y mantener consistencia en las reglas de validación.

**Import:** `github.com/yovannylopez/docsy-main/pkg/validators` (convención: alias **`validators`**). Los regex estáticos (email, teléfono por defecto, reglas de contraseña y nombre) se compilan **una vez** al cargar el paquete; los patrones dinámicos se reutilizan vía **caché** por cadena de patrón.

## 🎯 **Características**

- ✅ **Validaciones reutilizables** para campos comunes
- ✅ **Validadores específicos** para autenticación y autorización
- ✅ **Mensajes de error personalizables** en español
- ✅ **Validadores configurables** para diferentes políticas
- ✅ **Fácil integración** con cualquier parte del proyecto

## 📦 **Validadores Disponibles**

### **Validadores Básicos**

#### `StringValidator`
Valida strings con longitud mínima/máxima (contadas en **runas Unicode**, no en bytes) y patrones regex.

```go
// String requerido con longitud 5-50 caracteres
validator := RequiredString(5, 50)

// String opcional con patrón específico
validator := StringWithPattern(`^[A-Za-z]+$`, false, 2, 20)
```

#### `EmailValidator`
Valida formatos de email.

```go
// Email requerido
validator := RequiredEmail()

// Email opcional
validator := OptionalEmail()
```

#### `NumberValidator`
Valida números con rangos y restricciones. Los límites **`Min`** y **`Max`** son punteros (`*float64`): **`nil`** significa *sin límite* en ese extremo, de modo que rangos como **`[0, 100]`** funcionan correctamente (antes un mínimo `0` podía interpretarse como “sin mínimo”).

```go
// Número positivo requerido
validator := RequiredPositiveNumber()

// Número en rango inclusivo [0, 100]
validator := RequiredNumberRange(0, 100)
```

#### Mensajes orientados al cliente (`ValidationError`)

`Validate` puede devolver `ValidationError`. El método `Error()` incluye campo y valor para logs; para respuestas HTTP usa el mensaje corto:

- **`(ValidationError).ClientMessage()`** — solo el texto del mensaje (p. ej. `campo requerido`, `longitud mínima: 10 caracteres`).
- **`ErrorClientMessage(err)`** — si `err` es o envuelve un `ValidationError`, devuelve `ClientMessage()`; si no, devuelve `err.Error()`.

Úsalo en handlers: `responses.BadRequest(c, validators.ErrorClientMessage(err))`.

#### Helpers de string

- **`OptionalMaxLength(max)`** — string opcional; si viene vacío es válido; si tiene valor, no puede superar `max` runas.
- **`MinLengthPassword(minLen)`** — contraseña con longitud mínima únicamente (sin mayúsculas/símbolos, etc.). Para políticas completas sigue usando `StandardPassword` / `SimplePassword` o un `PasswordValidator` personalizado.

### **Validadores de Autenticación**

#### `PasswordValidator`
Valida contraseñas con políticas configurables.

```go
// Contraseña estándar (8+ chars, mayúsculas, minúsculas, números, símbolos)
validator := StandardPassword()

// Contraseña simple (6+ chars, solo minúsculas)
validator := SimplePassword()

// Contraseña personalizada
customPassword := PasswordValidator{
    Required:       true,
    MinLength:      10,
    RequireUpper:   true,
    RequireLower:   true,
    RequireNumbers: true,
    RequireSymbols: false,
}
```

#### `PhoneValidator`
Valida números de teléfono.

```go
// Teléfono internacional requerido
validator := InternationalPhone()

// Teléfono estándar opcional
validator := StandardPhone()

// Teléfono personalizado (solo españoles)
customPhone := PhoneValidator{
    Required: true,
    Pattern:  `^\+34\d{9}$`,
}
```

#### `RoleValidator`
Valida roles de usuario. Si **`AllowedRoles`** está **vacío** y el valor no es vacío/nil, la validación falla con un mensaje explícito (evita aceptar cualquier rol por error).

```go
// Roles del sistema (alineados con datos sembrados en migraciones)
validator := SystemRoles()

// Roles de usuario (sin super_admin)
validator := UserRoles()

// Roles personalizados
customRoles := RoleValidator{
    Required:     true,
    AllowedRoles: []string{"user", "moderator"},
}
```

#### `NameValidator`
Valida nombres de persona.

```go
// Nombre requerido (2-100 caracteres)
validator := PersonName()

// Nombre opcional
validator := OptionalPersonName()
```

## 🚀 **Uso en el Proyecto**

### **Validación Individual**

```go
import "github.com/yovannylopez/docsy-main/pkg/validators"

// Validar email
emailValidator := validators.RequiredEmail()
if err := emailValidator.Validate("test@example.com"); err != nil {
    // Manejar error
}

// Validar contraseña
passwordValidator := validators.StandardPassword()
if err := passwordValidator.Validate("Password123!"); err != nil {
    // Manejar error
}
```

### **Validación de campos de creación de usuario**

```go
import "github.com/yovannylopez/docsy-main/pkg/validators"

func validateCreateUserFields(email, password, firstName, lastName, roleName string) error {
    checks := []struct {
        field string
        fn    func(string) error
        value string
    }{
        {"email", validators.RequiredEmail().Validate, email},
        {"password", validators.StandardPassword().Validate, password},
        {"first_name", validators.PersonName().Validate, firstName},
        {"last_name", validators.PersonName().Validate, lastName},
        {"role_name", validators.SystemRoles().Validate, roleName},
    }
    for _, c := range checks {
        if err := c.fn(c.value); err != nil {
            return fmt.Errorf("%s: %w", c.field, err)
        }
    }
    return nil
}
```

En transport, los DTOs de `internal/users` usan tags `validate` con Echo Bind; los validadores de este paquete sirven para reglas reutilizables en use cases o handlers cuando haga falta.

### **Validación en Handlers**

```go
func (h *UsersHandler) CreateUser(c echo.Context) error {
    var request dtos.CreateUserRequest
    if err := c.Bind(&request); err != nil {
        return responses.BadRequest(c, "invalid JSON")
    }
    if err := validators.RequiredEmail().Validate(request.Email); err != nil {
        return responses.BadRequest(c, validators.ErrorClientMessage(err))
    }
    if err := validators.StandardPassword().Validate(request.Password); err != nil {
        return responses.BadRequest(c, validators.ErrorClientMessage(err))
    }
    // ... use case ...
}
```

## 🔧 **Configuración de Validadores**

### **Validadores Compuestos**

```go
// Combinar múltiples validadores
compositeValidator := validators.Combine(
    validators.RequiredEmail(),
    validators.StandardPassword(),
    validators.PersonName(),
)
```

### **Validadores Personalizados**

```go
// Crear validador de contraseña con política específica
customPassword := validators.PasswordValidator{
    Required:       true,
    MinLength:      12,
    RequireUpper:   true,
    RequireLower:   true,
    RequireNumbers: true,
    RequireSymbols: true,
}

// Crear validador de teléfono para región específica
customPhone := validators.PhoneValidator{
    Required: true,
    Pattern:  `^\+1\d{10}$`, // Solo números de EE.UU.
}
```

## 📋 **Políticas de Validación Estándar**

### **Contraseñas**
- **Estándar**: 8+ caracteres, mayúsculas, minúsculas, números, símbolos
- **Simple**: 6+ caracteres, solo minúsculas

### **Emails**
- Formato RFC 5322 básico
- Validación de dominio

### **Teléfonos**
- Formato internacional E.164
- Patrones configurables por región

### **Nombres**
- 2-100 caracteres (**runas**; acentos y CJK cuentan como un carácter cada uno)
- Solo letras, espacios y caracteres especiales básicos
- Soporte para acentos y ñ

### **Roles**
- **SystemRoles:** `super_admin`, `correspondence_admin`, `correspondence_operator`, `dependency_manager`, `funcionario`, `user`, `visualizador`
- **UserRoles:** mismos salvo `super_admin`

## 🧪 **Testing**

```go
func TestPasswordValidation(t *testing.T) {
    validator := validators.StandardPassword()
    
    // Casos válidos
    validPasswords := []string{
        "Password123!",
        "MySecurePass1@",
        "ComplexP@ssw0rd",
    }
    
    for _, password := range validPasswords {
        if err := validator.Validate(password); err != nil {
            t.Errorf("Password válida rechazada: %s - %v", password, err)
        }
    }
    
    // Casos inválidos
    invalidPasswords := []string{
        "weak",
        "password",
        "PASSWORD",
        "12345678",
    }
    
    for _, password := range invalidPasswords {
        if err := validator.Validate(password); err == nil {
            t.Errorf("Password inválida aceptada: %s", password)
        }
    }
}
```

## 🔄 **Migración desde Validaciones Existentes**

### **Antes (solo tags en el struct)**
```go
type CreateUserRequest struct {
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8"`
    FirstName string `json:"first_name" validate:"required,min=2"`
    RoleName  string `json:"role_name" validate:"required"`
}
```

### **Después (validadores reutilizables en handler o use case)**
```go
if err := validators.RequiredEmail().Validate(request.Email); err != nil {
    return responses.BadRequest(c, validators.ErrorClientMessage(err))
}
if err := validators.SystemRoles().Validate(request.RoleName); err != nil {
    return responses.BadRequest(c, validators.ErrorClientMessage(err))
}
```

## 🎯 **Beneficios**

1. **Reutilización**: Una validación, múltiples usos
2. **Consistencia**: Mismas reglas en toda la aplicación
3. **Mantenibilidad**: Cambios centralizados
4. **Testabilidad**: Validaciones fáciles de probar
5. **Flexibilidad**: Validadores configurables
6. **Legibilidad**: Código más limpio y expresivo

## 📚 **Próximos Pasos**

- [ ] Agregar validadores para otros dominios (documentos, configuraciones, etc.)
- [ ] Implementar validadores para validaciones de negocio complejas
- [ ] Agregar soporte para validaciones condicionales
- [ ] Crear validadores para validaciones de API externas
- [ ] Implementar cache de validadores para mejor rendimiento 