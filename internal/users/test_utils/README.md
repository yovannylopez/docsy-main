# Users Test Utils (Object Mother)

Paquete **Object Mother** del bounded context `users`. Los tests en `internal/users/` deben obtener DTOs y entidades (`CreateUsersRequest`, usuarios de listado/búsqueda, etc.) desde aquí, reutilizando formas de usuario alineadas con `internal/auth/domain/entities` vía `Clone` del mother de auth cuando aplique.

## Uso

```go
import userstest "github.com/yovannylopez/docsy-main/internal/users/test_utils"

stubs := userstest.NewUsersStubs()
req := stubs.CreateUsersRequestStandard()
u := stubs.UserJohnPerez()
```

## API principal

- `NewUsersStubs()` — factoría del slice users.
- Altas: `CreateUsersRequestStandard`, `CreateUsersRequestMinimal`.
- Entidades de ejemplo: `UserJohnPerez`, `UserMariaGarcia`, `UserJohnDoeSearch`, `UserJaneSmithSearch`, `UserTestGeneric`, `UserExistingByEmail`, `UserForUpdate`.
- Búsqueda: `SearchRequestWithActivo`, `SearchRequestNoActivo`.
- Actualización: `UpdateRequestFirstName`, `UpdateRequestEmail`.
- Punteros a `string` en tests: usar `authtest.StringPtr` desde `internal/auth/test_utils` (canónico compartido).

Normativa completa: `.cursor/skills/go-testing/SKILL.md`.
