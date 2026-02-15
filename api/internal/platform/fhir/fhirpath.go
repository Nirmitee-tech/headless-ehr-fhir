package fhir

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ============================================================================
// FHIRPathEngine — public API
// ============================================================================

// FHIRPathEngine evaluates FHIRPath expressions against FHIR resources
// represented as map[string]interface{}.  It implements the subset of the
// FHIRPath specification required by FHIR R4 profiles, search parameter
// extraction, CQL, subscriptions and ABAC policies.
type FHIRPathEngine struct{}

// NewFHIRPathEngine creates a new FHIRPath evaluation engine.
func NewFHIRPathEngine() *FHIRPathEngine {
	return &FHIRPathEngine{}
}

// Evaluate evaluates a FHIRPath expression against a resource and returns the
// result as a collection (slice of interface{} values).  An empty collection
// is returned when the path resolves to nothing.
func (e *FHIRPathEngine) Evaluate(resource map[string]interface{}, expression string) ([]interface{}, error) {
	if resource == nil {
		return []interface{}{}, nil
	}
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return nil, fmt.Errorf("fhirpath: empty expression")
	}

	tokens, err := tokenize(expression)
	if err != nil {
		return nil, fmt.Errorf("fhirpath: tokenize: %w", err)
	}

	p := &parser{tokens: tokens}
	ast, err := p.parseExpression(0)
	if err != nil {
		return nil, fmt.Errorf("fhirpath: parse: %w", err)
	}
	if p.pos < len(p.tokens) {
		// There are leftover tokens — not necessarily an error for union
		// handled at the parse level, but truly unexpected tokens are.
		tok := p.tokens[p.pos]
		if tok.kind != tkEOF {
			return nil, fmt.Errorf("fhirpath: unexpected token %q at position %d", tok.value, tok.pos)
		}
	}

	ctx := &evalContext{resource: resource}
	result, err := ctx.eval(ast, []interface{}{resource})
	if err != nil {
		return nil, fmt.Errorf("fhirpath: eval: %w", err)
	}
	return result, nil
}

// EvaluateBool evaluates a FHIRPath expression and converts the result to a
// boolean following the FHIRPath singleton-evaluation rules:
//   - empty collection → false
//   - single boolean   → that boolean
//   - single non-nil   → true
//   - multiple items   → true (non-empty collection)
func (e *FHIRPathEngine) EvaluateBool(resource map[string]interface{}, expression string) (bool, error) {
	result, err := e.Evaluate(resource, expression)
	if err != nil {
		return false, err
	}
	return collectionToBool(result), nil
}

// EvaluateString evaluates a FHIRPath expression and returns the first result
// as a string.  Returns "" for an empty collection.
func (e *FHIRPathEngine) EvaluateString(resource map[string]interface{}, expression string) (string, error) {
	result, err := e.Evaluate(resource, expression)
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", nil
	}
	return fmt.Sprintf("%v", result[0]), nil
}

// ============================================================================
// Token types
// ============================================================================

type tokenKind int

const (
	tkIdent    tokenKind = iota // identifier or keyword
	tkNumber                    // integer or decimal
	tkString                    // 'single-quoted'
	tkDateTime                  // @2024-01-01 ...
	tkDot                       // .
	tkLParen                    // (
	tkRParen                    // )
	tkLBrack                    // [
	tkRBrack                    // ]
	tkComma                     // ,
	tkEq                        // =
	tkNe                        // !=
	tkLt                        // <
	tkGt                        // >
	tkLe                        // <=
	tkGe                        // >=
	tkPipe                      // |
	tkEOF                       // end-of-input
)

type token struct {
	kind  tokenKind
	value string
	pos   int
}

// ============================================================================
// Lexer / Tokenizer
// ============================================================================

func tokenize(input string) ([]token, error) {
	var tokens []token
	i := 0
	n := len(input)

	for i < n {
		ch := input[i]

		// skip whitespace
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}

		start := i

		switch {
		case ch == '.':
			tokens = append(tokens, token{tkDot, ".", start})
			i++
		case ch == '(':
			tokens = append(tokens, token{tkLParen, "(", start})
			i++
		case ch == ')':
			tokens = append(tokens, token{tkRParen, ")", start})
			i++
		case ch == '[':
			tokens = append(tokens, token{tkLBrack, "[", start})
			i++
		case ch == ']':
			tokens = append(tokens, token{tkRBrack, "]", start})
			i++
		case ch == ',':
			tokens = append(tokens, token{tkComma, ",", start})
			i++
		case ch == '|':
			tokens = append(tokens, token{tkPipe, "|", start})
			i++
		case ch == '=':
			tokens = append(tokens, token{tkEq, "=", start})
			i++
		case ch == '!':
			if i+1 < n && input[i+1] == '=' {
				tokens = append(tokens, token{tkNe, "!=", start})
				i += 2
			} else {
				return nil, fmt.Errorf("unexpected character '!' at position %d", start)
			}
		case ch == '<':
			if i+1 < n && input[i+1] == '=' {
				tokens = append(tokens, token{tkLe, "<=", start})
				i += 2
			} else {
				tokens = append(tokens, token{tkLt, "<", start})
				i++
			}
		case ch == '>':
			if i+1 < n && input[i+1] == '=' {
				tokens = append(tokens, token{tkGe, ">=", start})
				i += 2
			} else {
				tokens = append(tokens, token{tkGt, ">", start})
				i++
			}
		case ch == '\'':
			// string literal
			i++ // skip opening quote
			var sb strings.Builder
			for i < n && input[i] != '\'' {
				if input[i] == '\\' && i+1 < n {
					i++
					switch input[i] {
					case '\\':
						sb.WriteByte('\\')
					case '\'':
						sb.WriteByte('\'')
					case 'n':
						sb.WriteByte('\n')
					case 't':
						sb.WriteByte('\t')
					default:
						sb.WriteByte(input[i])
					}
				} else {
					sb.WriteByte(input[i])
				}
				i++
			}
			if i >= n {
				return nil, fmt.Errorf("unterminated string at position %d", start)
			}
			i++ // skip closing quote
			tokens = append(tokens, token{tkString, sb.String(), start})
		case ch == '@':
			// datetime literal  @YYYY-MM-DD or @YYYY-MM-DDTHH:MM:SS...
			i++ // skip @
			j := i
			for j < n && (input[j] == '-' || input[j] == ':' || input[j] == 'T' ||
				input[j] == '+' || input[j] == 'Z' || (input[j] >= '0' && input[j] <= '9') || input[j] == '.') {
				j++
			}
			tokens = append(tokens, token{tkDateTime, input[i:j], start})
			i = j
		case ch == '-' || (ch >= '0' && ch <= '9'):
			// number (possibly negative)
			j := i
			if ch == '-' {
				j++
			}
			for j < n && input[j] >= '0' && input[j] <= '9' {
				j++
			}
			if j < n && input[j] == '.' {
				// Could be a decimal OR a dot-navigation after a number.
				// Look ahead: if the next char after '.' is a digit, it's a decimal.
				if j+1 < n && input[j+1] >= '0' && input[j+1] <= '9' {
					j++ // skip .
					for j < n && input[j] >= '0' && input[j] <= '9' {
						j++
					}
				}
			}
			tokens = append(tokens, token{tkNumber, input[i:j], start})
			i = j
		case ch == '_' || unicode.IsLetter(rune(ch)):
			j := i
			for j < n && (input[j] == '_' || unicode.IsLetter(rune(input[j])) || unicode.IsDigit(rune(input[j]))) {
				j++
			}
			tokens = append(tokens, token{tkIdent, input[i:j], start})
			i = j
		default:
			return nil, fmt.Errorf("unexpected character %q at position %d", string(ch), start)
		}
	}

	tokens = append(tokens, token{tkEOF, "", n})
	return tokens, nil
}

// ============================================================================
// AST node types
// ============================================================================

type nodeKind int

const (
	ndLiteral    nodeKind = iota // string, number, bool, datetime
	ndPath                       // identifier (field name or resource type)
	ndDot                        // a.b
	ndIndex                      // a[n]
	ndFunction                   // a.fn(args...)
	ndCompare                    // a op b  (=, !=, <, >, <=, >=)
	ndAnd                        // a and b
	ndOr                         // a or b
	ndImplies                    // a implies b
	ndUnion                      // a | b
	ndNegate                     // unary minus
)

type astNode struct {
	kind     nodeKind
	value    interface{} // literal value, or identifier name, or operator string
	children []*astNode  // operands / arguments
}

// ============================================================================
// Parser — recursive descent
// ============================================================================

type parser struct {
	tokens []token
	pos    int
}

func (p *parser) peek() token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return token{kind: tkEOF, pos: -1}
}

func (p *parser) advance() token {
	t := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return t
}

func (p *parser) expect(kind tokenKind) (token, error) {
	t := p.advance()
	if t.kind != kind {
		return t, fmt.Errorf("expected token kind %d but got %q at position %d", kind, t.value, t.pos)
	}
	return t, nil
}

// Operator precedence (lowest to highest):
//   implies  (1)
//   or       (2)
//   and      (3)
//   |        (4)  — union
//   = != < > <= >= (5)
//   unary    (6)
//   . [] ()  (7)

func (p *parser) parseExpression(minPrec int) (*astNode, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		tok := p.peek()
		prec, kind, opValue := p.infixInfo(tok)
		if prec < minPrec {
			break
		}
		p.advance()
		right, err := p.parseExpression(prec + 1)
		if err != nil {
			return nil, err
		}
		node := &astNode{kind: kind, children: []*astNode{left, right}}
		if kind == ndCompare {
			node.value = opValue
		}
		left = node
	}
	return left, nil
}

func (p *parser) infixInfo(tok token) (int, nodeKind, string) {
	switch {
	case tok.kind == tkIdent && tok.value == "implies":
		return 1, ndImplies, "implies"
	case tok.kind == tkIdent && tok.value == "or":
		return 2, ndOr, "or"
	case tok.kind == tkIdent && tok.value == "and":
		return 3, ndAnd, "and"
	case tok.kind == tkPipe:
		return 4, ndUnion, "|"
	case tok.kind == tkEq:
		return 5, ndCompare, "="
	case tok.kind == tkNe:
		return 5, ndCompare, "!="
	case tok.kind == tkLt:
		return 5, ndCompare, "<"
	case tok.kind == tkGt:
		return 5, ndCompare, ">"
	case tok.kind == tkLe:
		return 5, ndCompare, "<="
	case tok.kind == tkGe:
		return 5, ndCompare, ">="
	}
	return -1, 0, ""
}

func (p *parser) parseUnary() (*astNode, error) {
	return p.parsePostfix()
}

func (p *parser) parsePostfix() (*astNode, error) {
	node, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.peek()
		if tok.kind == tkDot {
			p.advance() // consume '.'
			next := p.peek()
			if next.kind != tkIdent {
				return nil, fmt.Errorf("expected identifier after '.' at position %d", next.pos)
			}
			ident := p.advance()

			// Check if this is a function call: ident(
			if p.peek().kind == tkLParen {
				p.advance() // consume '('
				args, err := p.parseArgList()
				if err != nil {
					return nil, err
				}
				_, err = p.expect(tkRParen)
				if err != nil {
					return nil, err
				}
				node = &astNode{
					kind:     ndFunction,
					value:    ident.value,
					children: append([]*astNode{node}, args...),
				}
			} else {
				// Field access: node.field
				right := &astNode{kind: ndPath, value: ident.value}
				node = &astNode{kind: ndDot, children: []*astNode{node, right}}
			}
		} else if tok.kind == tkLBrack {
			p.advance() // consume '['
			idxTok, err := p.expect(tkNumber)
			if err != nil {
				return nil, fmt.Errorf("expected number in index at position %d", tok.pos)
			}
			_, err = p.expect(tkRBrack)
			if err != nil {
				return nil, err
			}
			idx, _ := strconv.ParseInt(idxTok.value, 10, 64)
			node = &astNode{
				kind:  ndIndex,
				value: idx,
				children: []*astNode{node},
			}
		} else {
			break
		}
	}
	return node, nil
}

func (p *parser) parsePrimary() (*astNode, error) {
	tok := p.peek()

	switch tok.kind {
	case tkLParen:
		p.advance()
		// Check for unary minus inside parens: (-5)
		if p.peek().kind == tkNumber {
			numTok := p.peek()
			if strings.HasPrefix(numTok.value, "-") {
				// Already a negative number literal
				inner, err := p.parseExpression(0)
				if err != nil {
					return nil, err
				}
				_, err = p.expect(tkRParen)
				if err != nil {
					return nil, err
				}
				return inner, nil
			}
		}
		inner, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		_, err = p.expect(tkRParen)
		if err != nil {
			return nil, err
		}
		return inner, nil

	case tkString:
		p.advance()
		return &astNode{kind: ndLiteral, value: tok.value}, nil

	case tkNumber:
		p.advance()
		if strings.Contains(tok.value, ".") {
			f, err := strconv.ParseFloat(tok.value, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid decimal %q at position %d", tok.value, tok.pos)
			}
			return &astNode{kind: ndLiteral, value: f}, nil
		}
		i, err := strconv.ParseInt(tok.value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q at position %d", tok.value, tok.pos)
		}
		return &astNode{kind: ndLiteral, value: i}, nil

	case tkDateTime:
		p.advance()
		t, err := parseDateTimeLiteral(tok.value)
		if err != nil {
			return nil, fmt.Errorf("invalid datetime %q at position %d: %w", tok.value, tok.pos, err)
		}
		return &astNode{kind: ndLiteral, value: t}, nil

	case tkIdent:
		p.advance()
		name := tok.value

		// Boolean literals
		if name == "true" {
			return &astNode{kind: ndLiteral, value: true}, nil
		}
		if name == "false" {
			return &astNode{kind: ndLiteral, value: false}, nil
		}

		// Standalone function calls: now(), today(), iif(...)
		if p.peek().kind == tkLParen {
			p.advance() // consume '('
			args, err := p.parseArgList()
			if err != nil {
				return nil, err
			}
			_, err = p.expect(tkRParen)
			if err != nil {
				return nil, err
			}
			// For standalone functions, the implicit input is nil (no receiver).
			return &astNode{
				kind:     ndFunction,
				value:    name,
				children: args,
			}, nil
		}

		return &astNode{kind: ndPath, value: name}, nil

	case tkEOF:
		return nil, fmt.Errorf("unexpected end of expression")

	default:
		return nil, fmt.Errorf("unexpected token %q at position %d", tok.value, tok.pos)
	}
}

func (p *parser) parseArgList() ([]*astNode, error) {
	var args []*astNode
	if p.peek().kind == tkRParen {
		return args, nil
	}
	for {
		arg, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if p.peek().kind != tkComma {
			break
		}
		p.advance() // consume ','
	}
	return args, nil
}

// ============================================================================
// Evaluator
// ============================================================================

type evalContext struct {
	resource map[string]interface{}
}

// eval evaluates an AST node against an input collection and returns a result
// collection.
func (ctx *evalContext) eval(node *astNode, input []interface{}) ([]interface{}, error) {
	if node == nil {
		return input, nil
	}
	switch node.kind {
	case ndLiteral:
		return []interface{}{node.value}, nil

	case ndPath:
		return ctx.evalPath(node, input)

	case ndDot:
		left, err := ctx.eval(node.children[0], input)
		if err != nil {
			return nil, err
		}
		return ctx.eval(node.children[1], left)

	case ndIndex:
		coll, err := ctx.eval(node.children[0], input)
		if err != nil {
			return nil, err
		}
		idx := node.value.(int64)
		coll = flattenCollection(coll)
		if int(idx) < 0 || int(idx) >= len(coll) {
			return []interface{}{}, nil
		}
		return []interface{}{coll[int(idx)]}, nil

	case ndFunction:
		return ctx.evalFunction(node, input)

	case ndCompare:
		return ctx.evalCompare(node, input)

	case ndAnd:
		return ctx.evalAnd(node, input)

	case ndOr:
		return ctx.evalOr(node, input)

	case ndImplies:
		return ctx.evalImplies(node, input)

	case ndUnion:
		return ctx.evalUnion(node, input)

	default:
		return nil, fmt.Errorf("unknown node kind %d", node.kind)
	}
}

// evalPath resolves an identifier against the input collection.
func (ctx *evalContext) evalPath(node *astNode, input []interface{}) ([]interface{}, error) {
	name := node.value.(string)

	// Check if this is a FHIR resource type name that matches the root resource.
	if isResourceTypeName(name) {
		rt, _ := ctx.resource["resourceType"].(string)
		if rt == name {
			return []interface{}{ctx.resource}, nil
		}
		// Resource type mismatch — return empty.
		return []interface{}{}, nil
	}

	// Navigate into each item in the input collection.
	var result []interface{}
	for _, item := range input {
		result = append(result, navigateField(item, name)...)
	}
	return result, nil
}

// navigateField extracts a named field from a value.
func navigateField(item interface{}, field string) []interface{} {
	switch v := item.(type) {
	case map[string]interface{}:
		val, ok := v[field]
		if !ok {
			return nil
		}
		if arr, isArr := val.([]interface{}); isArr {
			return arr
		}
		return []interface{}{val}
	default:
		return nil
	}
}

// flattenCollection flattens nested slices in a collection.
func flattenCollection(coll []interface{}) []interface{} {
	var out []interface{}
	for _, item := range coll {
		if arr, ok := item.([]interface{}); ok {
			out = append(out, arr...)
		} else {
			out = append(out, item)
		}
	}
	return out
}

// ============================================================================
// Comparison
// ============================================================================

func (ctx *evalContext) evalCompare(node *astNode, input []interface{}) ([]interface{}, error) {
	// The parser stores the comparison operator ("=", "!=", "<", etc.) in
	// node.value when constructing ndCompare nodes.
	op, _ := node.value.(string)
	if op == "" {
		return nil, fmt.Errorf("comparison node missing operator")
	}

	leftColl, err := ctx.eval(node.children[0], input)
	if err != nil {
		return nil, err
	}
	rightColl, err := ctx.eval(node.children[1], input)
	if err != nil {
		return nil, err
	}

	// FHIRPath comparison: if either side is empty, result is empty.
	if len(leftColl) == 0 || len(rightColl) == 0 {
		return []interface{}{}, nil
	}

	lv := leftColl[0]
	rv := rightColl[0]

	result, err := compareValues(lv, rv, op)
	if err != nil {
		return nil, err
	}
	return []interface{}{result}, nil
}

func compareValues(lv, rv interface{}, op string) (bool, error) {
	// Normalize numeric types for comparison.
	ln, lok := toNumber(lv)
	rn, rok := toNumber(rv)
	if lok && rok {
		return compareNumbers(ln, rn, op), nil
	}

	// Boolean comparison
	lb, lbOk := lv.(bool)
	rb, rbOk := rv.(bool)
	if lbOk && rbOk {
		switch op {
		case "=":
			return lb == rb, nil
		case "!=":
			return lb != rb, nil
		}
		return false, nil
	}

	// Time comparison
	lt, ltOk := lv.(time.Time)
	rt, rtOk := rv.(time.Time)
	if ltOk && rtOk {
		return compareTimes(lt, rt, op), nil
	}

	// String comparison (default)
	ls := fmt.Sprintf("%v", lv)
	rs := fmt.Sprintf("%v", rv)

	switch op {
	case "=":
		return ls == rs, nil
	case "!=":
		return ls != rs, nil
	case "<":
		return ls < rs, nil
	case ">":
		return ls > rs, nil
	case "<=":
		return ls <= rs, nil
	case ">=":
		return ls >= rs, nil
	}
	return false, fmt.Errorf("unknown comparison operator %q", op)
}

func toNumber(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	case json_number:
		f, err := strconv.ParseFloat(string(n), 64)
		return f, err == nil
	}
	return 0, false
}

// json_number is a type alias to handle json.Number if present.
type json_number = string

func compareNumbers(l, r float64, op string) bool {
	switch op {
	case "=":
		return l == r
	case "!=":
		return l != r
	case "<":
		return l < r
	case ">":
		return l > r
	case "<=":
		return l <= r
	case ">=":
		return l >= r
	}
	return false
}

func compareTimes(l, r time.Time, op string) bool {
	switch op {
	case "=":
		return l.Equal(r)
	case "!=":
		return !l.Equal(r)
	case "<":
		return l.Before(r)
	case ">":
		return l.After(r)
	case "<=":
		return !l.After(r)
	case ">=":
		return !l.Before(r)
	}
	return false
}

// ============================================================================
// Logical operators
// ============================================================================

func (ctx *evalContext) evalAnd(node *astNode, input []interface{}) ([]interface{}, error) {
	leftColl, err := ctx.eval(node.children[0], input)
	if err != nil {
		return nil, err
	}
	lb := collectionToBool(leftColl)
	if !lb {
		return []interface{}{false}, nil // short-circuit
	}
	rightColl, err := ctx.eval(node.children[1], input)
	if err != nil {
		return nil, err
	}
	return []interface{}{collectionToBool(rightColl)}, nil
}

func (ctx *evalContext) evalOr(node *astNode, input []interface{}) ([]interface{}, error) {
	leftColl, err := ctx.eval(node.children[0], input)
	if err != nil {
		return nil, err
	}
	lb := collectionToBool(leftColl)
	if lb {
		return []interface{}{true}, nil // short-circuit
	}
	rightColl, err := ctx.eval(node.children[1], input)
	if err != nil {
		return nil, err
	}
	return []interface{}{collectionToBool(rightColl)}, nil
}

func (ctx *evalContext) evalImplies(node *astNode, input []interface{}) ([]interface{}, error) {
	leftColl, err := ctx.eval(node.children[0], input)
	if err != nil {
		return nil, err
	}
	lb := collectionToBool(leftColl)
	if !lb {
		return []interface{}{true}, nil // false implies anything is true
	}
	rightColl, err := ctx.eval(node.children[1], input)
	if err != nil {
		return nil, err
	}
	return []interface{}{collectionToBool(rightColl)}, nil
}

// ============================================================================
// Union
// ============================================================================

func (ctx *evalContext) evalUnion(node *astNode, input []interface{}) ([]interface{}, error) {
	leftColl, err := ctx.eval(node.children[0], input)
	if err != nil {
		return nil, err
	}
	rightColl, err := ctx.eval(node.children[1], input)
	if err != nil {
		return nil, err
	}
	// Deduplicate
	seen := make(map[string]bool)
	var result []interface{}
	for _, v := range append(leftColl, rightColl...) {
		key := fmt.Sprintf("%v", v)
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}
	return result, nil
}

// ============================================================================
// Function evaluation
// ============================================================================

func (ctx *evalContext) evalFunction(node *astNode, input []interface{}) ([]interface{}, error) {
	name := node.value.(string)

	// Standalone functions (no receiver): now(), today(), iif(...)
	if len(node.children) == 0 || (len(node.children) > 0 && isStandaloneFunction(name)) {
		return ctx.evalStandaloneFunction(name, node.children, input)
	}

	// Method-style: receiver.fn(args...)
	// children[0] is the receiver, children[1:] are the arguments.
	receiver := node.children[0]
	args := node.children[1:]

	receiverColl, err := ctx.eval(receiver, input)
	if err != nil {
		return nil, err
	}

	switch name {
	// Collection functions
	case "where":
		return ctx.fnWhere(receiverColl, args)
	case "exists":
		return ctx.fnExists(receiverColl, args)
	case "all":
		return ctx.fnAll(receiverColl, args)
	case "count":
		return []interface{}{int64(len(receiverColl))}, nil
	case "first":
		if len(receiverColl) == 0 {
			return []interface{}{}, nil
		}
		return []interface{}{receiverColl[0]}, nil
	case "last":
		if len(receiverColl) == 0 {
			return []interface{}{}, nil
		}
		return []interface{}{receiverColl[len(receiverColl)-1]}, nil
	case "tail":
		if len(receiverColl) <= 1 {
			return []interface{}{}, nil
		}
		return receiverColl[1:], nil
	case "empty":
		return []interface{}{len(receiverColl) == 0}, nil
	case "distinct":
		return ctx.fnDistinct(receiverColl), nil
	case "select":
		return ctx.fnSelect(receiverColl, args)
	case "ofType":
		return ctx.fnOfType(receiverColl, args)
	case "hasValue":
		return []interface{}{len(receiverColl) == 1 && receiverColl[0] != nil}, nil
	case "not":
		b := collectionToBool(receiverColl)
		return []interface{}{!b}, nil

	// String functions
	case "startsWith":
		return ctx.fnStringPredicate(receiverColl, args, strings.HasPrefix)
	case "endsWith":
		return ctx.fnStringPredicate(receiverColl, args, strings.HasSuffix)
	case "contains":
		return ctx.fnStringPredicate(receiverColl, args, strings.Contains)
	case "matches":
		return ctx.fnMatches(receiverColl, args)
	case "length":
		return ctx.fnLength(receiverColl)
	case "upper":
		return ctx.fnStringTransform(receiverColl, strings.ToUpper)
	case "lower":
		return ctx.fnStringTransform(receiverColl, strings.ToLower)
	case "replace":
		return ctx.fnReplace(receiverColl, args, input)
	case "substring":
		return ctx.fnSubstring(receiverColl, args, input)

	// Type functions
	case "is":
		return ctx.fnIs(receiverColl, args)
	case "as":
		return ctx.fnAs(receiverColl, args)

	// Math functions
	case "abs":
		return ctx.fnMathUnary(receiverColl, math.Abs)
	case "ceiling":
		return ctx.fnMathUnary(receiverColl, math.Ceil)
	case "floor":
		return ctx.fnMathUnary(receiverColl, math.Floor)
	case "round":
		return ctx.fnMathUnary(receiverColl, math.Round)

	// Date/time functions
	case "toDate":
		return ctx.fnToDate(receiverColl)
	case "toDateTime":
		return ctx.fnToDateTime(receiverColl)

	default:
		return nil, fmt.Errorf("unknown function %q", name)
	}
}

func isStandaloneFunction(name string) bool {
	switch name {
	case "now", "today", "iif":
		return true
	}
	return false
}

func (ctx *evalContext) evalStandaloneFunction(name string, args []*astNode, input []interface{}) ([]interface{}, error) {
	switch name {
	case "now":
		return []interface{}{time.Now().UTC()}, nil
	case "today":
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return []interface{}{today}, nil
	case "iif":
		return ctx.fnIif(args, input)
	}
	return nil, fmt.Errorf("unknown standalone function %q", name)
}

// ============================================================================
// Collection function implementations
// ============================================================================

func (ctx *evalContext) fnWhere(coll []interface{}, args []*astNode) ([]interface{}, error) {
	if len(args) == 0 {
		return coll, nil
	}
	expr := args[0]
	var result []interface{}
	for _, item := range coll {
		itemColl := []interface{}{item}
		val, err := ctx.eval(expr, itemColl)
		if err != nil {
			return nil, err
		}
		if collectionToBool(val) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (ctx *evalContext) fnExists(coll []interface{}, args []*astNode) ([]interface{}, error) {
	if len(args) == 0 {
		return []interface{}{len(coll) > 0}, nil
	}
	// exists(expr) — true if any item matches
	expr := args[0]
	for _, item := range coll {
		itemColl := []interface{}{item}
		val, err := ctx.eval(expr, itemColl)
		if err != nil {
			return nil, err
		}
		if collectionToBool(val) {
			return []interface{}{true}, nil
		}
	}
	return []interface{}{false}, nil
}

func (ctx *evalContext) fnAll(coll []interface{}, args []*astNode) ([]interface{}, error) {
	if len(args) == 0 {
		return []interface{}{true}, nil
	}
	expr := args[0]
	for _, item := range coll {
		itemColl := []interface{}{item}
		val, err := ctx.eval(expr, itemColl)
		if err != nil {
			return nil, err
		}
		if !collectionToBool(val) {
			return []interface{}{false}, nil
		}
	}
	return []interface{}{true}, nil
}

func (ctx *evalContext) fnDistinct(coll []interface{}) []interface{} {
	seen := make(map[string]bool)
	var result []interface{}
	for _, v := range coll {
		key := fmt.Sprintf("%v", v)
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}
	return result
}

func (ctx *evalContext) fnSelect(coll []interface{}, args []*astNode) ([]interface{}, error) {
	if len(args) == 0 {
		return coll, nil
	}
	expr := args[0]
	var result []interface{}
	for _, item := range coll {
		itemColl := []interface{}{item}
		val, err := ctx.eval(expr, itemColl)
		if err != nil {
			return nil, err
		}
		result = append(result, val...)
	}
	return result, nil
}

func (ctx *evalContext) fnOfType(coll []interface{}, args []*astNode) ([]interface{}, error) {
	if len(args) == 0 {
		return coll, nil
	}
	typeName := ""
	if args[0].kind == ndPath {
		typeName = args[0].value.(string)
	} else if args[0].kind == ndLiteral {
		typeName = fmt.Sprintf("%v", args[0].value)
	}

	var result []interface{}
	for _, item := range coll {
		if matchesType(item, typeName) {
			result = append(result, item)
		}
	}
	return result, nil
}

// ============================================================================
// String function implementations
// ============================================================================

func (ctx *evalContext) fnStringPredicate(coll []interface{}, args []*astNode, fn func(string, string) bool) ([]interface{}, error) {
	if len(coll) == 0 || len(args) == 0 {
		return []interface{}{}, nil
	}
	argColl, err := ctx.eval(args[0], coll)
	if err != nil {
		return nil, err
	}
	if len(argColl) == 0 {
		return []interface{}{}, nil
	}
	s := fmt.Sprintf("%v", coll[0])
	arg := fmt.Sprintf("%v", argColl[0])
	return []interface{}{fn(s, arg)}, nil
}

func (ctx *evalContext) fnMatches(coll []interface{}, args []*astNode) ([]interface{}, error) {
	if len(coll) == 0 || len(args) == 0 {
		return []interface{}{}, nil
	}
	argColl, err := ctx.eval(args[0], coll)
	if err != nil {
		return nil, err
	}
	if len(argColl) == 0 {
		return []interface{}{}, nil
	}
	s := fmt.Sprintf("%v", coll[0])
	pattern := fmt.Sprintf("%v", argColl[0])
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex %q: %w", pattern, err)
	}
	return []interface{}{re.MatchString(s)}, nil
}

func (ctx *evalContext) fnLength(coll []interface{}) ([]interface{}, error) {
	if len(coll) == 0 {
		return []interface{}{}, nil
	}
	s := fmt.Sprintf("%v", coll[0])
	return []interface{}{int64(len(s))}, nil
}

func (ctx *evalContext) fnStringTransform(coll []interface{}, fn func(string) string) ([]interface{}, error) {
	if len(coll) == 0 {
		return []interface{}{}, nil
	}
	s := fmt.Sprintf("%v", coll[0])
	return []interface{}{fn(s)}, nil
}

func (ctx *evalContext) fnReplace(coll []interface{}, args []*astNode, input []interface{}) ([]interface{}, error) {
	if len(coll) == 0 || len(args) < 2 {
		return []interface{}{}, nil
	}
	patternColl, err := ctx.eval(args[0], input)
	if err != nil {
		return nil, err
	}
	replacementColl, err := ctx.eval(args[1], input)
	if err != nil {
		return nil, err
	}
	if len(patternColl) == 0 || len(replacementColl) == 0 {
		return coll, nil
	}
	s := fmt.Sprintf("%v", coll[0])
	pattern := fmt.Sprintf("%v", patternColl[0])
	replacement := fmt.Sprintf("%v", replacementColl[0])
	return []interface{}{strings.ReplaceAll(s, pattern, replacement)}, nil
}

func (ctx *evalContext) fnSubstring(coll []interface{}, args []*astNode, input []interface{}) ([]interface{}, error) {
	if len(coll) == 0 || len(args) == 0 {
		return []interface{}{}, nil
	}
	startColl, err := ctx.eval(args[0], input)
	if err != nil {
		return nil, err
	}
	if len(startColl) == 0 {
		return []interface{}{}, nil
	}
	s := fmt.Sprintf("%v", coll[0])
	startF, ok := toNumber(startColl[0])
	if !ok {
		return []interface{}{}, nil
	}
	start := int(startF)
	if start < 0 {
		start = 0
	}
	if start >= len(s) {
		return []interface{}{""}, nil
	}

	if len(args) >= 2 {
		lenColl, err := ctx.eval(args[1], input)
		if err != nil {
			return nil, err
		}
		if len(lenColl) > 0 {
			lenF, ok := toNumber(lenColl[0])
			if ok {
				end := start + int(lenF)
				if end > len(s) {
					end = len(s)
				}
				return []interface{}{s[start:end]}, nil
			}
		}
	}

	return []interface{}{s[start:]}, nil
}

// ============================================================================
// Type function implementations
// ============================================================================

func (ctx *evalContext) fnIs(coll []interface{}, args []*astNode) ([]interface{}, error) {
	if len(coll) == 0 || len(args) == 0 {
		return []interface{}{false}, nil
	}
	typeName := ""
	if args[0].kind == ndPath {
		typeName = args[0].value.(string)
	}
	return []interface{}{matchesType(coll[0], typeName)}, nil
}

func (ctx *evalContext) fnAs(coll []interface{}, args []*astNode) ([]interface{}, error) {
	if len(coll) == 0 || len(args) == 0 {
		return []interface{}{}, nil
	}
	typeName := ""
	if args[0].kind == ndPath {
		typeName = args[0].value.(string)
	}
	var result []interface{}
	for _, item := range coll {
		if matchesType(item, typeName) {
			result = append(result, item)
		}
	}
	return result, nil
}

func matchesType(v interface{}, typeName string) bool {
	switch strings.ToLower(typeName) {
	case "string":
		_, ok := v.(string)
		return ok
	case "integer", "int":
		switch v.(type) {
		case int, int64, int32:
			return true
		}
		return false
	case "decimal", "float":
		_, ok := v.(float64)
		return ok
	case "boolean", "bool":
		_, ok := v.(bool)
		return ok
	case "date", "datetime":
		_, ok := v.(time.Time)
		return ok
	default:
		// Check if it's a FHIR resource type.
		if m, ok := v.(map[string]interface{}); ok {
			rt, _ := m["resourceType"].(string)
			return rt == typeName
		}
		return false
	}
}

// ============================================================================
// Math function implementations
// ============================================================================

func (ctx *evalContext) fnMathUnary(coll []interface{}, fn func(float64) float64) ([]interface{}, error) {
	if len(coll) == 0 {
		return []interface{}{}, nil
	}
	f, ok := toNumber(coll[0])
	if !ok {
		return []interface{}{}, nil
	}
	result := fn(f)
	// Return as int64 if it's a whole number.
	if result == math.Trunc(result) && !math.IsInf(result, 0) && !math.IsNaN(result) {
		return []interface{}{int64(result)}, nil
	}
	return []interface{}{result}, nil
}

// ============================================================================
// Date/time function implementations
// ============================================================================

func (ctx *evalContext) fnToDate(coll []interface{}) ([]interface{}, error) {
	if len(coll) == 0 {
		return []interface{}{}, nil
	}
	s, ok := coll[0].(string)
	if !ok {
		if t, tok := coll[0].(time.Time); tok {
			return []interface{}{t}, nil
		}
		return []interface{}{}, nil
	}
	t, err := parseDateTimeLiteral(s)
	if err != nil {
		return []interface{}{}, nil
	}
	return []interface{}{t}, nil
}

func (ctx *evalContext) fnToDateTime(coll []interface{}) ([]interface{}, error) {
	if len(coll) == 0 {
		return []interface{}{}, nil
	}
	s, ok := coll[0].(string)
	if !ok {
		if t, tok := coll[0].(time.Time); tok {
			return []interface{}{t}, nil
		}
		return []interface{}{}, nil
	}
	t, err := parseDateTimeLiteral(s)
	if err != nil {
		return []interface{}{}, nil
	}
	return []interface{}{t}, nil
}

// ============================================================================
// iif function
// ============================================================================

func (ctx *evalContext) fnIif(args []*astNode, input []interface{}) ([]interface{}, error) {
	if len(args) < 2 {
		return []interface{}{}, nil
	}

	condColl, err := ctx.eval(args[0], input)
	if err != nil {
		return nil, err
	}
	cond := collectionToBool(condColl)

	if cond {
		return ctx.eval(args[1], input)
	}
	if len(args) >= 3 {
		return ctx.eval(args[2], input)
	}
	return []interface{}{}, nil
}

// ============================================================================
// Utility functions
// ============================================================================

// collectionToBool converts a FHIRPath collection to a boolean value following
// the FHIRPath singleton-evaluation of collections to booleans:
//   - empty → false
//   - single boolean → that boolean value
//   - single non-boolean non-nil → true
//   - multiple items → true (non-empty)
func collectionToBool(coll []interface{}) bool {
	if len(coll) == 0 {
		return false
	}
	if len(coll) == 1 {
		switch v := coll[0].(type) {
		case bool:
			return v
		case nil:
			return false
		default:
			return true
		}
	}
	return true // non-empty collection
}

// isResourceTypeName returns true if the name looks like a FHIR resource type
// (starts with uppercase).
func isResourceTypeName(name string) bool {
	if len(name) == 0 {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}

// parseDateTimeLiteral parses various date/datetime string formats.
func parseDateTimeLiteral(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04Z",
		"2006-01-02T15:04",
		"2006-01-02",
		"2006-01",
		"2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse datetime %q", s)
}
