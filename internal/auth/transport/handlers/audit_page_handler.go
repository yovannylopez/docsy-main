package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/auth/domain"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/entities"
	"github.com/yovannylopez/docsy-main/internal/auth/domain/ports"
	weblayout "github.com/yovannylopez/docsy-main/internal/shared/transport/web"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
)

const (
	auditListTitle        = "Auditoría"
	auditListSubtitle     = "Logs de actividad del sistema"
	msgAuditLoadError     = "No se pudieron cargar los logs de auditoría. Intenta de nuevo."
	shortUserIDDisplayLen = 8
)

// AuditFiltersView holds filter field values for the audit list page.
type AuditFiltersView struct {
	UserID   string
	Action   string
	Resource string
	Result   string
}

// AuditLogRowView is a single audit log row for templates.
type AuditLogRowView struct {
	CreatedAtFormatted string
	UserIDShort        string
	ActionLabel        string
	ResourceLabel      string
	ResultLabel        string
	ResultBadgeClass   string
	Message            string
	IPAddress          string
}

// AuditListPageData holds view data for the audit list page.
type AuditListPageData struct {
	weblayout.AppLayoutData
	AuditLogs       []AuditLogRowView
	Total           int
	Filters         AuditFiltersView
	Error           string
	Pagination      weblayout.PaginationData
	ActionOptions   []weblayout.SelectOption
	ResourceOptions []weblayout.SelectOption
	ResultOptions   []weblayout.SelectOption
}

// AuditPageHandler serves server-rendered audit pages.
type AuditPageHandler struct {
	listAuditLogsUC ports.ListAuditLogsUseCase
}

// NewAuditPageHandler creates an AuditPageHandler.
func NewAuditPageHandler(listAuditLogsUC ports.ListAuditLogsUseCase) *AuditPageHandler {
	return &AuditPageHandler{listAuditLogsUC: listAuditLogsUC}
}

// List renders the audit log list page or an HTMX table partial.
func (h *AuditPageHandler) List(c echo.Context) error {
	params, err := pagination.NewDefaultParser().ParseFromQuery(c.QueryParam("limit"), c.QueryParam("offset"))
	if err != nil {
		return h.renderAuditList(c, AuditListPageData{
			AppLayoutData:   weblayout.AppLayoutFromEcho(c, auditListTitle, auditListSubtitle, "/auditoria"),
			Error:           err.Error(),
			ActionOptions:   auditActionOptions(),
			ResourceOptions: auditResourceOptions(),
			ResultOptions:   auditResultOptions(),
		})
	}

	filtersView := AuditFiltersView{
		UserID:   strings.TrimSpace(c.QueryParam("user_id")),
		Action:   c.QueryParam("action"),
		Resource: c.QueryParam("resource"),
		Result:   c.QueryParam("result"),
	}

	filters, parseErr := buildAuditFilters(c, *params, filtersView)
	if parseErr != nil {
		return h.renderAuditList(c, AuditListPageData{
			AppLayoutData:   weblayout.AppLayoutFromEcho(c, auditListTitle, auditListSubtitle, "/auditoria"),
			Filters:         filtersView,
			Error:           parseErr.Error(),
			ActionOptions:   auditActionOptions(),
			ResourceOptions: auditResourceOptions(),
			ResultOptions:   auditResultOptions(),
		})
	}

	logs, total, execErr := h.listAuditLogsUC.Execute(c.Request().Context(), filters)
	if execErr != nil {
		return h.renderAuditList(c, AuditListPageData{
			AppLayoutData:   weblayout.AppLayoutFromEcho(c, auditListTitle, auditListSubtitle, "/auditoria"),
			Filters:         filtersView,
			Error:           msgAuditLoadError,
			ActionOptions:   auditActionOptions(),
			ResourceOptions: auditResourceOptions(),
			ResultOptions:   auditResultOptions(),
		})
	}

	data := AuditListPageData{
		AppLayoutData:   weblayout.AppLayoutFromEcho(c, auditListTitle, auditListSubtitle, "/auditoria"),
		AuditLogs:       mapAuditLogRows(logs),
		Total:           total,
		Filters:         filtersView,
		Pagination:      weblayout.NewPaginationData(params.Offset, params.Limit, total, "/auditoria", c.QueryParams()),
		ActionOptions:   auditActionOptions(),
		ResourceOptions: auditResourceOptions(),
		ResultOptions:   auditResultOptions(),
	}

	return h.renderAuditList(c, data)
}

func (h *AuditPageHandler) renderAuditList(c echo.Context, data AuditListPageData) error {
	if weblayout.IsHTMXRequest(c) && strings.Contains(c.Request().Header.Get("HX-Target"), "audit-table") {
		return c.Render(http.StatusOK, "partials/audit-table", data)
	}
	return c.Render(http.StatusOK, "audit/list", data)
}

func buildAuditFilters(c echo.Context, params pagination.Params, view AuditFiltersView) (*dtos.AuditLogFilters, error) {
	filters := &dtos.AuditLogFilters{
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	if view.UserID != "" {
		filters.UserID = &view.UserID
	}
	if view.Action != "" {
		filters.Action = &view.Action
	}
	if view.Resource != "" {
		filters.Resource = &view.Resource
	}
	if view.Result != "" {
		filters.Result = &view.Result
	}

	if startDateStr := c.QueryParam("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, fmt.Errorf("formato de start_date inválido: %w", err)
		}
		filters.StartDate = &startDate
	}

	if endDateStr := c.QueryParam("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, fmt.Errorf("formato de end_date inválido: %w", err)
		}
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filters.EndDate = &endDate
	}

	return filters, nil
}

func mapAuditLogRows(logs []entities.AuditLog) []AuditLogRowView {
	rows := make([]AuditLogRowView, 0, len(logs))
	for _, log := range logs {
		rows = append(rows, AuditLogRowView{
			CreatedAtFormatted: log.CreatedAt.Format("02/01/2006 15:04"),
			UserIDShort:        shortUserID(log.UserID),
			ActionLabel:        humanizeAuditAction(log.Action),
			ResourceLabel:      humanizeAuditResource(log.Resource),
			ResultLabel:        humanizeAuditResult(log.Result),
			ResultBadgeClass:   auditResultBadgeClass(log.Result),
			Message:            weblayout.DerefString(log.Message),
			IPAddress:          weblayout.DerefString(log.IPAddress),
		})
	}
	return rows
}

func shortUserID(userID *string) string {
	if userID == nil || *userID == "" {
		return "—"
	}
	id := *userID
	if len(id) > shortUserIDDisplayLen {
		return id[:shortUserIDDisplayLen] + "…"
	}
	return id
}

func humanizeAuditAction(action string) string {
	labels := map[string]string{
		domain.AuditActionCreate:               "Crear",
		domain.AuditActionUpdate:               "Actualizar",
		domain.AuditActionDelete:               "Eliminar",
		domain.AuditActionRead:                 "Leer",
		domain.AuditActionUserLoginAttempt:     "Intento de login",
		domain.AuditActionUserLogout:           "Cierre de sesión",
		domain.AuditActionUserCreated:          "Usuario creado",
		domain.AuditActionPasswordChanged:      "Contraseña cambiada",
		domain.AuditActionPasswordChangeFailed: "Cambio de contraseña fallido",
	}
	if label, ok := labels[action]; ok {
		return label
	}
	return action
}

func humanizeAuditResource(resource *string) string {
	if resource == nil || *resource == "" {
		return "—"
	}
	return *resource
}

func humanizeAuditResult(result string) string {
	switch result {
	case domain.AuditResultSuccess:
		return "Éxito"
	case domain.AuditResultFailure:
		return "Fallo"
	case domain.AuditResultError:
		return "Error"
	default:
		return result
	}
}

func auditResultBadgeClass(result string) string {
	switch result {
	case domain.AuditResultSuccess:
		return "bg-green-100 text-green-800 dark:bg-green-950/40 dark:text-green-200"
	case domain.AuditResultFailure:
		return "bg-red-100 text-red-800 dark:bg-red-950/40 dark:text-red-200"
	case domain.AuditResultError:
		return "bg-amber-100 text-amber-800 dark:bg-amber-950/40 dark:text-amber-200"
	default:
		return "bg-muted text-muted-foreground"
	}
}

func auditActionOptions() []weblayout.SelectOption {
	return []weblayout.SelectOption{
		{Value: "", Label: "Todas"},
		{Value: domain.AuditActionCreate, Label: "Crear"},
		{Value: domain.AuditActionUpdate, Label: "Actualizar"},
		{Value: domain.AuditActionDelete, Label: "Eliminar"},
		{Value: domain.AuditActionRead, Label: "Leer"},
		{Value: domain.AuditActionUserLoginAttempt, Label: "Intento de login"},
		{Value: domain.AuditActionUserLogout, Label: "Cierre de sesión"},
		{Value: domain.AuditActionUserCreated, Label: "Usuario creado"},
		{Value: domain.AuditActionPasswordChanged, Label: "Contraseña cambiada"},
	}
}

func auditResourceOptions() []weblayout.SelectOption {
	return []weblayout.SelectOption{
		{Value: "", Label: "Todos"},
		{Value: "users", Label: "Usuarios"},
		{Value: "auth", Label: "Autenticación"},
		{Value: "sessions", Label: "Sesiones"},
	}
}

func auditResultOptions() []weblayout.SelectOption {
	return []weblayout.SelectOption{
		{Value: "", Label: "Todos"},
		{Value: domain.AuditResultSuccess, Label: "Éxito"},
		{Value: domain.AuditResultFailure, Label: "Fallo"},
		{Value: domain.AuditResultError, Label: "Error"},
	}
}
