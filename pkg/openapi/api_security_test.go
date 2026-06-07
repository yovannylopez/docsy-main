package openapi_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

func TestApplySecurityAndRBACDocumentation_PublicVsProtected(t *testing.T) {
	gen := openapi.NewGenerator("T", "D", "1.0")
	spec := gen.GetSpec()

	spec.Paths["/api/public/health"] = openapi.PathItem{
		Get: &openapi.Operation{Responses: map[string]openapi.Response{"200": {Description: "ok"}}},
	}
	spec.Paths["/api/v1/auth/login"] = openapi.PathItem{
		Post: &openapi.Operation{Responses: map[string]openapi.Response{"200": {Description: "ok"}}},
	}
	spec.Paths["/api/v1/users"] = openapi.PathItem{
		Get: &openapi.Operation{Responses: map[string]openapi.Response{"200": {Description: "ok"}}},
	}

	openapi.RegisterStandardErrorResponseSchema(gen)
	openapi.ApplySecurityAndRBACDocumentation(gen)

	require.Contains(t, spec.Components.SecuritySchemes, openapi.SecuritySchemeBearer)
	require.Empty(t, spec.Paths["/api/public/health"].Get.Security)
	require.Empty(t, spec.Paths["/api/v1/auth/login"].Post.Security)
	require.Len(t, spec.Paths["/api/v1/users"].Get.Security, 1)
	require.Contains(t, spec.Paths["/api/v1/users"].Get.Security[0], openapi.SecuritySchemeBearer)

	_, has403 := spec.Paths["/api/v1/users"].Get.Responses["403"]
	require.True(t, has403)
}
