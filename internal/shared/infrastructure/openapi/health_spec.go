// Package openapi documents public health endpoints (they do not use the standard business JSON envelope).
package openapi

import (
	"github.com/yovannylopez/docsy-main/pkg/openapi"
)

// SetupHealthSpec registers health and ready endpoints in the OpenAPI specification.
func SetupHealthSpec(generator *openapi.Generator) {
	// Configure specific operations
	setupHealthCheckOperation(generator)
	setupReadyCheckOperation(generator)
}

// setupHealthCheckOperation configures the health check operation
func setupHealthCheckOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()

	if pathItem, exists := spec.Paths["/api/public/health"]; exists && pathItem.Get != nil {
		operation := pathItem.Get

		operation.Summary = "Health Check"
		operation.Description = "Checks the general health status of the service"
		operation.OperationID = "healthCheck"
		operation.Tags = []string{"health"}

		operation.Responses = map[string]openapi.Response{
			"200": {
				Description: "Service running correctly",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Type: "object",
							Properties: map[string]*openapi.Schema{
								"status": {
									Type: "string",
								},
								"timestamp": {
									Type: "string",
								},
								"version": {
									Type: "string",
								},
							},
						},
					},
				},
			},
			"503": {
				Description: "Service unavailable",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Type: "object",
							Properties: map[string]*openapi.Schema{
								"status": {
									Type: "string",
								},
								"error": {
									Type: "string",
								},
							},
						},
					},
				},
			},
		}

		spec.Paths["/api/public/health"] = pathItem
	}
}

// setupReadyCheckOperation configures the ready check operation
func setupReadyCheckOperation(generator *openapi.Generator) {
	spec := generator.GetSpec()

	if pathItem, exists := spec.Paths["/api/public/ready"]; exists && pathItem.Get != nil {
		operation := pathItem.Get

		operation.Summary = "Ready Check"
		operation.Description = "Checks if the service is ready to receive traffic"
		operation.OperationID = "readyCheck"
		operation.Tags = []string{"health"}

		operation.Responses = map[string]openapi.Response{
			"200": {
				Description: "Service ready to receive traffic",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Type: "object",
							Properties: map[string]*openapi.Schema{
								"status": {
									Type: "string",
								},
								"database": {
									Type: "string",
								},
								"timestamp": {
									Type: "string",
								},
							},
						},
					},
				},
			},
			"503": {
				Description: "Service is not ready",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Type: "object",
							Properties: map[string]*openapi.Schema{
								"status": {
									Type: "string",
								},
								"error": {
									Type: "string",
								},
								"database": {
									Type: "string",
								},
							},
						},
					},
				},
			},
		}

		spec.Paths["/api/public/ready"] = pathItem
	}
}
