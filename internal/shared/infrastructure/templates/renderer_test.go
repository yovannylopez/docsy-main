package templates

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	weblayout "github.com/yovannylopez/docsy-main/internal/shared/transport/web"
)

func TestNewRenderer_ParsesTemplates(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)
	require.NotNil(t, renderer)
}

func TestRenderer_RenderLoginTemplate(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	data := map[string]string{
		"Title":    "Hola de nuevo",
		"Subtitle": "Ingresa tus credenciales",
	}
	err = renderer.Render(&buf, "login", data, nil)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "Hola de nuevo")
	require.Contains(t, buf.String(), `hx-post="/login"`)
}

func TestTemplatesPath_ResolvesFromProjectRoot(t *testing.T) {
	path, err := TemplatesPath()
	require.NoError(t, err)
	require.Contains(t, path, "web/templates")
}

func TestRenderer_RenderProfileMenuTemplate(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	data := map[string]string{
		"UserName": "Ana García",
	}
	err = renderer.Render(&buf, "partials/profile-menu", data, nil)
	require.NoError(t, err)

	html := buf.String()
	require.Contains(t, html, "Ana García")
	require.Contains(t, html, "Perfil")
	require.Contains(t, html, "Configuración")
	require.Contains(t, html, "Variantes de color")
	require.Contains(t, html, `setThemeColor('green')`)
	require.Contains(t, html, `aria-label="Verde"`)
	require.Contains(t, html, `h-6 w-6`)
	require.NotContains(t, html, ">Verde</span>")
	require.NotContains(t, html, "Distribución")
	require.Contains(t, html, `hx-post="/logout"`)
}

func TestRenderer_RenderHomeIncludesProfileMenu(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	data := weblayout.NewAppLayoutData("Inicio", "Panel principal", "Ana García", "/")
	err = renderer.Render(&buf, "home", data, nil)
	require.NoError(t, err)

	html := buf.String()
	require.Contains(t, html, "Variantes de color")
	require.Contains(t, html, "Cerrar sesión")
}
