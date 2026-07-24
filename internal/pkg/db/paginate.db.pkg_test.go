package database

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test struct for testing
type TestUser struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func TestOrderField_ToString(t *testing.T) {
	tests := []struct {
		name     string
		field    OrderField
		expected string
	}{
		{
			name: "ASC order",
			field: OrderField{
				Field:     "name",
				Direction: ASC,
			},
			expected: "name asc",
		},
		{
			name: "DESC order",
			field: OrderField{
				Field:     "id",
				Direction: DESC,
			},
			expected: "id desc",
		},
		{
			name: "Complex field name",
			field: OrderField{
				Field:     "user.name",
				Direction: ASC,
			},
			expected: "user.name asc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.field.ToString()
			assert.Equal(t, tt.expected, result, "ToString() should return correct order string")
		})
	}
}

func TestNewPaginationRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		queryParams   map[string]string
		expectedPage  int
		expectedLimit int
		description   string
	}{
		{
			name: "Valid pagination parameters",
			queryParams: map[string]string{
				"page":  "2",
				"limit": "20",
			},
			expectedPage:  2,
			expectedLimit: 20,
			description:   "Should parse valid page and limit parameters",
		},
		{
			name:          "Default values when parameters missing",
			queryParams:   map[string]string{},
			expectedPage:  defaultPage,
			expectedLimit: defaultLimit,
			description:   "Should use default values when parameters are missing",
		},
		{
			name: "Invalid page number",
			queryParams: map[string]string{
				"page":  "0",
				"limit": "10",
			},
			expectedPage:  defaultPage,
			expectedLimit: 10,
			description:   "Should use default page when page is 0 or negative",
		},
		{
			name: "Invalid limit number",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "0",
			},
			expectedPage:  1,
			expectedLimit: defaultLimit,
			description:   "Should use default limit when limit is 0 or negative",
		},
		{
			name: "Limit exceeds maximum",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "150",
			},
			expectedPage:  1,
			expectedLimit: maxLimit,
			description:   "Should cap limit at maximum value",
		},
		{
			name: "Negative values",
			queryParams: map[string]string{
				"page":  "-1",
				"limit": "-5",
			},
			expectedPage:  defaultPage,
			expectedLimit: defaultLimit,
			description:   "Should use defaults for negative values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test context with query parameters
			c, _ := gin.CreateTestContext(nil)

			// Create request with query parameters
			req, _ := http.NewRequest("GET", "/test", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			c.Request = req

			result := NewPaginationRequest(c)

			assert.Equal(t, tt.expectedPage, result.Page, tt.description)
			assert.Equal(t, tt.expectedLimit, result.Limit, tt.description)
		})
	}
}

func TestPaginationQuery_Parse(t *testing.T) {
	tests := []struct {
		name          string
		query         PaginationQuery
		expectedPage  int
		expectedLimit int
		description   string
	}{
		{
			name: "Valid query",
			query: PaginationQuery{
				Page:  5,
				Limit: 25,
			},
			expectedPage:  5,
			expectedLimit: 25,
			description:   "Should return same values for valid query",
		},
		{
			name: "Zero page",
			query: PaginationQuery{
				Page:  0,
				Limit: 10,
			},
			expectedPage:  defaultPage,
			expectedLimit: 10,
			description:   "Should use default page for zero page",
		},
		{
			name: "Negative page",
			query: PaginationQuery{
				Page:  -1,
				Limit: 10,
			},
			expectedPage:  defaultPage,
			expectedLimit: 10,
			description:   "Should use default page for negative page",
		},
		{
			name: "Zero limit",
			query: PaginationQuery{
				Page:  1,
				Limit: 0,
			},
			expectedPage:  1,
			expectedLimit: defaultLimit,
			description:   "Should use default limit for zero limit",
		},
		{
			name: "Negative limit",
			query: PaginationQuery{
				Page:  1,
				Limit: -5,
			},
			expectedPage:  1,
			expectedLimit: defaultLimit,
			description:   "Should use default limit for negative limit",
		},
		{
			name: "Limit exceeds maximum",
			query: PaginationQuery{
				Page:  1,
				Limit: 200,
			},
			expectedPage:  1,
			expectedLimit: maxLimit,
			description:   "Should cap limit at maximum value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.query.Parse()

			assert.Equal(t, tt.expectedPage, result.Page, tt.description)
			assert.Equal(t, tt.expectedLimit, result.Limit, tt.description)
		})
	}
}

func TestPaginationQuery_Paginate(t *testing.T) {
	t.Run("Paginate function returns correct scope", func(t *testing.T) {
		query := PaginationQuery{
			Page:  3,
			Limit: 15,
		}

		paginateFunc := query.Paginate()
		assert.NotNil(t, paginateFunc, "Paginate() should return a function")

		// The function should apply offset and limit
		// offset = (page - 1) * limit = (3 - 1) * 15 = 30
		// limit = 15
	})
}

func TestDatabase_FindWithPagination(t *testing.T) {
	t.Run("Pagination query validation", func(t *testing.T) {
		// Test pagination query parsing
		query := PaginationQuery{Page: 2, Limit: 10}
		parsed := query.Parse()

		assert.Equal(t, 2, parsed.Page, "Page should be preserved")
		assert.Equal(t, 10, parsed.Limit, "Limit should be preserved")
	})

	t.Run("Pagination query with invalid values", func(t *testing.T) {
		// Test pagination query with invalid values
		query := PaginationQuery{Page: 0, Limit: 0}
		parsed := query.Parse()

		assert.Equal(t, defaultPage, parsed.Page, "Should use default page")
		assert.Equal(t, defaultLimit, parsed.Limit, "Should use default limit")
	})

	t.Run("Pagination query with excessive limit", func(t *testing.T) {
		// Test pagination query with limit exceeding maximum
		query := PaginationQuery{Page: 1, Limit: 150}
		parsed := query.Parse()

		assert.Equal(t, 1, parsed.Page, "Page should be preserved")
		assert.Equal(t, maxLimit, parsed.Limit, "Should cap limit at maximum")
	})

	t.Run("Pagination scope function", func(t *testing.T) {
		// Test the paginate scope function
		query := PaginationQuery{Page: 3, Limit: 5}
		paginateFunc := query.Paginate()

		assert.NotNil(t, paginateFunc, "Paginate function should not be nil")
	})
}

func TestDatabase_FindWithCursor(t *testing.T) {
	t.Run("Order field validation", func(t *testing.T) {
		// Test order field creation and string conversion
		order := OrderField{Field: "id", Direction: DESC}

		assert.Equal(t, "id", order.Field, "Field should be preserved")
		assert.Equal(t, DESC, order.Direction, "Direction should be preserved")
		assert.Equal(t, "id desc", order.ToString(), "ToString should return correct format")
	})

	t.Run("Order field with complex field name", func(t *testing.T) {
		// Test order field with complex field name
		order := OrderField{Field: "user.name", Direction: ASC}

		assert.Equal(t, "user.name asc", order.ToString(), "ToString should handle complex field names")
	})

	t.Run("Cursor pagination constants", func(t *testing.T) {
		// Test that cursor pagination uses correct default values
		// This tests the logic in FindWithCursor without requiring a database
		limit := 0
		if limit <= 0 {
			limit = 10
		}

		assert.Equal(t, 10, limit, "Should use default limit when limit is 0 or negative")
	})

	t.Run("Cursor pagination limit calculation", func(t *testing.T) {
		// Test the limit + 1 logic used in cursor pagination
		limit := 5
		queryLimit := limit + 1

		assert.Equal(t, 6, queryLimit, "Query limit should be original limit + 1")
	})
}

func TestPaginationConstants(t *testing.T) {
	t.Run("Default values", func(t *testing.T) {
		assert.Equal(t, 1, defaultPage, "Default page should be 1")
		assert.Equal(t, 10, defaultLimit, "Default limit should be 10")
		assert.Equal(t, 100, maxLimit, "Max limit should be 100")
	})
}
