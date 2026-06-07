package openapi_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	authopenapi "github.com/yovannylopez/docsy-main/internal/auth/infrastructure/openapi"
	pkgopenapi "github.com/yovannylopez/docsy-main/pkg/openapi"
)

func TestSetupAuthSpecAndAuditSpec_RegistersPaths(t *testing.T) {
	gen := pkgopenapi.NewGenerator("Docsy API", "Test", "1.0.0")
	spec := gen.GetSpec()

	for _, p := range []string{
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/api/v1/auth/logout",
		"/api/v1/auth/change-password",
	} {
		spec.Paths[p] = pkgopenapi.PathItem{
			Post: &pkgopenapi.Operation{Responses: map[string]pkgopenapi.Response{}},
		}
	}

	require.NotPanics(t, func() {
		authopenapi.SetupAuthSpec(gen)
		authopenapi.SetupAuditSpec(gen)
		pkgopenapi.RegisterStandardErrorResponseSchema(gen)
		pkgopenapi.ApplySecurityAndRBACDocumentation(gen)
	})

	require.Contains(t, spec.Paths, "/api/v1/auditoria")
	require.NotNil(t, spec.Paths["/api/v1/auditoria"].Get)
	require.NotEmpty(t, spec.Paths["/api/v1/auditoria"].Get.OperationID)
	require.Contains(t, spec.Components.SecuritySchemes, pkgopenapi.SecuritySchemeBearer)
	require.Empty(t, spec.Paths["/api/v1/auth/login"].Post.Security)
	require.NotEmpty(t, spec.Paths["/api/v1/auditoria"].Get.Security)
}
