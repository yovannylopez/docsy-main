package main

import (
	"fmt"
	"log"

	"github.com/yovannylopez/docsy-main/pkg/pagination"
)

func main() {
	// Example usage of the pagination package

	// Create parser with default configuration
	parser := pagination.NewDefaultParser()

	// Simulate parsing of query parameters
	limitStr := "15"
	offsetStr := "30"

	// Parse parameters
	params, err := parser.ParseFromQuery(limitStr, offsetStr)
	if err != nil {
		log.Fatalf("Error parsing pagination params: %v", err)
	}

	fmt.Printf("Parsed params: Limit=%d, Offset=%d\n", params.Limit, params.Offset)

	// Simulate data retrieved from the repository
	data := []string{"item1", "item2", "item3", "item4", "item5"}
	total := 150

	// Create paginated response
	response := pagination.CreateResponse(data, params, total)

	fmt.Printf("Response: %+v\n", response)

	// Example of page/offset conversion
	page := pagination.GetPageFromOffset(params.Offset, params.Limit)
	fmt.Printf("Current page: %d\n", page)

	offset := pagination.GetOffsetFromPage(page, params.Limit)
	fmt.Printf("Offset for page %d: %d\n", page, offset)
}
