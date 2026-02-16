package fhir

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

// SearchParamType defines the FHIR search parameter type.
type SearchParamType int

const (
	SearchParamToken     SearchParamType = iota // Token: status, code, category (exact match or system|code)
	SearchParamDate                             // Date: supports prefixes (gt, lt, ge, le, eq, etc.)
	SearchParamString                           // String: case-insensitive prefix match, supports :exact, :contains
	SearchParamReference                        // Reference: handles "ResourceType/uuid" or "uuid"
	SearchParamNumber                           // Number: supports prefixes (gt, lt, ge, le, eq, etc.)
	SearchParamQuantity                         // Quantity: number with unit (treated as number on value column)
	SearchParamURI                              // URI: exact match
)

// SearchParamConfig maps a FHIR search parameter to its database representation.
type SearchParamConfig struct {
	Type      SearchParamType
	Column    string // Primary DB column (code column for tokens)
	SysColumn string // System column for token params (e.g., "code_system")
}

// SearchQuery builds SQL WHERE clauses from FHIR search parameters.
// It encapsulates the common search pattern used across all domain repositories.
type SearchQuery struct {
	table   string
	cols    string
	where   string
	args    []interface{}
	idx     int
	orderBy string
}

// NewSearchQuery creates a new SearchQuery for the given table and columns.
func NewSearchQuery(table, cols string) *SearchQuery {
	return &SearchQuery{
		table: table,
		cols:  cols,
		idx:   1,
	}
}

// Idx returns the next available parameter index.
func (q *SearchQuery) Idx() int { return q.idx }

// Add appends a raw WHERE clause fragment (without leading "AND").
func (q *SearchQuery) Add(clause string, args ...interface{}) {
	q.where += " AND " + clause
	q.args = append(q.args, args...)
	q.idx += len(args)
}

// AddToken adds a token search clause. Handles system|code, |code, system|, or just code.
func (q *SearchQuery) AddToken(sysCol, codeCol, value string) {
	clause, args, nextIdx := TokenSearchClause(sysCol, codeCol, value, q.idx)
	q.where += " AND " + clause
	q.args = append(q.args, args...)
	q.idx = nextIdx
}

// AddDate adds a date search clause with FHIR prefix support (gt, lt, ge, le, eq, etc.).
func (q *SearchQuery) AddDate(column, value string) {
	clause, args, nextIdx := DateSearchClause(column, value, q.idx)
	q.where += " AND " + clause
	q.args = append(q.args, args...)
	q.idx = nextIdx
}

// AddString adds a string search clause with modifier support (exact, contains, prefix).
func (q *SearchQuery) AddString(column, value string, modifier SearchModifier) {
	clause, args, nextIdx := StringSearchClause(column, value, modifier, q.idx)
	q.where += " AND " + clause
	q.args = append(q.args, args...)
	q.idx = nextIdx
}

// AddRef adds a reference search clause. Handles "ResourceType/uuid" or "uuid".
func (q *SearchQuery) AddRef(column, value string) {
	clause, args, nextIdx := ReferenceSearchClause(column, value, q.idx)
	q.where += " AND " + clause
	q.args = append(q.args, args...)
	q.idx = nextIdx
}

// AddNumber adds a number search clause with FHIR prefix support.
func (q *SearchQuery) AddNumber(column, value string) {
	clause, args, nextIdx := NumberSearchClause(column, value, q.idx)
	q.where += " AND " + clause
	q.args = append(q.args, args...)
	q.idx = nextIdx
}

// ApplyParam applies a single FHIR search parameter using the config.
func (q *SearchQuery) ApplyParam(config SearchParamConfig, value string) {
	switch config.Type {
	case SearchParamDate:
		q.AddDate(config.Column, value)
	case SearchParamToken:
		if config.SysColumn != "" {
			q.AddToken(config.SysColumn, config.Column, value)
		} else {
			// Simple token without system column: exact match
			q.where += fmt.Sprintf(" AND %s = $%d", config.Column, q.idx)
			q.args = append(q.args, value)
			q.idx++
		}
	case SearchParamString:
		q.AddString(config.Column, value, "")
	case SearchParamReference:
		q.AddRef(config.Column, value)
	case SearchParamNumber, SearchParamQuantity:
		q.AddNumber(config.Column, value)
	case SearchParamURI:
		q.where += fmt.Sprintf(" AND %s = $%d", config.Column, q.idx)
		q.args = append(q.args, value)
		q.idx++
	}
}

// ApplyParams applies all matching FHIR search parameters from the given map.
func (q *SearchQuery) ApplyParams(params map[string]string, configs map[string]SearchParamConfig) {
	for name, value := range params {
		if config, ok := configs[name]; ok {
			q.ApplyParam(config, value)
		}
	}
}

// OrderBy sets the ORDER BY clause (without the "ORDER BY" keyword).
func (q *SearchQuery) OrderBy(orderBy string) {
	q.orderBy = orderBy
}

// CountSQL returns the count query SQL.
func (q *SearchQuery) CountSQL() string {
	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE 1=1%s", q.table, q.where)
}

// CountArgs returns the arguments for the count query.
func (q *SearchQuery) CountArgs() []interface{} {
	return q.args
}

// DataSQL returns the data query SQL with ORDER BY and LIMIT/OFFSET.
func (q *SearchQuery) DataSQL(limit, offset int) string {
	sql := fmt.Sprintf("SELECT %s FROM %s WHERE 1=1%s", q.cols, q.table, q.where)
	if q.orderBy != "" {
		sql += " ORDER BY " + q.orderBy
	}
	sql += fmt.Sprintf(" LIMIT $%d OFFSET $%d", q.idx, q.idx+1)
	return sql
}

// DataArgs returns the arguments for the data query (search args + limit + offset).
func (q *SearchQuery) DataArgs(limit, offset int) []interface{} {
	result := make([]interface{}, len(q.args)+2)
	copy(result, q.args)
	result[len(q.args)] = limit
	result[len(q.args)+1] = offset
	return result
}

// ApplySort processes the _sort parameter and sets ORDER BY using config column mappings.
// The _sort value is a comma-separated list of param names, optionally prefixed with - for DESC.
// Falls back to the provided defaultOrder if _sort is empty.
func (q *SearchQuery) ApplySort(sortParam, defaultOrder string, configs map[string]SearchParamConfig) {
	if sortParam == "" {
		q.orderBy = defaultOrder
		return
	}
	var parts []string
	for _, field := range strings.Split(sortParam, ",") {
		field = strings.TrimSpace(field)
		desc := false
		if strings.HasPrefix(field, "-") {
			desc = true
			field = field[1:]
		}
		if config, ok := configs[field]; ok {
			col := config.Column
			if desc {
				parts = append(parts, col+" DESC")
			} else {
				parts = append(parts, col+" ASC")
			}
		}
	}
	if len(parts) > 0 {
		q.orderBy = strings.Join(parts, ", ")
	} else {
		q.orderBy = defaultOrder
	}
}

// ExtractSearchParams extracts all FHIR search parameters from the query string,
// excluding FHIR control parameters (_count, _offset, _sort, _elements, etc.).
// Unknown params are included — the repo's ApplyParams will ignore ones not in its config.
func ExtractSearchParams(c echo.Context) map[string]string {
	params := map[string]string{}
	for k, v := range c.QueryParams() {
		if len(v) == 0 || strings.HasPrefix(k, "_") {
			continue
		}
		params[k] = v[0]
	}
	return params
}
