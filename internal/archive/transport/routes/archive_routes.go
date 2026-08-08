package routes

import (
	"github.com/labstack/echo/v4"

	"github.com/yovannylopez/docsy-main/internal/archive/transport/handlers"
	authmw "github.com/yovannylopez/docsy-main/internal/auth/transport/middleware"
)

// RegisterArchiveRoutes registers JSON archive routes under the protected /api/v1 group.
func RegisterArchiveRoutes(g *echo.Group, handler *handlers.ArchiveHandler) {
	ag := g.Group("/archive")
	ag.GET("/workspaces/me", handler.GetMyWorkspace, authmw.RequirePermission("archive.read"))
	ag.GET("/workspaces", handler.ListWorkspaces, authmw.RequirePermission("archive.read"))
	ag.POST("/workspaces/household", handler.CreateHousehold, authmw.RequirePermission("archive.manage"))
	ag.GET("/workspaces/:id/members", handler.ListMembers, authmw.RequirePermission("archive.read"))
	ag.POST("/workspaces/:id/members", handler.InviteMember, authmw.RequirePermission("archive.manage"))
	ag.PATCH("/workspaces/:id/members/:userId", handler.UpdateMemberRole, authmw.RequirePermission("archive.manage"))
	ag.DELETE("/workspaces/:id/members/:userId", handler.RemoveMember, authmw.RequirePermission("archive.manage"))
	ag.GET("/categories", handler.ListCategories, authmw.RequirePermission("archive.read"))
	ag.GET("/documents", handler.ListDocuments, authmw.RequirePermission("archive.read"))
	ag.POST("/documents", handler.CreateDocument, authmw.RequirePermission("archive.write"))
	ag.GET("/documents/:id", handler.GetDocument, authmw.RequirePermission("archive.read"))
	ag.PATCH("/documents/:id", handler.UpdateDocument, authmw.RequirePermission("archive.write"))
	ag.POST("/documents/:id/archive", handler.ArchiveDocument, authmw.RequirePermission("archive.write"))
	ag.GET("/documents/:id/files", handler.ListDocumentFiles, authmw.RequirePermission("archive.read"))
	ag.POST("/documents/:id/files", handler.UploadDocumentFile, authmw.RequirePermission("archive.write"))
	ag.GET("/documents/:id/files/:fileId", handler.DownloadDocumentFile, authmw.RequirePermission("archive.read"))
	ag.DELETE("/documents/:id/files/:fileId", handler.DeleteDocumentFile, authmw.RequirePermission("archive.write"))
}
