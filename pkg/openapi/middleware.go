package openapi

import (
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/http_status"
)

// Middleware creates a middleware that serves GET (and other methods) at /openapi.json and /docs.
// Swagger UI loads assets from unpkg.com (requires network); in environments without CDN, use self-hosted or another UI.
func Middleware(generator *Generator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Serve OpenAPI specification at /openapi.json
			if c.Request().URL.Path == "/openapi.json" {
				spec, err := generator.Generate()
				if err != nil {
					return echo.NewHTTPError(http_status.InternalError.Code, "Error generating OpenAPI spec")
				}

				c.Response().Header().Set("Content-Type", constants.ContentTypeJSON)
				return c.JSONBlob(http_status.OK.Code, spec)
			}

			// Serve Swagger UI at /docs
			if c.Request().URL.Path == "/docs" {
				return c.HTML(http_status.OK.Code, generateSwaggerUI())
			}

			return next(c)
		}
	}
}

// generateSwaggerUI generates the HTML for Swagger UI
func generateSwaggerUI() string {
	return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>
`
}

// SetupOpenAPIRoutes registers the middleware that exposes /openapi.json and /docs.
// Does not duplicate echo.GET for the same routes (avoids double registration and divergent behaviors).
func SetupOpenAPIRoutes(e *echo.Echo, generator *Generator) {
	e.Use(Middleware(generator))
}
