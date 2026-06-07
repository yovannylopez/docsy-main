package openapi

import "strings"

// SecuritySchemeBearer is the scheme name in components.securitySchemes (Authorization: Bearer).
const SecuritySchemeBearer = "bearerAuth"

// ApplySecurityAndRBACDocumentation registers the HTTP Bearer scheme (JWT), assigns security per operation
// (public and anonymous auth without JWT; the rest with bearerAuth) and documents 403 RBAC with ErrorResponse.
// Must be executed after RegisterStandardErrorResponseSchema so that the ErrorResponse schema exists.
func ApplySecurityAndRBACDocumentation(g *Generator) {
	spec := g.GetSpec()

	g.AddSecurityScheme(SecuritySchemeBearer, &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description: "Access JWT issued on login. Header: Authorization: Bearer <access_token>. " +
			"Endpoints under /api/v1 (except public and login/refresh) require RBAC permissions.",
	})

	if spec.Info.Description != "" && !strings.Contains(spec.Info.Description, "RBAC permissions") {
		spec.Info.Description += "\n\nMost routes in /api/v1 require JWT Bearer and role-based permissions. " +
			"403 indicates valid authentication but no permission for the operation."
	}

	for path, item := range spec.Paths {
		apply := func(method string, op *Operation) {
			if op == nil {
				return
			}
			m := strings.ToLower(method)
			if isAnonymousAuthOrPublic(path, m) {
				op.Security = []map[string][]string{}
			} else {
				op.Security = []map[string][]string{{SecuritySchemeBearer: {}}}
			}
			if !isAnonymousAuthOrPublic(path, m) {
				mergeRBAC403Response(op)
			}
		}
		apply("get", item.Get)
		apply("post", item.Post)
		apply("put", item.Put)
		apply("patch", item.Patch)
		apply("delete", item.Delete)
		apply("options", item.Options)
		apply("head", item.Head)
		apply("trace", item.Trace)
	}
}

func isAnonymousAuthOrPublic(path, method string) bool {
	if strings.HasPrefix(path, "/api/public/") {
		return true
	}
	if method != "post" {
		return false
	}
	switch path {
	case "/api/v1/auth/login", "/api/v1/auth/refresh":
		return true
	default:
		return false
	}
}

func mergeRBAC403Response(op *Operation) {
	if op.Responses == nil {
		op.Responses = map[string]Response{}
	}
	forbidden := Response{
		Description: "Forbidden: valid token but no RBAC permission for this operation",
		Content:     JSONErrorRefContent(nil),
	}
	existing, ok := op.Responses["403"]
	if !ok {
		op.Responses["403"] = forbidden
		return
	}
	if len(existing.Content) > 0 {
		return
	}
	if existing.Description == "" || existing.Description == "Forbidden" {
		existing.Description = forbidden.Description
	}
	existing.Content = forbidden.Content
	op.Responses["403"] = existing
}
