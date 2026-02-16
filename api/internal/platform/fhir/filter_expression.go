package fhir

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// Advanced _filter expression parser and SQL compiler
//
// This file implements a full recursive-descent parser for FHIR _filter
// expressions, supporting boolean combinators (AND, OR, NOT), parenthesized
// subexpressions, operator precedence, and all FHIR filter operators.
//
// The result is an expression tree (FilterExprNode) that can be validated,
// compiled to SQL, serialized back to string form, and simplified.
// ---------------------------------------------------------------------------

// FilterExprType identifies the kind of filter expression node.
type FilterExprType int

const (
	FilterExprParam FilterExprType = iota // Leaf: paramName op value
	FilterExprAnd                         // And: left AND right
	FilterExprOr                          // Or: left OR right
	FilterExprNot                         // Not: negate child
)

// FilterExprNode is a tree node for advanced filter expressions.
// Leaf nodes use Param/Operator/Value fields.
// And/Or nodes use Left/Right.
// Not nodes use Child.
type FilterExprNode struct {
	Type     FilterExprType
	Left     *FilterExprNode // For And/Or
	Right    *FilterExprNode // For And/Or
	Child    *FilterExprNode // For Not
	Param    string          // For Param: parameter name
	Operator string          // For Param: eq, ne, gt, lt, ge, le, co, sw, ew, sa, eb, ap, pr, in, ni, ss, sb
	Value    string          // For Param: the comparison value (empty for pr)
}

// FilterOperator describes a filter comparison operator.
type FilterOperator string

const (
	FilterOperatorEqual          FilterOperator = "eq"
	FilterOperatorNotEqual       FilterOperator = "ne"
	FilterOperatorGreaterThan    FilterOperator = "gt"
	FilterOperatorLessThan       FilterOperator = "lt"
	FilterOperatorGreaterOrEqual FilterOperator = "ge"
	FilterOperatorLessOrEqual    FilterOperator = "le"
	FilterOperatorContains       FilterOperator = "co"
	FilterOperatorStartsWith     FilterOperator = "sw"
	FilterOperatorEndsWith       FilterOperator = "ew"
	FilterOperatorStartsAfter    FilterOperator = "sa"
	FilterOperatorEndsBefore     FilterOperator = "eb"
	FilterOperatorApproximately  FilterOperator = "ap"
	FilterOperatorPresent        FilterOperator = "pr"
	FilterOperatorIn             FilterOperator = "in"
	FilterOperatorNotIn          FilterOperator = "ni"
	FilterOperatorSubsumes       FilterOperator = "ss"
	FilterOperatorSubsumedBy     FilterOperator = "sb"
)

// validFilterOperators is the set of all recognised _filter operators.
var validFilterOperators = map[string]bool{
	"eq": true, "ne": true,
	"gt": true, "lt": true,
	"ge": true, "le": true,
	"co": true, "sw": true, "ew": true,
	"sa": true, "eb": true, "ap": true,
	"pr": true,
	"in": true, "ni": true,
	"ss": true, "sb": true,
}

// FilterColumnMapping maps a FHIR parameter name to SQL columns and a type.
type FilterColumnMapping struct {
	Column    string // Primary DB column
	SysColumn string // For token types: the system column
	ParamType string // string, token, date, number, reference, quantity
}

// FilterContext provides context for compiling a filter expression to SQL.
type FilterContext struct {
	ResourceType   string
	ColumnMappings map[string]FilterColumnMapping
	TableAlias     string // e.g., "r" for the main resource table
}

// ---------------------------------------------------------------------------
// Tokenizer
// ---------------------------------------------------------------------------

type filterTokenType int

const (
	tokenWord   filterTokenType = iota // An unquoted word
	tokenString                        // A double-quoted string (quotes stripped)
	tokenLParen
	tokenRParen
	tokenAnd
	tokenOr
	tokenNot
)

type filterToken struct {
	Type  filterTokenType
	Value string
}

// tokenizeFilter splits a _filter string into lexical tokens.
func tokenizeFilter(filter string) ([]filterToken, error) {
	var tokens []filterToken
	i := 0
	n := len(filter)

	for i < n {
		ch := filter[i]

		// Skip whitespace.
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}

		// Parentheses.
		if ch == '(' {
			tokens = append(tokens, filterToken{Type: tokenLParen, Value: "("})
			i++
			continue
		}
		if ch == ')' {
			tokens = append(tokens, filterToken{Type: tokenRParen, Value: ")"})
			i++
			continue
		}

		// Quoted string.
		if ch == '"' {
			j := i + 1
			for j < n && filter[j] != '"' {
				j++
			}
			if j >= n {
				return nil, fmt.Errorf("unclosed quoted string starting at position %d", i)
			}
			tokens = append(tokens, filterToken{Type: tokenString, Value: filter[i+1 : j]})
			i = j + 1
			continue
		}

		// Word: read until whitespace, paren, or quote.
		j := i
		for j < n && filter[j] != ' ' && filter[j] != '\t' && filter[j] != '\n' &&
			filter[j] != '\r' && filter[j] != '(' && filter[j] != ')' && filter[j] != '"' {
			j++
		}
		word := filter[i:j]
		i = j

		// Check if the word is a keyword.
		switch strings.ToLower(word) {
		case "and":
			tokens = append(tokens, filterToken{Type: tokenAnd, Value: "and"})
		case "or":
			tokens = append(tokens, filterToken{Type: tokenOr, Value: "or"})
		case "not":
			tokens = append(tokens, filterToken{Type: tokenNot, Value: "not"})
		default:
			tokens = append(tokens, filterToken{Type: tokenWord, Value: word})
		}
	}

	return tokens, nil
}

// ---------------------------------------------------------------------------
// Recursive descent parser
//
// Grammar (with precedence):
//   expr     -> orExpr
//   orExpr   -> andExpr ("or" andExpr)*
//   andExpr  -> unaryExpr ("and" unaryExpr)*
//   unaryExpr -> "not" unaryExpr | primary
//   primary  -> "(" expr ")" | paramExpr
//   paramExpr -> WORD OPERATOR (VALUE | epsilon for "pr")
// ---------------------------------------------------------------------------

type filterParser struct {
	tokens []filterToken
	pos    int
}

func (p *filterParser) peek() *filterToken {
	if p.pos >= len(p.tokens) {
		return nil
	}
	return &p.tokens[p.pos]
}

func (p *filterParser) advance() *filterToken {
	if p.pos >= len(p.tokens) {
		return nil
	}
	t := &p.tokens[p.pos]
	p.pos++
	return t
}

func (p *filterParser) expect(tt filterTokenType) (*filterToken, error) {
	t := p.peek()
	if t == nil {
		return nil, fmt.Errorf("unexpected end of expression, expected token type %d", tt)
	}
	if t.Type != tt {
		return nil, fmt.Errorf("unexpected token %q, expected token type %d", t.Value, tt)
	}
	return p.advance(), nil
}

// ParseFilterExpression parses a FHIR _filter string into an expression tree.
func ParseFilterExpression(filter string) (*FilterExprNode, error) {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return nil, fmt.Errorf("empty filter expression")
	}

	tokens, err := tokenizeFilter(filter)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty filter expression")
	}

	p := &filterParser{tokens: tokens}
	expr, err := p.parseOrExpr()
	if err != nil {
		return nil, err
	}

	// Ensure all tokens were consumed.
	if p.pos < len(p.tokens) {
		remaining := p.tokens[p.pos]
		return nil, fmt.Errorf("unexpected token %q at position %d", remaining.Value, p.pos)
	}

	return expr, nil
}

func (p *filterParser) parseOrExpr() (*FilterExprNode, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}

	for {
		t := p.peek()
		if t == nil || t.Type != tokenOr {
			break
		}
		p.advance() // consume "or"
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &FilterExprNode{
			Type:  FilterExprOr,
			Left:  left,
			Right: right,
		}
	}

	return left, nil
}

func (p *filterParser) parseAndExpr() (*FilterExprNode, error) {
	left, err := p.parseUnaryExpr()
	if err != nil {
		return nil, err
	}

	for {
		t := p.peek()
		if t == nil || t.Type != tokenAnd {
			break
		}
		p.advance() // consume "and"
		right, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}
		left = &FilterExprNode{
			Type:  FilterExprAnd,
			Left:  left,
			Right: right,
		}
	}

	return left, nil
}

func (p *filterParser) parseUnaryExpr() (*FilterExprNode, error) {
	t := p.peek()
	if t != nil && t.Type == tokenNot {
		p.advance() // consume "not"
		child, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}
		return &FilterExprNode{
			Type:  FilterExprNot,
			Child: child,
		}, nil
	}
	return p.parsePrimary()
}

func (p *filterParser) parsePrimary() (*FilterExprNode, error) {
	t := p.peek()
	if t == nil {
		return nil, fmt.Errorf("unexpected end of expression, expected parameter or '('")
	}

	// Parenthesized subexpression.
	if t.Type == tokenLParen {
		p.advance() // consume "("
		expr, err := p.parseOrExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.expect(tokenRParen)
		if err != nil {
			return nil, fmt.Errorf("expected ')' to close parenthesized expression")
		}
		return expr, nil
	}

	// Must be a parameter expression: param op value
	return p.parseParamExpr()
}

func (p *filterParser) parseParamExpr() (*FilterExprNode, error) {
	paramTok := p.peek()
	if paramTok == nil || (paramTok.Type != tokenWord && paramTok.Type != tokenString) {
		if paramTok == nil {
			return nil, fmt.Errorf("unexpected end of expression, expected parameter name")
		}
		return nil, fmt.Errorf("unexpected token %q, expected parameter name", paramTok.Value)
	}
	p.advance()
	paramName := paramTok.Value

	// Operator
	opTok := p.peek()
	if opTok == nil || (opTok.Type != tokenWord && opTok.Type != tokenString) {
		return nil, fmt.Errorf("expected operator after parameter %q", paramName)
	}
	p.advance()
	operator := opTok.Value

	if !validFilterOperators[operator] {
		return nil, fmt.Errorf("unknown filter operator %q after parameter %q", operator, paramName)
	}

	// "pr" (present) is a unary operator with no value.
	if operator == "pr" {
		return &FilterExprNode{
			Type:     FilterExprParam,
			Param:    paramName,
			Operator: operator,
			Value:    "",
		}, nil
	}

	// Value
	valTok := p.peek()
	if valTok == nil {
		return nil, fmt.Errorf("expected value after operator %q for parameter %q", operator, paramName)
	}
	// Accept word or string token as a value.
	if valTok.Type != tokenWord && valTok.Type != tokenString {
		return nil, fmt.Errorf("expected value after operator %q for parameter %q, got %q", operator, paramName, valTok.Value)
	}
	p.advance()

	return &FilterExprNode{
		Type:     FilterExprParam,
		Param:    paramName,
		Operator: operator,
		Value:    valTok.Value,
	}, nil
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------

// validStringOps lists operators valid for string parameters.
var validStringOps = map[string]bool{
	"eq": true, "ne": true, "co": true, "sw": true, "ew": true, "pr": true,
}

// validDateOps lists operators valid for date parameters.
var validDateOps = map[string]bool{
	"eq": true, "ne": true, "gt": true, "lt": true,
	"ge": true, "le": true, "sa": true, "eb": true, "ap": true, "pr": true,
}

// validTokenOps lists operators valid for token parameters.
var validTokenOps = map[string]bool{
	"eq": true, "ne": true, "in": true, "ni": true,
	"ss": true, "sb": true, "pr": true,
}

// validNumberOps lists operators valid for number/quantity parameters.
var validNumberOps = map[string]bool{
	"eq": true, "ne": true, "gt": true, "lt": true,
	"ge": true, "le": true, "pr": true,
}

// ValidateFilterExpression validates that a filter expression is well-formed
// with respect to the given column mappings and parameter types.
func ValidateFilterExpression(expr *FilterExprNode, ctx *FilterContext) []ValidationIssue {
	if expr == nil {
		return nil
	}

	var issues []ValidationIssue

	switch expr.Type {
	case FilterExprParam:
		mapping, ok := ctx.ColumnMappings[expr.Param]
		if !ok {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Diagnostics: fmt.Sprintf("unknown filter parameter %q for resource type %q", expr.Param, ctx.ResourceType),
			})
			return issues
		}
		if !isOperatorValidForType(expr.Operator, mapping.ParamType) {
			issues = append(issues, ValidationIssue{
				Severity:    SeverityError,
				Code:        VIssueTypeValue,
				Diagnostics: fmt.Sprintf("operator %q is not valid for parameter %q of type %q", expr.Operator, expr.Param, mapping.ParamType),
			})
		}

	case FilterExprAnd, FilterExprOr:
		issues = append(issues, ValidateFilterExpression(expr.Left, ctx)...)
		issues = append(issues, ValidateFilterExpression(expr.Right, ctx)...)

	case FilterExprNot:
		issues = append(issues, ValidateFilterExpression(expr.Child, ctx)...)
	}

	return issues
}

// isOperatorValidForType checks if a filter operator is valid for a given parameter type.
func isOperatorValidForType(op, paramType string) bool {
	switch paramType {
	case "string":
		return validStringOps[op]
	case "date":
		return validDateOps[op]
	case "token":
		return validTokenOps[op]
	case "number", "quantity":
		return validNumberOps[op]
	case "reference":
		return op == "eq" || op == "ne" || op == "pr"
	default:
		// For unknown types, accept all operators.
		return true
	}
}

// ---------------------------------------------------------------------------
// SQL compilation
// ---------------------------------------------------------------------------

// CompileFilterToSQL converts a filter expression tree to a SQL WHERE clause.
// startIdx is the starting positional parameter index ($1, $2, ...).
// Returns the SQL string, bind arguments, and an error if compilation fails.
func CompileFilterToSQL(expr *FilterExprNode, ctx *FilterContext, startIdx int) (string, []interface{}, error) {
	if expr == nil {
		return "", nil, fmt.Errorf("nil filter expression")
	}

	switch expr.Type {
	case FilterExprParam:
		return compileParamToSQL(expr.Param, expr.Operator, expr.Value, ctx, startIdx)

	case FilterExprAnd:
		leftSQL, leftArgs, err := CompileFilterToSQL(expr.Left, ctx, startIdx)
		if err != nil {
			return "", nil, err
		}
		rightSQL, rightArgs, err := CompileFilterToSQL(expr.Right, ctx, startIdx+len(leftArgs))
		if err != nil {
			return "", nil, err
		}
		sql := fmt.Sprintf("(%s AND %s)", leftSQL, rightSQL)
		args := append(leftArgs, rightArgs...)
		return sql, args, nil

	case FilterExprOr:
		leftSQL, leftArgs, err := CompileFilterToSQL(expr.Left, ctx, startIdx)
		if err != nil {
			return "", nil, err
		}
		rightSQL, rightArgs, err := CompileFilterToSQL(expr.Right, ctx, startIdx+len(leftArgs))
		if err != nil {
			return "", nil, err
		}
		sql := fmt.Sprintf("(%s OR %s)", leftSQL, rightSQL)
		args := append(leftArgs, rightArgs...)
		return sql, args, nil

	case FilterExprNot:
		childSQL, childArgs, err := CompileFilterToSQL(expr.Child, ctx, startIdx)
		if err != nil {
			return "", nil, err
		}
		sql := fmt.Sprintf("NOT (%s)", childSQL)
		return sql, childArgs, nil

	default:
		return "", nil, fmt.Errorf("unknown filter expression type %d", expr.Type)
	}
}

// compileParamToSQL compiles a leaf parameter comparison to SQL.
func compileParamToSQL(param, operator, value string, ctx *FilterContext, startIdx int) (string, []interface{}, error) {
	mapping, ok := ctx.ColumnMappings[param]
	if !ok {
		return "", nil, fmt.Errorf("unknown filter parameter %q", param)
	}

	column := mapping.Column
	if ctx.TableAlias != "" {
		column = ctx.TableAlias + "." + mapping.Column
	}

	sysColumn := mapping.SysColumn
	if sysColumn != "" && ctx.TableAlias != "" {
		sysColumn = ctx.TableAlias + "." + mapping.SysColumn
	}

	// Handle the "pr" (present) operator.
	if operator == "pr" {
		sql := compilePresentFilter(column, false)
		return sql, nil, nil
	}

	switch mapping.ParamType {
	case "string":
		sql, args := compileStringFilter(column, operator, value, startIdx)
		return sql, args, nil

	case "token":
		sql, args := compileTokenFilter(column, sysColumn, operator, value, startIdx)
		return sql, args, nil

	case "date":
		sql, args := compileDateFilter(column, operator, value, startIdx)
		return sql, args, nil

	case "number", "quantity":
		sql, args := compileNumberFilter(column, operator, value, startIdx)
		return sql, args, nil

	case "reference":
		// References use eq/ne with simple column match.
		sql, args := compileStringFilter(column, operator, value, startIdx)
		return sql, args, nil

	default:
		// Fallback: treat as string.
		sql, args := compileStringFilter(column, operator, value, startIdx)
		return sql, args, nil
	}
}

// compileStringFilter compiles string comparisons to SQL.
func compileStringFilter(column, operator, value string, startIdx int) (string, []interface{}) {
	placeholder := fmt.Sprintf("$%d", startIdx)

	switch operator {
	case "eq":
		return fmt.Sprintf("%s = %s", column, placeholder), []interface{}{value}
	case "ne":
		return fmt.Sprintf("%s != %s", column, placeholder), []interface{}{value}
	case "co":
		return fmt.Sprintf("%s ILIKE %s", column, placeholder), []interface{}{"%" + value + "%"}
	case "sw":
		return fmt.Sprintf("%s ILIKE %s", column, placeholder), []interface{}{value + "%"}
	case "ew":
		return fmt.Sprintf("%s ILIKE %s", column, placeholder), []interface{}{"%" + value}
	default:
		return fmt.Sprintf("%s = %s", column, placeholder), []interface{}{value}
	}
}

// compileTokenFilter compiles token comparisons to SQL.
func compileTokenFilter(column, sysColumn, operator, value string, startIdx int) (string, []interface{}) {
	switch operator {
	case "in":
		return compileTokenInFilter(column, value, startIdx, false)
	case "ni":
		return compileTokenInFilter(column, value, startIdx, true)
	case "eq", "ne", "ss", "sb":
		// Check for system|code format.
		if strings.Contains(value, "|") && sysColumn != "" {
			return compileTokenSystemCode(column, sysColumn, operator, value, startIdx)
		}
		placeholder := fmt.Sprintf("$%d", startIdx)
		switch operator {
		case "ne":
			return fmt.Sprintf("%s != %s", column, placeholder), []interface{}{value}
		default: // eq, ss, sb all compile to = for basic token matching
			return fmt.Sprintf("%s = %s", column, placeholder), []interface{}{value}
		}
	default:
		placeholder := fmt.Sprintf("$%d", startIdx)
		return fmt.Sprintf("%s = %s", column, placeholder), []interface{}{value}
	}
}

// compileTokenSystemCode handles the system|code format for token filters.
func compileTokenSystemCode(column, sysColumn, operator, value string, startIdx int) (string, []interface{}) {
	parts := strings.SplitN(value, "|", 2)
	system := parts[0]
	code := ""
	if len(parts) > 1 {
		code = parts[1]
	}

	if system != "" && code != "" {
		sysPH := fmt.Sprintf("$%d", startIdx)
		codePH := fmt.Sprintf("$%d", startIdx+1)
		if operator == "ne" {
			return fmt.Sprintf("NOT (%s = %s AND %s = %s)", sysColumn, sysPH, column, codePH),
				[]interface{}{system, code}
		}
		return fmt.Sprintf("(%s = %s AND %s = %s)", sysColumn, sysPH, column, codePH),
			[]interface{}{system, code}
	} else if system != "" {
		sysPH := fmt.Sprintf("$%d", startIdx)
		if operator == "ne" {
			return fmt.Sprintf("%s != %s", sysColumn, sysPH), []interface{}{system}
		}
		return fmt.Sprintf("%s = %s", sysColumn, sysPH), []interface{}{system}
	} else if code != "" {
		codePH := fmt.Sprintf("$%d", startIdx)
		if operator == "ne" {
			return fmt.Sprintf("%s != %s", column, codePH), []interface{}{code}
		}
		return fmt.Sprintf("%s = %s", column, codePH), []interface{}{code}
	}

	// Empty system and code: match anything.
	return "1=1", nil
}

// compileTokenInFilter compiles an IN or NOT IN clause for tokens.
func compileTokenInFilter(column, value string, startIdx int, negate bool) (string, []interface{}) {
	values := strings.Split(value, ",")
	placeholders := make([]string, len(values))
	args := make([]interface{}, len(values))
	for i, v := range values {
		placeholders[i] = fmt.Sprintf("$%d", startIdx+i)
		args[i] = strings.TrimSpace(v)
	}

	op := "IN"
	if negate {
		op = "NOT IN"
	}
	sql := fmt.Sprintf("%s %s (%s)", column, op, strings.Join(placeholders, ", "))
	return sql, args
}

// compileDateFilter compiles date comparisons to SQL.
func compileDateFilter(column, operator, value string, startIdx int) (string, []interface{}) {
	placeholder := fmt.Sprintf("$%d", startIdx)

	switch operator {
	case "eq":
		return fmt.Sprintf("%s = %s", column, placeholder), []interface{}{value}
	case "ne":
		return fmt.Sprintf("%s != %s", column, placeholder), []interface{}{value}
	case "gt", "sa":
		return fmt.Sprintf("%s > %s", column, placeholder), []interface{}{value}
	case "lt", "eb":
		return fmt.Sprintf("%s < %s", column, placeholder), []interface{}{value}
	case "ge":
		return fmt.Sprintf("%s >= %s", column, placeholder), []interface{}{value}
	case "le":
		return fmt.Sprintf("%s <= %s", column, placeholder), []interface{}{value}
	case "ap":
		// Approximate: BETWEEN value-1day AND value+1day.
		low := fmt.Sprintf("$%d", startIdx)
		high := fmt.Sprintf("$%d", startIdx+1)
		sql := fmt.Sprintf("%s BETWEEN %s AND %s", column, low, high)
		return sql, []interface{}{value + "::date - interval '1 day'", value + "::date + interval '1 day'"}
	default:
		return fmt.Sprintf("%s = %s", column, placeholder), []interface{}{value}
	}
}

// compileNumberFilter compiles number comparisons to SQL.
func compileNumberFilter(column, operator, value string, startIdx int) (string, []interface{}) {
	placeholder := fmt.Sprintf("$%d", startIdx)

	switch operator {
	case "eq":
		return fmt.Sprintf("%s = %s", column, placeholder), []interface{}{value}
	case "ne":
		return fmt.Sprintf("%s != %s", column, placeholder), []interface{}{value}
	case "gt":
		return fmt.Sprintf("%s > %s", column, placeholder), []interface{}{value}
	case "lt":
		return fmt.Sprintf("%s < %s", column, placeholder), []interface{}{value}
	case "ge":
		return fmt.Sprintf("%s >= %s", column, placeholder), []interface{}{value}
	case "le":
		return fmt.Sprintf("%s <= %s", column, placeholder), []interface{}{value}
	default:
		return fmt.Sprintf("%s = %s", column, placeholder), []interface{}{value}
	}
}

// compilePresentFilter compiles the "pr" (present) operator.
// If negated is true, it produces IS NULL instead of IS NOT NULL.
func compilePresentFilter(column string, negated bool) string {
	if negated {
		return column + " IS NULL"
	}
	return column + " IS NOT NULL"
}

// ---------------------------------------------------------------------------
// Default column mappings
// ---------------------------------------------------------------------------

// DefaultFilterColumnMappings returns column mappings for standard FHIR resource types.
func DefaultFilterColumnMappings(resourceType string) map[string]FilterColumnMapping {
	switch resourceType {
	case "Patient":
		return map[string]FilterColumnMapping{
			"name":       {Column: "last_name", ParamType: "string"},
			"family":     {Column: "last_name", ParamType: "string"},
			"given":      {Column: "first_name", ParamType: "string"},
			"birthdate":  {Column: "birth_date", ParamType: "date"},
			"gender":     {Column: "gender", ParamType: "token"},
			"active":     {Column: "active", ParamType: "token"},
			"identifier": {Column: "identifier_value", SysColumn: "identifier_system", ParamType: "token"},
			"address":    {Column: "address", ParamType: "string"},
			"telecom":    {Column: "telecom", ParamType: "token"},
		}
	case "Observation":
		return map[string]FilterColumnMapping{
			"code":           {Column: "code_value", SysColumn: "code_system", ParamType: "token"},
			"status":         {Column: "status", ParamType: "token"},
			"date":           {Column: "effective_date", ParamType: "date"},
			"category":       {Column: "category", SysColumn: "category_system", ParamType: "token"},
			"value-quantity": {Column: "value_quantity", ParamType: "number"},
			"subject":        {Column: "patient_id", ParamType: "reference"},
			"patient":        {Column: "patient_id", ParamType: "reference"},
		}
	case "Condition":
		return map[string]FilterColumnMapping{
			"code":            {Column: "code_value", SysColumn: "code_system", ParamType: "token"},
			"clinical-status": {Column: "clinical_status", ParamType: "token"},
			"subject":         {Column: "patient_id", ParamType: "reference"},
			"patient":         {Column: "patient_id", ParamType: "reference"},
			"onset-date":      {Column: "onset_date", ParamType: "date"},
		}
	case "Encounter":
		return map[string]FilterColumnMapping{
			"status":  {Column: "status", ParamType: "token"},
			"class":   {Column: "class", ParamType: "token"},
			"date":    {Column: "period_start", ParamType: "date"},
			"subject": {Column: "patient_id", ParamType: "reference"},
			"patient": {Column: "patient_id", ParamType: "reference"},
			"type":    {Column: "type", ParamType: "token"},
		}
	case "MedicationRequest":
		return map[string]FilterColumnMapping{
			"status":     {Column: "status", ParamType: "token"},
			"intent":     {Column: "intent", ParamType: "token"},
			"medication": {Column: "medication_code", SysColumn: "medication_system", ParamType: "token"},
			"subject":    {Column: "patient_id", ParamType: "reference"},
			"patient":    {Column: "patient_id", ParamType: "reference"},
			"date":       {Column: "authored_on", ParamType: "date"},
		}
	default:
		return map[string]FilterColumnMapping{}
	}
}

// ---------------------------------------------------------------------------
// Expression to string
// ---------------------------------------------------------------------------

// FilterExpressionToString converts a filter expression tree back to its string form.
func FilterExpressionToString(expr *FilterExprNode) string {
	if expr == nil {
		return ""
	}

	switch expr.Type {
	case FilterExprParam:
		if expr.Operator == "pr" {
			return expr.Param + " pr"
		}
		return fmt.Sprintf("%s %s %q", expr.Param, expr.Operator, expr.Value)

	case FilterExprAnd:
		left := FilterExpressionToString(expr.Left)
		right := FilterExpressionToString(expr.Right)
		return left + " and " + right

	case FilterExprOr:
		left := FilterExpressionToString(expr.Left)
		right := FilterExpressionToString(expr.Right)
		return left + " or " + right

	case FilterExprNot:
		child := FilterExpressionToString(expr.Child)
		return "not " + child

	default:
		return ""
	}
}

// ---------------------------------------------------------------------------
// Simplification
// ---------------------------------------------------------------------------

// SimplifyFilterExpression optimizes a filter expression tree.
// Currently handles:
// - Double negation removal: NOT(NOT(x)) -> x
// - Recursive simplification of children.
func SimplifyFilterExpression(expr *FilterExprNode) *FilterExprNode {
	if expr == nil {
		return nil
	}

	switch expr.Type {
	case FilterExprParam:
		return expr

	case FilterExprAnd:
		expr.Left = SimplifyFilterExpression(expr.Left)
		expr.Right = SimplifyFilterExpression(expr.Right)
		return expr

	case FilterExprOr:
		expr.Left = SimplifyFilterExpression(expr.Left)
		expr.Right = SimplifyFilterExpression(expr.Right)
		return expr

	case FilterExprNot:
		expr.Child = SimplifyFilterExpression(expr.Child)
		// Double negation removal: NOT(NOT(x)) -> x
		if expr.Child != nil && expr.Child.Type == FilterExprNot {
			return expr.Child.Child
		}
		return expr

	default:
		return expr
	}
}

// ---------------------------------------------------------------------------
// Integration with SearchQuery
// ---------------------------------------------------------------------------

// ApplyFilterParam integrates a _filter expression into an existing SearchQuery.
// It parses the filter string, compiles it to SQL, and adds the resulting clause.
func ApplyFilterParam(q *SearchQuery, filterValue string, ctx *FilterContext) error {
	if filterValue == "" {
		return fmt.Errorf("empty filter expression")
	}

	expr, err := ParseFilterExpression(filterValue)
	if err != nil {
		return fmt.Errorf("invalid _filter expression: %w", err)
	}

	sql, args, err := CompileFilterToSQL(expr, ctx, q.Idx())
	if err != nil {
		return fmt.Errorf("failed to compile _filter to SQL: %w", err)
	}

	q.Add(sql, args...)
	return nil
}
