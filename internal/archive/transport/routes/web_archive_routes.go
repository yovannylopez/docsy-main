package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/archive/transport/handlers"
	authmiddleware "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
)

// WebArchiveRoutes registers server-rendered archive routes.
type WebArchiveRoutes struct {
	pageHandler *handlers.ArchivePageHandler
	webAuthMW   *authmiddleware.WebAuthMiddleware
	extraMW     []echo.MiddlewareFunc
}

// NewWebArchiveRoutes creates WebArchiveRoutes.
func NewWebArchiveRoutes(
	pageHandler *handlers.ArchivePageHandler,
	webAuthMW *authmiddleware.WebAuthMiddleware,
	extraMW ...echo.MiddlewareFunc,
) *WebArchiveRoutes {
	return &WebArchiveRoutes{
		pageHandler: pageHandler,
		webAuthMW:   webAuthMW,
		extraMW:     extraMW,
	}
}

// Setup registers authenticated web archive routes.
func (wr *WebArchiveRoutes) Setup(e *echo.Echo) {
	g := e.Group("")
	g.Use(wr.webAuthMW.RequireAuth())
	for _, mw := range wr.extraMW {
		g.Use(mw)
	}
	g.GET("/archivo", wr.pageHandler.ShowArchive, wr.webAuthMW.RequirePermission("archive.read"))
	g.GET("/archivo/hogares/nuevo", wr.pageHandler.ShowCreateHousehold, wr.webAuthMW.RequirePermission("archive.manage"))
	g.POST("/archivo/hogares/nuevo", wr.pageHandler.SubmitCreateHousehold, wr.webAuthMW.RequirePermission("archive.manage"))
	g.GET("/archivo/hogares/:id/miembros", wr.pageHandler.ShowMembers, wr.webAuthMW.RequirePermission("archive.read"))
	g.POST("/archivo/hogares/:id/miembros", wr.pageHandler.InviteMember, wr.webAuthMW.RequirePermission("archive.manage"))
	g.POST("/archivo/hogares/:id/miembros/:userId/eliminar", wr.pageHandler.RemoveMember, wr.webAuthMW.RequirePermission("archive.manage"))
	g.GET("/archivo/documentos", wr.pageHandler.ListDocuments, wr.webAuthMW.RequirePermission("archive.read"))
	g.GET("/archivo/documentos/nuevo", wr.pageHandler.ShowCreate, wr.webAuthMW.RequirePermission("archive.write"))
	g.POST("/archivo/documentos/nuevo", wr.pageHandler.SubmitCreate, wr.webAuthMW.RequirePermission("archive.write"))
	g.GET("/archivo/documentos/:id/editar", wr.pageHandler.ShowEdit, wr.webAuthMW.RequirePermission("archive.write"))
	g.POST("/archivo/documentos/:id/editar", wr.pageHandler.SubmitEdit, wr.webAuthMW.RequirePermission("archive.write"))
	g.POST("/archivo/documentos/:id/archivos", wr.pageHandler.UploadDocumentFile, wr.webAuthMW.RequirePermission("archive.write"))
	g.GET("/archivo/documentos/:id/archivos/:fileId", wr.pageHandler.DownloadDocumentFile, wr.webAuthMW.RequirePermission("archive.read"))
	g.POST("/archivo/documentos/:id/archivos/:fileId/eliminar", wr.pageHandler.DeleteDocumentFile, wr.webAuthMW.RequirePermission("archive.write"))
}
