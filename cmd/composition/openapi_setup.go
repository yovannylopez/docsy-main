package composition

import (
	authOpenAPI "github.com/yovannylopez/docsy-main/internal/auth/infrastructure/openapi"
	sharedOpenAPI "github.com/yovannylopez/docsy-main/internal/shared/infrastructure/openapi"
	usersOpenAPI "github.com/yovannylopez/docsy-main/internal/users/infrastructure/openapi"
	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

// SetupAllSpecs registers all the OpenAPI specifications of the base modules.
// To add a new module, call its SetupXxxSpec(generator) here.
func SetupAllSpecs(generator *openapi.Generator) {
	authOpenAPI.SetupAuthSpec(generator)
	authOpenAPI.SetupAuditSpec(generator)
	usersOpenAPI.SetupUsersSpec(generator)
	sharedOpenAPI.SetupHealthSpec(generator)

	// Add here the specs of your business modules:
	// productsOpenAPI.SetupProductsSpec(generator)

	openapi.RegisterStandardErrorResponseSchema(generator)
	openapi.ApplySecurityAndRBACDocumentation(generator)
}
