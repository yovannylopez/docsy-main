package composition

import (
	archiveOpenAPI "github.com/yovannylopez/docsy-main/internal/archive/infrastructure/openapi"
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
	archiveOpenAPI.SetupArchiveSpec(generator)
	sharedOpenAPI.SetupHealthSpec(generator)

	openapi.RegisterStandardErrorResponseSchema(generator)
	openapi.ApplySecurityAndRBACDocumentation(generator)
}
