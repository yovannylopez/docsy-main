package templates

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

const errTemplatesNotFound = "no templates folder 'web/templates' found. " +
	"Configure the TEMPLATES_PATH environment variable or run from the project root"

// Renderer implements echo.Renderer for server-rendered HTML templates.
type Renderer struct {
	templates *template.Template
}

// NewRenderer loads all .gohtml templates from web/templates/ and returns an Echo renderer.
func NewRenderer() (*Renderer, error) {
	templatesPath, err := resolveTemplatesPath()
	if err != nil {
		return nil, err
	}

	funcMap := template.FuncMap{
		"assetURL": func(path string) string {
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			if strings.HasPrefix(path, "/assets/") {
				return "/static" + path
			}
			if strings.HasPrefix(path, "/js/") || strings.HasPrefix(path, "/css/") {
				return "/static" + path
			}
			return path
		},
		"hasPrefix": strings.HasPrefix,
	}

	tmpl := template.New("").Funcs(funcMap)

	err = filepath.Walk(templatesPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || !strings.HasSuffix(path, ".gohtml") {
			return nil
		}

		rel, err := filepath.Rel(templatesPath, path)
		if err != nil {
			return fmt.Errorf("resolve template path %q: %w", path, err)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read template %q: %w", path, err)
		}

		name := strings.TrimSuffix(filepath.ToSlash(rel), ".gohtml")
		if _, err := tmpl.New(name).Parse(string(content)); err != nil {
			return fmt.Errorf("parse template %q: %w", name, err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load templates from %s: %w", templatesPath, err)
	}

	return &Renderer{templates: tmpl}, nil
}

// Render renders a named template with the given data.
func (r *Renderer) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	if r.templates == nil {
		return fmt.Errorf("templates not initialized")
	}

	if err := r.templates.ExecuteTemplate(w, name, data); err != nil {
		return fmt.Errorf("execute template %q: %w", name, err)
	}

	return nil
}

// TemplatesPath returns the resolved templates directory (for tests and diagnostics).
func TemplatesPath() (string, error) {
	return resolveTemplatesPath()
}

func resolveTemplatesPath() (string, error) {
	if envPath := os.Getenv("TEMPLATES_PATH"); envPath != "" {
		if abs, err := filepath.Abs(envPath); err == nil {
			if info, err2 := os.Stat(abs); err2 == nil && info.IsDir() {
				return abs, nil
			}
		}
	}

	candidates := []string{
		"web/templates",
		"../web/templates",
		"../../web/templates",
	}
	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if info, err2 := os.Stat(abs); err2 == nil && info.IsDir() {
				return abs, nil
			}
		}
	}

	if wd, err := os.Getwd(); err == nil {
		dir := wd
		for i := 0; i < 5; i++ {
			candidate := filepath.Join(dir, "web", "templates")
			if info, err2 := os.Stat(candidate); err2 == nil && info.IsDir() {
				if abs, err3 := filepath.Abs(candidate); err3 == nil {
					return abs, nil
				}

				return candidate, nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return "", fmt.Errorf("%s", errTemplatesNotFound)
}
