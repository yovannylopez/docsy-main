// Package container wires the dependencies for the TEMPLATE_MODULE module.
package container

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/domain/ports"
	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/infrastructure/repositories"
	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/transport/handlers"
	"github.com/yovannylopez/docsy-main/internal/TEMPLATE_MODULE/usecases"
)

// TemplateContainer holds all wired dependencies for TEMPLATE_MODULE.
type TemplateContainer struct {
	Repository    ports.TemplateRepository
	CreateUseCase ports.CreateTemplateService
	GetUseCase    ports.GetTemplateService
	ListUseCase   ports.ListTemplateService
	UpdateUseCase ports.UpdateTemplateService
	DeleteUseCase ports.DeleteTemplateService
	Handler       *handlers.TemplateHandler
}

// NewTemplateContainer creates and wires all TEMPLATE_MODULE dependencies.
func NewTemplateContainer(db *sqlx.DB) (*TemplateContainer, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}

	repo := repositories.NewTemplateRepositoryAdapter(db)

	createUC := usecases.NewCreateTemplateUseCase(repo)
	getUC := usecases.NewGetTemplateUseCase(repo)
	listUC := usecases.NewListTemplateUseCase(repo)
	updateUC := usecases.NewUpdateTemplateUseCase(repo)
	deleteUC := usecases.NewDeleteTemplateUseCase(repo)

	handler := handlers.NewTemplateHandler(createUC, getUC, listUC, updateUC, deleteUC)

	return &TemplateContainer{
		Repository:    repo,
		CreateUseCase: createUC,
		GetUseCase:    getUC,
		ListUseCase:   listUC,
		UpdateUseCase: updateUC,
		DeleteUseCase: deleteUC,
		Handler:       handler,
	}, nil
}
