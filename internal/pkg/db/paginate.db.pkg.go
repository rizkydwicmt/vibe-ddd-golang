package database

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"vibe-ddd-golang/internal/pkg/reqbind"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaginationResult struct {
	CurrentPage int         `json:"currentPage"`
	PerPage     int         `json:"perPage"`
	TotalItems  int64       `json:"totalItems"`
	TotalPages  int         `json:"totalPages"`
	Data        interface{} `json:"data"`
}

type CursorResult struct {
	Items      interface{} `json:"items"`
	NextCursor string      `json:"nextCursor"`
	HasMore    bool        `json:"hasMore"`
	PerPage    int         `json:"perPage"`
}

type OrderField struct {
	Field     string
	Direction DirectionEnum
}

func (o OrderField) ToString() string {
	return fmt.Sprintf("%s %s", o.Field, o.Direction)
}

type PaginationQuery struct {
	Page  int `form:"page" json:"page"`
	Limit int `form:"limit" json:"limit"`
}

type PaginationFilter struct {
	SortBy   string `json:"sortBy"`
	SortType string `json:"sortType"`
}

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

func NewPaginationRequest(c *gin.Context) *PaginationQuery {
	var query PaginationQuery
	_ = reqbind.BindQuery(c, &query)
	return query.Parse()
}

func (q *PaginationQuery) Parse() *PaginationQuery {
	page := q.Page
	if page <= 0 {
		page = defaultPage
	}

	limit := q.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	return &PaginationQuery{
		Page:  page,
		Limit: limit,
	}
}

func (q *PaginationQuery) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		qry := q.Parse()

		offset := (qry.Page - 1) * qry.Limit
		return db.Offset(offset).Limit(qry.Limit)
	}
}

// FindWithPagination executes the query with pagination and returns PaginationResult
//
// Example basic usage:
//
// var users []User
// pagination := database.NewPaginationRequest(c)
// result, err := db.FindWithPagination(pagination, &users)
// // result.Data contains first 10 users
// // result.TotalItems contains total count of users
//
// Example with conditions:
//
// var users []User
// pagination := database.NewPaginationRequest(c)
// db = db.Where("name LIKE ?", "%john%")
// result, err := db.FindWithPagination(pagination, &users)
// // result.Data contains first 10 users with name containing "john"
func (db *Database) FindWithPagination(query *PaginationQuery, dest []interface{}, conditions ...interface{}) (*PaginationResult, error) {
	var totalItems int64

	if err := db.Model(dest).Count(&totalItems).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(query.Limit)))

	if err := db.Scopes(query.Paginate()).Find(dest, conditions...).Error; err != nil {
		return nil, err
	}

	return &PaginationResult{
		CurrentPage: query.Page,
		PerPage:     query.Limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		Data:        dest,
	}, nil
}

// FindWithPaginationFromQuery executes a complex query with pagination and returns PaginationResult
//
// This method is designed to handle complex queries with joins, where clauses, and groups
// that cannot be easily handled by the simple FindWithPagination method.
//
// Example usage with joins:
//
// var serviceRequests []entity.ServiceRequest
// pagination := database.PaginationQuery{Page: 1, Limit: 10}
// query := db.Model(&entity.ServiceRequest{}).
//
//	Joins("left join mydx_workflow_studios as mws on mydx_m_service_request.id = mws.group_service_id").
//	Where("mws.is_deleted = false").
//	Where("mydx_m_service_request.is_deleted = false").
//	Group("mydx_m_service_request.id")
//
// result, err := db.FindWithPaginationFromQuery(pagination, query, &serviceRequests)
func (db *Database) FindWithPaginationFromQuery(query *PaginationQuery, dbQuery *gorm.DB, dest interface{}) (*PaginationResult, error) {
	var totalItems int64

	countQuery := dbQuery.Session(&gorm.Session{})

	if strings.Contains(strings.ToLower(dbQuery.Statement.SQL.String()), "group by") {
		subQuery := countQuery.Select("COUNT(DISTINCT " + getMainTableID(dbQuery) + ")")
		if err := subQuery.Row().Scan(&totalItems); err != nil {
			return nil, err
		}
	} else {
		if err := countQuery.Count(&totalItems).Error; err != nil {
			return nil, err
		}
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(query.Limit)))

	if err := dbQuery.Scopes(query.Paginate()).Find(dest).Error; err != nil {
		return nil, err
	}

	return &PaginationResult{
		CurrentPage: query.Page,
		PerPage:     query.Limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		Data:        dest,
	}, nil
}

// getMainTableID extracts the main table's ID field for counting
// This is a helper function to determine the primary key for distinct counting
func getMainTableID(dbQuery *gorm.DB) string {
	if dbQuery.Statement.Table != "" {
		return dbQuery.Statement.Table + ".id"
	}
	return "id"
}

// FindWithCursor executes the query with infinite scrolling and returns CursorResult
//
// Example basic usage:
//
//	var users []User
//	result, err := db.FindWithCursor("", 10, &users, "id")
//	// result.Items contains first 10 users
//	// result.NextCursor contains the cursor for the next page
//	// result.HasMore is true if there are more items
func (db *Database) FindWithCursor(encryptedCursor string, limit int, dest interface{}, order OrderField) (*CursorResult, error) {
	if limit <= 0 {
		limit = 10
	}

	limit++
	query := db.DB

	if encryptedCursor != "" {
		cursor, err := db.cursorCrypto.Decrypt(encryptedCursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		if cursor != "" {
			query = query.Where(order.Field+" < ?", cursor)
		}
	}

	query = query.Order(order.ToString()).Limit(limit)

	if err := query.Find(dest).Error; err != nil {
		return nil, err
	}

	result := &CursorResult{
		Items:   dest,
		PerPage: limit - 1,
	}

	items := reflect.ValueOf(dest).Elem()
	if items.Len() == limit {
		items.Set(items.Slice(0, items.Len()-1))
		result.HasMore = true

		lastItem := items.Index(items.Len() - 1)
		fieldParts := strings.Split(order.Field, ".")
		cursorField := lastItem.FieldByName(cases.Title(language.Und).String(fieldParts[len(fieldParts)-1]))
		nextCursor, err := db.cursorCrypto.Encrypt(fmt.Sprint(cursorField.Interface()))
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt cursor: %w", err)
		}
		result.NextCursor = nextCursor
	}

	return result, nil
}
