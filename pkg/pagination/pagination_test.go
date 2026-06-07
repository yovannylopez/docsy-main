package pagination

import (
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/yovannylopez/docsy-main/pkg/constants"
)

func TestDefaultConfig_MatchesConstants(t *testing.T) {
	if DefaultConfig.DefaultLimit != constants.DefaultPageSize || DefaultConfig.MaxLimit != constants.MaxPageSize {
		t.Fatalf("DefaultConfig desalineado con pkg/constants: %+v vs DefaultPageSize=%d MaxPageSize=%d",
			DefaultConfig, constants.DefaultPageSize, constants.MaxPageSize)
	}
}

func TestParser_ParseFromQuery(t *testing.T) {
	tests := []struct {
		name           string
		limitStr       string
		offsetStr      string
		config         Config
		expectedParams *Params
		expectedError  bool
	}{
		{
			name:           "valid parameters",
			limitStr:       "20",
			offsetStr:      "40",
			config:         DefaultConfig,
			expectedParams: &Params{Limit: 20, Offset: 40},
			expectedError:  false,
		},
		{
			name:           "empty parameters - use defaults",
			limitStr:       "",
			offsetStr:      "",
			config:         DefaultConfig,
			expectedParams: &Params{Limit: 20, Offset: 0},
			expectedError:  false,
		},
		{
			name:           "invalid limit",
			limitStr:       "invalid",
			offsetStr:      "0",
			config:         DefaultConfig,
			expectedParams: nil,
			expectedError:  true,
		},
		{
			name:           "invalid offset",
			limitStr:       "10",
			offsetStr:      "invalid",
			config:         DefaultConfig,
			expectedParams: nil,
			expectedError:  true,
		},
		{
			name:           "limit too high",
			limitStr:       "200",
			offsetStr:      "0",
			config:         DefaultConfig,
			expectedParams: nil,
			expectedError:  true,
		},
		{
			name:           "negative offset",
			limitStr:       "10",
			offsetStr:      "-5",
			config:         DefaultConfig,
			expectedParams: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.config)
			params, err := parser.ParseFromQuery(tt.limitStr, tt.offsetStr)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if params != nil {
					t.Errorf("expected nil params but got %v", params)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if params == nil {
					t.Errorf("expected params but got nil")
				} else {
					if params.Limit != tt.expectedParams.Limit {
						t.Errorf("expected limit %d, got %d", tt.expectedParams.Limit, params.Limit)
					}
					if params.Offset != tt.expectedParams.Offset {
						t.Errorf("expected offset %d, got %d", tt.expectedParams.Offset, params.Offset)
					}
				}
			}
		})
	}
}

func TestParser_Validate_SentinelErrors(t *testing.T) {
	p := NewParser(DefaultConfig)
	assertErrorIs := func(t *testing.T, err error, target error) {
		t.Helper()
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, target) {
			t.Fatalf("errors.Is: want %v, err=%v", target, err)
		}
	}

	t.Run("limit bajo", func(t *testing.T) {
		assertErrorIs(t, p.Validate(&Params{Limit: 0, Offset: 0}), ErrLimitOutOfRange)
	})
	t.Run("limit alto", func(t *testing.T) {
		assertErrorIs(t, p.Validate(&Params{Limit: 500, Offset: 0}), ErrLimitOutOfRange)
	})
	t.Run("offset negativo", func(t *testing.T) {
		assertErrorIs(t, p.Validate(&Params{Limit: 10, Offset: -1}), ErrNegativeOffset)
	})
}

func TestParser_Validate(t *testing.T) {
	tests := []struct {
		name          string
		params        *Params
		config        Config
		expectedError bool
	}{
		{
			name:          "valid parameters",
			params:        &Params{Limit: 20, Offset: 40},
			config:        DefaultConfig,
			expectedError: false,
		},
		{
			name:          "limit too low",
			params:        &Params{Limit: 0, Offset: 0},
			config:        DefaultConfig,
			expectedError: true,
		},
		{
			name:          "limit too high",
			params:        &Params{Limit: 200, Offset: 0},
			config:        DefaultConfig,
			expectedError: true,
		},
		{
			name:          "negative offset",
			params:        &Params{Limit: 20, Offset: -5},
			config:        DefaultConfig,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.config)
			err := parser.Validate(tt.params)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCreateMetadata(t *testing.T) {
	tests := []struct {
		name         string
		params       *Params
		total        int
		expectedMeta Metadata
	}{
		{
			name:   "first page",
			params: &Params{Limit: 10, Offset: 0},
			total:  25,
			expectedMeta: Metadata{
				Total:       25,
				Limit:       10,
				Offset:      0,
				TotalPages:  3,
				CurrentPage: 1,
				HasNext:     true,
				HasPrev:     false,
			},
		},
		{
			name:   "middle page",
			params: &Params{Limit: 10, Offset: 10},
			total:  25,
			expectedMeta: Metadata{
				Total:       25,
				Limit:       10,
				Offset:      10,
				TotalPages:  3,
				CurrentPage: 2,
				HasNext:     true,
				HasPrev:     true,
			},
		},
		{
			name:   "last page",
			params: &Params{Limit: 10, Offset: 20},
			total:  25,
			expectedMeta: Metadata{
				Total:       25,
				Limit:       10,
				Offset:      20,
				TotalPages:  3,
				CurrentPage: 3,
				HasNext:     false,
				HasPrev:     true,
			},
		},
		{
			name:   "empty result",
			params: &Params{Limit: 10, Offset: 0},
			total:  0,
			expectedMeta: Metadata{
				Total:       0,
				Limit:       10,
				Offset:      0,
				TotalPages:  1,
				CurrentPage: 1,
				HasNext:     false,
				HasPrev:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := CreateMetadata(tt.params, tt.total)
			if meta != tt.expectedMeta {
				t.Errorf("expected %v, got %v", tt.expectedMeta, meta)
			}
		})
	}
}

func TestCreateResponse(t *testing.T) {
	data := []string{"item1", "item2", "item3"}
	params := &Params{Limit: 10, Offset: 0}
	total := 25

	response := CreateResponse(data, params, total)

	if !reflect.DeepEqual(response.Data, data) {
		t.Errorf("expected data %v, got %v", data, response.Data)
	}
	if response.Metadata.Total != 25 {
		t.Errorf("expected total 25, got %d", response.Metadata.Total)
	}
	if response.Metadata.Limit != 10 {
		t.Errorf("expected limit 10, got %d", response.Metadata.Limit)
	}
	if response.Metadata.Offset != 0 {
		t.Errorf("expected offset 0, got %d", response.Metadata.Offset)
	}
	if response.Metadata.TotalPages != 3 {
		t.Errorf("expected total pages 3, got %d", response.Metadata.TotalPages)
	}
	if response.Metadata.CurrentPage != 1 {
		t.Errorf("expected current page 1, got %d", response.Metadata.CurrentPage)
	}
	if !response.Metadata.HasNext {
		t.Errorf("expected has next true, got false")
	}
	if response.Metadata.HasPrev {
		t.Errorf("expected has prev false, got true")
	}
}

func TestGetPageFromOffset(t *testing.T) {
	tests := []struct {
		name     string
		offset   int
		limit    int
		expected int
	}{
		{"first page", 0, 10, 1},
		{"second page", 10, 10, 2},
		{"third page", 20, 10, 3},
		{"zero limit", 20, 0, 1},
		{"negative limit", 20, -5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPageFromOffset(tt.offset, tt.limit)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCreateMetadata_OffsetBeyondTotal(t *testing.T) {
	meta := CreateMetadata(&Params{Limit: 10, Offset: 100}, 25)
	if meta.TotalPages != 3 {
		t.Fatalf("total_pages: want 3, got %d", meta.TotalPages)
	}
	if meta.CurrentPage != 3 {
		t.Fatalf("current_page clamped to total_pages: want 3, got %d", meta.CurrentPage)
	}
	if meta.HasNext {
		t.Fatal("HasNext should be false when on last page")
	}
	if !meta.HasPrev {
		t.Fatal("HasPrev should be true when on last page with offset past range")
	}
}

func TestParseFromQuery_WrapsStrconvError(t *testing.T) {
	p := NewParser(DefaultConfig)
	_, err := p.ParseFromQuery("x", "0")
	if err == nil {
		t.Fatal("expected error")
	}
	var numErr *strconv.NumError
	if !errors.As(err, &numErr) {
		t.Fatalf("expected wrapped strconv error, got %v", err)
	}
}

func TestNewDefaultParser(t *testing.T) {
	p := NewDefaultParser()
	params, err := p.ParseFromQuery("", "")
	if err != nil {
		t.Fatal(err)
	}
	if params.Limit != DefaultConfig.DefaultLimit || params.Offset != 0 {
		t.Fatalf("got %+v", params)
	}
}

func TestGetOffsetFromPage(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		expected int
	}{
		{"first page", 1, 10, 0},
		{"second page", 2, 10, 10},
		{"third page", 3, 10, 20},
		{"zero page", 0, 10, 0},
		{"negative page", -1, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetOffsetFromPage(tt.page, tt.limit)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}
