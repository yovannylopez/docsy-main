package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	weblayout "github.com/yovannylopez/docsy-main/internal/shared/transport/web"
	"github.com/yovannylopez/docsy-main/internal/users/domain/dtos"
	domainerrors "github.com/yovannylopez/docsy-main/internal/users/domain/errors"
	"github.com/yovannylopez/docsy-main/internal/users/usecases"
	"github.com/yovannylopez/docsy-main/pkg/constants"
	"github.com/yovannylopez/docsy-main/pkg/pagination"
)

const (
	usersListTitle    = "Usuarios"
	usersListSubtitle = "Gestión de usuarios del sistema"
	usersCreateTitle  = "Crear usuario"
	usersEditTitle    = "Editar usuario"

	msgUsersLoadError        = "No se pudieron cargar los usuarios. Intenta de nuevo."
	msgUserCreateError       = "No se pudo crear el usuario. Verifica los datos e intenta de nuevo."
	msgUserUpdateError       = "No se pudo actualizar el usuario. Verifica los datos e intenta de nuevo."
	msgUserNotFound          = "El usuario solicitado no existe."
	msgConfirmFieldsMismatch = "Las contraseñas no coinciden."
	msgUserCreated           = "Usuario creado correctamente."
	msgUserUpdated           = "Usuario actualizado correctamente."
)

// UsersListPageData holds view data for the users list page.
type UsersListPageData struct {
	weblayout.AppLayoutData
	Users      []dtos.UserListResponse
	Total      int
	Query      string
	Error      string
	Success    string
	Pagination weblayout.PaginationData
}

// UserCreatePageData holds view data for the user creation page.
type UserCreatePageData struct {
	weblayout.AppLayoutData
	Form    UserCreateForm
	Error   string
	Success string
	Roles   []string
	IDTypes []weblayout.SelectOption
}

// UserCreateForm holds create form field values and validation errors.
type UserCreateForm struct {
	Email                string
	Password             string
	ConfirmPassword      string
	FirstName            string
	LastName             string
	RoleName             string
	IdentificationType   string
	IdentificationNumber string
	Phone                string
	EmailError           string
	PasswordError        string
	GeneralError         string
}

// UserEditPageData holds view data for the user edit page.
type UserEditPageData struct {
	weblayout.AppLayoutData
	User    UserEditForm
	Error   string
	Success string
	IDTypes []weblayout.SelectOption
}

// UserEditForm holds edit form values.
type UserEditForm struct {
	ID                   string
	Email                string
	FirstName            string
	LastName             string
	IdentificationType   string
	IdentificationNumber string
	Phone                string
	IsActive             bool
	IsVerified           bool
	MFAEnabled           bool
}

// UsersPageHandler serves server-rendered users pages (HTMX/HTML).
type UsersPageHandler struct {
	getUsersUseCase    *usecases.GetUsersUseCase
	createUsersUseCase *usecases.CreateUsersUseCase
	updateUserUseCase  *usecases.UpdateUserUseCase
	searchUsersUseCase *usecases.SearchUsersUseCase
	getUserByIDUseCase *usecases.GetUserByIDUseCase
}

// NewUsersPageHandler creates a UsersPageHandler.
func NewUsersPageHandler(
	getUsersUC *usecases.GetUsersUseCase,
	createUsersUC *usecases.CreateUsersUseCase,
	updateUserUC *usecases.UpdateUserUseCase,
	searchUsersUC *usecases.SearchUsersUseCase,
	getUserByIDUC *usecases.GetUserByIDUseCase,
) *UsersPageHandler {
	return &UsersPageHandler{
		getUsersUseCase:    getUsersUC,
		createUsersUseCase: createUsersUC,
		updateUserUseCase:  updateUserUC,
		searchUsersUseCase: searchUsersUC,
		getUserByIDUseCase: getUserByIDUC,
	}
}

// ListUsers renders the users list page.
func (h *UsersPageHandler) ListUsers(c echo.Context) error {
	params, err := usersPaginationParser().ParseFromQuery(c.QueryParam("limit"), c.QueryParam("offset"))
	if err != nil {
		return h.renderUsersList(c, UsersListPageData{
			AppLayoutData: weblayout.AppLayoutFromEcho(c, usersListTitle, usersListSubtitle, "/usuarios"),
			Error:         err.Error(),
		})
	}

	query := strings.TrimSpace(c.QueryParam("q"))
	var users []dtos.UserListResponse
	var total int

	if query != "" {
		resp, searchErr := h.searchUsersUseCase.Execute(c.Request().Context(), &dtos.SearchUsersRequest{
			Q:      query,
			Limit:  params.Limit,
			Offset: params.Offset,
		})
		if searchErr != nil {
			return h.renderUsersList(c, UsersListPageData{
				AppLayoutData: weblayout.AppLayoutFromEcho(c, usersListTitle, usersListSubtitle, "/usuarios"),
				Query:         query,
				Error:         msgUsersLoadError,
			})
		}
		users = resp.Usuarios
		total = resp.Total
	} else {
		resp, listErr := h.getUsersUseCase.Execute(c.Request().Context(), &usecases.GetUsersRequest{
			Limit:  params.Limit,
			Offset: params.Offset,
		})
		if listErr != nil {
			return h.renderUsersList(c, UsersListPageData{
				AppLayoutData: weblayout.AppLayoutFromEcho(c, usersListTitle, usersListSubtitle, "/usuarios"),
				Error:         msgUsersLoadError,
			})
		}
		users = resp.Usuarios
		total = resp.Total
	}

	data := UsersListPageData{
		AppLayoutData: weblayout.AppLayoutFromEcho(c, usersListTitle, usersListSubtitle, "/usuarios"),
		Users:         users,
		Total:         total,
		Query:         query,
		Success:       flashMessage(c, "created", msgUserCreated, "updated", msgUserUpdated),
		Pagination:    weblayout.NewPaginationData(params.Offset, params.Limit, total, "/usuarios", c.QueryParams()),
	}

	return h.renderUsersList(c, data)
}

// ShowCreate renders the user creation form.
func (h *UsersPageHandler) ShowCreate(c echo.Context) error {
	data := UserCreatePageData{
		AppLayoutData: weblayout.AppLayoutFromEcho(c, usersCreateTitle, "Registrar un nuevo usuario", "/usuarios/nuevo"),
		Roles:         defaultRoleOptions(),
		IDTypes:       identificationOptions(),
		Form: UserCreateForm{
			RoleName:           "user",
			IdentificationType: "cc",
		},
	}
	return c.Render(http.StatusOK, "users/create", data)
}

// SubmitCreate handles user creation from an HTML form.
func (h *UsersPageHandler) SubmitCreate(c echo.Context) error {
	form := bindUserCreateForm(c)

	if form.Password != form.ConfirmPassword {
		form.PasswordError = msgConfirmFieldsMismatch
		return h.renderCreateWithForm(c, form, "")
	}

	idType := form.IdentificationType
	idNumber := strings.TrimSpace(form.IdentificationNumber)
	phone := strings.TrimSpace(form.Phone)

	req := dtos.CreateUsersRequest{
		Users: []dtos.CreateUserRequest{{
			Email:                strings.TrimSpace(form.Email),
			Password:             form.Password,
			FirstName:            strings.TrimSpace(form.FirstName),
			LastName:             strings.TrimSpace(form.LastName),
			RoleName:             form.RoleName,
			IdentificationType:   &idType,
			IdentificationNumber: optionalString(idNumber),
			Phone:                optionalString(phone),
		}},
	}

	resp, err := h.createUsersUseCase.Execute(c.Request().Context(), &req, weblayout.CurrentUserID(c))
	if err != nil {
		form.GeneralError = mapCreateError(err)
		return h.renderCreateWithForm(c, form, form.GeneralError)
	}

	if resp.TotalCreated == 0 {
		if len(resp.Errors) > 0 {
			form.GeneralError = resp.Errors[0].Error
			form.EmailError = resp.Errors[0].Error
		} else {
			form.GeneralError = msgUserCreateError
		}
		return h.renderCreateWithForm(c, form, form.GeneralError)
	}

	if weblayout.IsHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", "/usuarios?created=1")
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusFound, "/usuarios?created=1")
}

// ShowEdit renders the user edit form.
func (h *UsersPageHandler) ShowEdit(c echo.Context) error {
	userID := c.Param("id")
	user, err := h.getUserByIDUseCase.Execute(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, usecases.ErrUserNotFound) {
			return c.Render(http.StatusNotFound, "forbidden", weblayout.AppLayoutFromEcho(
				c, "Usuario no encontrado", msgUserNotFound, "/usuarios",
			))
		}
		return c.Render(http.StatusInternalServerError, "users/edit", UserEditPageData{
			AppLayoutData: weblayout.AppLayoutFromEcho(c, usersEditTitle, "Modificar datos del usuario", "/usuarios/"+userID+"/editar"),
			Error:         msgUsersLoadError,
			IDTypes:       identificationOptions(),
		})
	}

	data := UserEditPageData{
		AppLayoutData: weblayout.AppLayoutFromEcho(c, usersEditTitle, "Modificar datos del usuario", "/usuarios/"+userID+"/editar"),
		IDTypes:       identificationOptions(),
		User: UserEditForm{
			ID:                   user.ID,
			Email:                user.Email,
			FirstName:            user.FirstName,
			LastName:             user.LastName,
			IdentificationType:   weblayout.DerefString(user.IdentificationType),
			IdentificationNumber: weblayout.DerefString(user.IdentificationNumber),
			Phone:                weblayout.DerefString(user.Phone),
			IsActive:             user.IsActive,
			IsVerified:           user.IsVerified,
			MFAEnabled:           user.MFAEnabled,
		},
		Success: flashMessage(c, "updated", msgUserUpdated, "", ""),
	}
	return c.Render(http.StatusOK, "users/edit", data)
}

// SubmitEdit handles user update from an HTML form.
func (h *UsersPageHandler) SubmitEdit(c echo.Context) error {
	userID := c.Param("id")
	form := bindUserEditForm(c, userID)

	email := strings.TrimSpace(form.Email)
	firstName := strings.TrimSpace(form.FirstName)
	lastName := strings.TrimSpace(form.LastName)
	idType := form.IdentificationType
	idNumber := strings.TrimSpace(form.IdentificationNumber)
	phone := strings.TrimSpace(form.Phone)
	isActive := parseCheckbox(c, "is_active")
	isVerified := parseCheckbox(c, "is_verified")
	mfaEnabled := parseCheckbox(c, "mfa_enabled")

	req := &dtos.UpdateUserRequest{
		Email:                &email,
		FirstName:            &firstName,
		LastName:             &lastName,
		IdentificationType:   &idType,
		IdentificationNumber: optionalString(idNumber),
		Phone:                optionalString(phone),
		IsActive:             &isActive,
		IsVerified:           &isVerified,
		MFAEnabled:           &mfaEnabled,
	}

	_, err := h.updateUserUseCase.Execute(c.Request().Context(), userID, req, weblayout.CurrentUserID(c))
	if err != nil {
		msg := msgUserUpdateError
		if errors.Is(err, usecases.ErrUserNotFound) {
			msg = msgUserNotFound
		} else if errors.Is(err, domainerrors.ErrEmailAlreadyExists) {
			msg = "El correo electrónico ya está en uso."
		}
		data := UserEditPageData{
			AppLayoutData: weblayout.AppLayoutFromEcho(c, usersEditTitle, "Modificar datos del usuario", "/usuarios/"+userID+"/editar"),
			User:          form,
			Error:         msg,
			IDTypes:       identificationOptions(),
		}
		return c.Render(http.StatusUnprocessableEntity, "users/edit", data)
	}

	redirectURL := fmt.Sprintf("/usuarios/%s/editar?updated=1", userID)
	if weblayout.IsHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", redirectURL)
		return c.NoContent(http.StatusOK)
	}
	return c.Redirect(http.StatusFound, redirectURL)
}

func (h *UsersPageHandler) renderUsersList(c echo.Context, data UsersListPageData) error {
	if weblayout.IsHTMXRequest(c) && strings.Contains(c.Request().Header.Get("HX-Target"), "users-table") {
		return c.Render(http.StatusOK, "partials/users-table", data)
	}
	return c.Render(http.StatusOK, "users/list", data)
}

func (h *UsersPageHandler) renderCreateWithForm(c echo.Context, form UserCreateForm, errMsg string) error {
	data := UserCreatePageData{
		AppLayoutData: weblayout.AppLayoutFromEcho(c, usersCreateTitle, "Registrar un nuevo usuario", "/usuarios/nuevo"),
		Form:          form,
		Error:         errMsg,
		Roles:         defaultRoleOptions(),
		IDTypes:       identificationOptions(),
	}
	status := http.StatusOK
	if errMsg != "" {
		status = http.StatusUnprocessableEntity
	}
	return c.Render(status, "users/create", data)
}

func bindUserCreateForm(c echo.Context) UserCreateForm {
	return UserCreateForm{
		Email:                c.FormValue("email"),
		Password:             c.FormValue("password"),
		ConfirmPassword:      c.FormValue("confirm_password"),
		FirstName:            c.FormValue("primer_nombre"),
		LastName:             c.FormValue("segundo_nombre"),
		RoleName:             c.FormValue("nombre_rol"),
		IdentificationType:   c.FormValue("tipo_identificacion"),
		IdentificationNumber: c.FormValue("numero_identificacion"),
		Phone:                c.FormValue("telefono"),
	}
}

func bindUserEditForm(c echo.Context, userID string) UserEditForm {
	return UserEditForm{
		ID:                   userID,
		Email:                c.FormValue("email"),
		FirstName:            c.FormValue("first_name"),
		LastName:             c.FormValue("last_name"),
		IdentificationType:   c.FormValue("identification_type"),
		IdentificationNumber: c.FormValue("identification_number"),
		Phone:                c.FormValue("phone"),
		IsActive:             parseCheckbox(c, "is_active"),
		IsVerified:           parseCheckbox(c, "is_verified"),
		MFAEnabled:           parseCheckbox(c, "mfa_enabled"),
	}
}

func parseCheckbox(c echo.Context, name string) bool {
	v := c.FormValue(name)
	return v == "1" || v == "on" || v == "true"
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func usersPaginationParser() *pagination.Parser {
	return pagination.NewParser(pagination.Config{
		DefaultLimit: 10, //nolint:mnd
		MaxLimit:     constants.MaxUsersLimit,
		MinLimit:     1,
	})
}

func flashMessage(c echo.Context, createdKey, createdMsg, updatedKey, updatedMsg string) string {
	if createdKey != "" && c.QueryParam(createdKey) == "1" {
		return createdMsg
	}
	if updatedKey != "" && c.QueryParam(updatedKey) == "1" {
		return updatedMsg
	}
	return ""
}

func mapCreateError(err error) string {
	switch {
	case errors.Is(err, domainerrors.ErrEmailRequired),
		errors.Is(err, domainerrors.ErrPasswordRequired),
		errors.Is(err, domainerrors.ErrFirstNameRequired),
		errors.Is(err, domainerrors.ErrLastNameRequired),
		errors.Is(err, domainerrors.ErrRoleNameRequired),
		errors.Is(err, domainerrors.ErrPasswordTooShort):
		return err.Error()
	default:
		return msgUserCreateError
	}
}

func defaultRoleOptions() []string {
	return []string{"user", "viewer", "super_admin"}
}

func identificationOptions() []weblayout.SelectOption {
	return []weblayout.SelectOption{
		{Value: "cc", Label: "Cédula de ciudadanía"},
		{Value: "ce", Label: "Cédula de extranjería"},
		{Value: "pa", Label: "Pasaporte"},
		{Value: "nit", Label: "NIT"},
		{Value: "rut", Label: "RUT"},
	}
}
