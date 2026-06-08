package templates

import (
	"bytes"
	"net/url"
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

func TestRenderer_RenderConfirmationModal(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	data := weblayout.ConfirmationModalData{
		ModalID:       "delete-user-modal",
		Title:         "Eliminar usuario",
		Message:       "¿Confirma que desea eliminar este usuario?",
		ConfirmText:   "Eliminar",
		ConfirmURL:    "/usuarios/1/delete",
		ConfirmMethod: "delete",
	}
	err = renderer.Render(&buf, "components/confirmation-modal", data, nil)
	require.NoError(t, err)

	html := buf.String()
	require.Contains(t, html, `id="delete-user-modal"`)
	require.Contains(t, html, "Eliminar usuario")
	require.Contains(t, html, `hx-delete="/usuarios/1/delete"`)
	require.Contains(t, html, `closeModal('delete-user-modal')`)
	require.Contains(t, html, "bg-card")
}

func TestRenderer_RenderSuccessModal(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	data := weblayout.SuccessModalData{
		ModalID: "user-created-modal",
		Title:   "Usuario creado",
		Summary: "El usuario se registró correctamente.",
		Sections: []weblayout.SuccessModalSection{
			{
				Title: "Detalle",
				Rows: []weblayout.SuccessModalRow{
					{Label: "Email", Value: "ana@example.com"},
				},
			},
		},
		PrimaryActionURL:   "/usuarios",
		PrimaryActionLabel: "Ver listado",
	}
	err = renderer.Render(&buf, "components/success-modal", data, nil)
	require.NoError(t, err)

	html := buf.String()
	require.Contains(t, html, "Usuario creado")
	require.Contains(t, html, "ana@example.com")
	require.Contains(t, html, `href="/usuarios"`)
	require.Contains(t, html, "Ver listado")
}

func TestRenderer_RenderModalShell(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	data := weblayout.ModalShellData{
		ModalID:       "warn-modal",
		ModalTitle:    "Atención",
		ModalMessage:  "Esta acción no se puede deshacer.",
		ConfirmURL:    "/usuarios/1",
		ConfirmMethod: "post",
		HeaderTone:    "danger",
	}
	err = renderer.Render(&buf, "components/modal-shell", data, nil)
	require.NoError(t, err)

	html := buf.String()
	require.Contains(t, html, "Atención")
	require.Contains(t, html, "bg-destructive")
	require.Contains(t, html, `hx-post="/usuarios/1"`)
}

func TestRenderer_RenderFormField(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	data := weblayout.FormFieldPageData{
		Field: weblayout.FormFieldData{
			ID:      "email",
			Name:    "email",
			Label:   "Correo electrónico",
			Type:    "email",
			Value:   "ana@example.com",
			Invalid: true,
			Error:   "El correo ya está registrado",
		},
	}
	err = renderer.Render(&buf, "partials/form-field", data, nil)
	require.NoError(t, err)

	html := buf.String()
	require.Contains(t, html, `id="email"`)
	require.Contains(t, html, "Correo electrónico")
	require.Contains(t, html, "is__invalid-input")
	require.Contains(t, html, "El correo ya está registrado")
}

func TestRenderer_RenderTablePagination(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	data := weblayout.NewPaginationData(0, 10, 42, "/usuarios", url.Values{"q": []string{"ana"}})
	err = renderer.Render(&buf, "partials/table-pagination", data, nil)
	require.NoError(t, err)

	html := buf.String()
	require.Contains(t, html, "Mostrando")
	require.Contains(t, html, "42")
	require.Contains(t, html, `offset=10`)
	require.Contains(t, html, "q=ana")
}

func TestRenderer_RenderAlertSuccess(t *testing.T) {
	renderer, err := NewRenderer()
	require.NoError(t, err)

	var buf bytes.Buffer
	data := weblayout.AlertSuccessData{
		SuccessTitle:   "Operación exitosa",
		SuccessMessage: "Los cambios se guardaron correctamente.",
	}
	err = renderer.Render(&buf, "partials/alert-success", data, nil)
	require.NoError(t, err)

	html := buf.String()
	require.Contains(t, html, "Operación exitosa")
	require.Contains(t, html, "Los cambios se guardaron correctamente.")
	require.Contains(t, html, "bg-green-50")
}
