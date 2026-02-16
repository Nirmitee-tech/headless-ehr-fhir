package fhir

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// FullTextConfig describes how full-text search is configured for a resource type.
type FullTextConfig struct {
	ResourceType string
	TextColumns  []string          // Columns to include in text search
	WeightMap    map[string]string // Column -> weight (A, B, C, D) for ranking
	Language     string            // PostgreSQL text search config (default: "english")
	IndexName    string            // Name of the GIN index
}

// FullTextQuery represents a parsed full-text search query.
type FullTextQuery struct {
	RawQuery      string
	TSQuery       string // PostgreSQL tsquery expression
	Language      string
	UseRanking    bool              // Whether to order by relevance
	MinRank       float64           // Minimum relevance threshold (0.0-1.0)
	HighlightOpts *HighlightOptions
}

// HighlightOptions controls ts_headline output.
type HighlightOptions struct {
	MaxWords     int
	MinWords     int
	MaxFragments int
	StartSel     string
	StopSel      string
}

// FullTextResult extends a search result with relevance information.
type FullTextResult struct {
	Rank      float64
	Headline  string   // Highlighted snippet
	MatchedOn []string // Which fields matched
}

// FullTextSearchEngine provides full-text search capabilities.
type FullTextSearchEngine struct {
	Configs map[string]*FullTextConfig // resource type -> config
}

// ---------------------------------------------------------------------------
// Engine construction
// ---------------------------------------------------------------------------

// NewFullTextSearchEngine creates an engine pre-loaded with default configs.
func NewFullTextSearchEngine() *FullTextSearchEngine {
	return &FullTextSearchEngine{
		Configs: DefaultFullTextConfigs(),
	}
}

// RegisterConfig adds or replaces a resource type configuration.
func (e *FullTextSearchEngine) RegisterConfig(config *FullTextConfig) {
	e.Configs[config.ResourceType] = config
}

// ---------------------------------------------------------------------------
// Query parsing
// ---------------------------------------------------------------------------

// ParseFullTextQuery parses a FHIR _text or _content value into a FullTextQuery.
func ParseFullTextQuery(raw string, language string) (*FullTextQuery, error) {
	if language == "" {
		language = "english"
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("full-text search query must not be empty")
	}

	// Check if the query has any alphanumeric characters
	hasAlphaNum := false
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r > 127 {
			hasAlphaNum = true
			break
		}
	}
	if !hasAlphaNum {
		return nil, fmt.Errorf("full-text search query must contain at least one word")
	}

	tsQuery := buildTSQueryFromInput(trimmed, language)

	return &FullTextQuery{
		RawQuery:   raw,
		TSQuery:    tsQuery,
		Language:   language,
		UseRanking: true,
		MinRank:    0.0,
		HighlightOpts: &HighlightOptions{
			MaxWords:     35,
			MinWords:     15,
			MaxFragments: 3,
			StartSel:     "<b>",
			StopSel:      "</b>",
		},
	}, nil
}

// buildTSQueryFromInput converts user input into a PostgreSQL tsquery expression.
// Supports:
//   - Simple words: "diabetes" -> plainto_tsquery
//   - Phrases in quotes: "type 2 diabetes" -> phraseto_tsquery
//   - Prefix matching: diab* -> 'diab':*
//   - AND (+), OR (|), NOT (-) operators
func buildTSQueryFromInput(input string, language string) string {
	input = strings.TrimSpace(input)

	// Check if the entire input is a quoted phrase
	if strings.HasPrefix(input, `"`) && strings.HasSuffix(input, `"`) && len(input) > 2 {
		phrase := input[1 : len(input)-1]
		escaped := EscapeFullTextQuery(phrase)
		return fmt.Sprintf("phraseto_tsquery('%s', '%s')", language, escaped)
	}

	// Check for operators (+, -, |)
	hasOperators := false
	for _, r := range input {
		if r == '+' || r == '-' {
			hasOperators = true
			break
		}
	}
	// Check for pipe-based OR
	if strings.Contains(input, "|") {
		hasOperators = true
	}

	// Check for prefix matching (word*)
	if strings.Contains(input, "*") && !hasOperators {
		return buildPrefixTSQuery(input, language)
	}

	if hasOperators {
		return buildOperatorTSQuery(input, language)
	}

	// Default: plain text search (implicit AND between words)
	escaped := EscapeFullTextQuery(input)
	return fmt.Sprintf("plainto_tsquery('%s', '%s')", language, escaped)
}

// buildPrefixTSQuery handles prefix matching like "diab*".
func buildPrefixTSQuery(input string, language string) string {
	terms := strings.Fields(input)
	var parts []string
	for _, term := range terms {
		if strings.HasSuffix(term, "*") {
			base := strings.TrimSuffix(term, "*")
			escaped := EscapeFullTextQuery(base)
			parts = append(parts, fmt.Sprintf("'%s':*", escaped))
		} else {
			escaped := EscapeFullTextQuery(term)
			parts = append(parts, fmt.Sprintf("plainto_tsquery('%s', '%s')", language, escaped))
		}
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts, " && ")
}

// buildOperatorTSQuery handles +, -, | operators in search input.
func buildOperatorTSQuery(input string, language string) string {
	terms := SplitSearchTerms(input)
	var parts []string

	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}

		if strings.HasPrefix(term, "+") {
			word := strings.TrimPrefix(term, "+")
			word = strings.TrimSpace(word)
			if word == "" {
				continue
			}
			escaped := EscapeFullTextQuery(word)
			parts = append(parts, fmt.Sprintf("plainto_tsquery('%s', '%s')", language, escaped))
		} else if strings.HasPrefix(term, "-") {
			word := strings.TrimPrefix(term, "-")
			word = strings.TrimSpace(word)
			if word == "" {
				continue
			}
			escaped := EscapeFullTextQuery(word)
			parts = append(parts, fmt.Sprintf("!! plainto_tsquery('%s', '%s')", language, escaped))
		} else if strings.Contains(term, "|") {
			// OR operator
			orTerms := strings.Split(term, "|")
			var orParts []string
			for _, ot := range orTerms {
				ot = strings.TrimSpace(ot)
				if ot == "" {
					continue
				}
				escaped := EscapeFullTextQuery(ot)
				orParts = append(orParts, fmt.Sprintf("plainto_tsquery('%s', '%s')", language, escaped))
			}
			if len(orParts) > 0 {
				parts = append(parts, "("+strings.Join(orParts, " || ")+")")
			}
		} else {
			escaped := EscapeFullTextQuery(term)
			parts = append(parts, fmt.Sprintf("plainto_tsquery('%s', '%s')", language, escaped))
		}
	}

	if len(parts) == 0 {
		return fmt.Sprintf("plainto_tsquery('%s', '%s')", language, EscapeFullTextQuery(input))
	}

	return strings.Join(parts, " && ")
}

// ParseTextSearchParam handles the _text parameter (searches narrative text).
func ParseTextSearchParam(value string) (*FullTextQuery, error) {
	return ParseFullTextQuery(value, "english")
}

// ParseContentSearchParam handles the _content parameter (searches all content).
func ParseContentSearchParam(value string) (*FullTextQuery, error) {
	return ParseFullTextQuery(value, "english")
}

// ---------------------------------------------------------------------------
// TSVector generation
// ---------------------------------------------------------------------------

// GenerateTSVector generates a PostgreSQL tsvector expression from config.
func GenerateTSVector(config *FullTextConfig) string {
	if len(config.TextColumns) == 0 {
		return ""
	}

	lang := config.Language
	if lang == "" {
		lang = "english"
	}

	var parts []string
	for _, col := range config.TextColumns {
		vec := fmt.Sprintf("to_tsvector('%s', COALESCE(%s, ''))", lang, col)
		if config.WeightMap != nil {
			if weight, ok := config.WeightMap[col]; ok {
				vec = fmt.Sprintf("setweight(to_tsvector('%s', COALESCE(%s, '')), '%s')", lang, col, weight)
			}
		}
		parts = append(parts, vec)
	}

	return strings.Join(parts, " || ")
}

// ---------------------------------------------------------------------------
// TSQuery generation
// ---------------------------------------------------------------------------

// GenerateTSQuery converts FHIR search text to a PostgreSQL tsquery expression string.
func GenerateTSQuery(text string, language string) string {
	if language == "" {
		language = "english"
	}
	escaped := EscapeFullTextQuery(text)
	return fmt.Sprintf("plainto_tsquery('%s', '%s')", language, escaped)
}

// ---------------------------------------------------------------------------
// SQL clause generation
// ---------------------------------------------------------------------------

// FullTextSearchClause generates a SQL WHERE clause for full-text search.
// Returns the clause and the arguments to bind.
func FullTextSearchClause(config *FullTextConfig, query *FullTextQuery, startIdx int) (string, []interface{}) {
	tsVec := GenerateTSVector(config)
	if tsVec == "" {
		return "1=0", nil
	}

	clause := fmt.Sprintf("(%s) @@ plainto_tsquery('%s', $%d)", tsVec, query.Language, startIdx)
	args := []interface{}{query.RawQuery}
	return clause, args
}

// FullTextRankClause generates a SQL ORDER BY expression for relevance ranking.
// Returns empty string if UseRanking is false.
func FullTextRankClause(config *FullTextConfig, query *FullTextQuery) string {
	if !query.UseRanking {
		return ""
	}

	tsVec := GenerateTSVector(config)
	if tsVec == "" {
		return ""
	}

	return fmt.Sprintf("ts_rank_cd(%s, %s) DESC", tsVec, query.TSQuery)
}

// ---------------------------------------------------------------------------
// Index DDL generation
// ---------------------------------------------------------------------------

// CreateFullTextIndex generates CREATE INDEX DDL for a resource type.
func CreateFullTextIndex(config *FullTextConfig) string {
	if len(config.TextColumns) == 0 {
		return ""
	}

	indexName := config.IndexName
	if indexName == "" {
		indexName = fmt.Sprintf("idx_%s_fulltext", strings.ToLower(config.ResourceType))
	}

	tableName := strings.ToLower(config.ResourceType) + "s"
	tsVec := GenerateTSVector(config)

	return fmt.Sprintf(
		"CREATE INDEX IF NOT EXISTS %s ON %s USING gin ((%s))",
		indexName, tableName, tsVec,
	)
}

// ---------------------------------------------------------------------------
// Highlight SQL generation
// ---------------------------------------------------------------------------

// BuildHighlightSQL generates ts_headline SQL for search result snippets.
func BuildHighlightSQL(config *FullTextConfig, query *FullTextQuery) string {
	if len(config.TextColumns) == 0 {
		return ""
	}

	lang := config.Language
	if lang == "" {
		lang = "english"
	}

	// Use the first text column for headline generation
	col := config.TextColumns[0]

	opts := query.HighlightOpts
	if opts == nil {
		opts = &HighlightOptions{
			MaxWords:     35,
			MinWords:     15,
			MaxFragments: 3,
			StartSel:     "<b>",
			StopSel:      "</b>",
		}
	}

	optStr := fmt.Sprintf(
		"MaxWords=%d, MinWords=%d, MaxFragments=%d, StartSel=%s, StopSel=%s",
		opts.MaxWords, opts.MinWords, opts.MaxFragments, opts.StartSel, opts.StopSel,
	)

	return fmt.Sprintf(
		"ts_headline('%s', COALESCE(%s, ''), %s, '%s')",
		lang, col, query.TSQuery, optStr,
	)
}

// ---------------------------------------------------------------------------
// Escaping and term splitting
// ---------------------------------------------------------------------------

// EscapeFullTextQuery safely escapes user input for tsquery.
// It escapes characters that have special meaning in PostgreSQL text search.
func EscapeFullTextQuery(input string) string {
	if input == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(input))

	for _, r := range input {
		switch r {
		case '\'':
			b.WriteString("''")
		case '\\':
			b.WriteString("\\\\")
		case '&', '|', '!', ':', '(', ')', '<', '>', '*':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

// SplitSearchTerms splits multi-word search input into terms.
// Quoted strings are preserved as a single term.
func SplitSearchTerms(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	var terms []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(input); i++ {
		ch := input[i]
		if ch == '"' {
			if inQuote {
				// End of quoted phrase
				term := current.String()
				if term != "" {
					terms = append(terms, term)
				}
				current.Reset()
				inQuote = false
			} else {
				// Start of quoted phrase; flush any current term
				term := strings.TrimSpace(current.String())
				if term != "" {
					terms = append(terms, term)
				}
				current.Reset()
				inQuote = true
			}
		} else if ch == ' ' && !inQuote {
			term := strings.TrimSpace(current.String())
			if term != "" {
				terms = append(terms, term)
			}
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}

	// Flush remaining
	term := strings.TrimSpace(current.String())
	if term != "" {
		terms = append(terms, term)
	}

	return terms
}

// ---------------------------------------------------------------------------
// Phrase and proximity queries
// ---------------------------------------------------------------------------

// BuildPhraseQuery creates a phrase search tsquery (words must appear together in order).
func BuildPhraseQuery(terms []string, language string) string {
	if len(terms) == 0 {
		return ""
	}
	if language == "" {
		language = "english"
	}
	phrase := strings.Join(terms, " ")
	escaped := EscapeFullTextQuery(phrase)
	return fmt.Sprintf("phraseto_tsquery('%s', '%s')", language, escaped)
}

// BuildProximityQuery creates a proximity-based tsquery where terms must appear
// within the specified distance of each other.
func BuildProximityQuery(terms []string, distance int, language string) string {
	if len(terms) == 0 {
		return ""
	}
	if language == "" {
		language = "english"
	}
	if len(terms) == 1 {
		escaped := EscapeFullTextQuery(terms[0])
		return fmt.Sprintf("plainto_tsquery('%s', '%s')", language, escaped)
	}

	// Build proximity using <N> operator: 'word1' <N> 'word2'
	var parts []string
	for _, term := range terms {
		escaped := EscapeFullTextQuery(term)
		parts = append(parts, fmt.Sprintf("'%s'", escaped))
	}
	return strings.Join(parts, fmt.Sprintf(" <%d> ", distance))
}

// ---------------------------------------------------------------------------
// Default configs for standard FHIR resource types
// ---------------------------------------------------------------------------

// DefaultFullTextConfigs returns configs for standard FHIR resource types.
func DefaultFullTextConfigs() map[string]*FullTextConfig {
	return map[string]*FullTextConfig{
		"Patient": {
			ResourceType: "Patient",
			TextColumns:  []string{"family_name", "given_name", "text_div"},
			WeightMap:    map[string]string{"family_name": "A", "given_name": "A", "text_div": "B"},
			Language:     "english",
			IndexName:    "idx_patient_fulltext",
		},
		"Observation": {
			ResourceType: "Observation",
			TextColumns:  []string{"code_display", "value_string", "text_div", "note"},
			WeightMap:    map[string]string{"code_display": "A", "value_string": "B", "text_div": "C", "note": "C"},
			Language:     "english",
			IndexName:    "idx_observation_fulltext",
		},
		"Condition": {
			ResourceType: "Condition",
			TextColumns:  []string{"code_display", "text_div", "note"},
			WeightMap:    map[string]string{"code_display": "A", "text_div": "B", "note": "C"},
			Language:     "english",
			IndexName:    "idx_condition_fulltext",
		},
		"MedicationRequest": {
			ResourceType: "MedicationRequest",
			TextColumns:  []string{"medication_display", "text_div", "note"},
			WeightMap:    map[string]string{"medication_display": "A", "text_div": "B", "note": "C"},
			Language:     "english",
			IndexName:    "idx_medicationrequest_fulltext",
		},
		"DiagnosticReport": {
			ResourceType: "DiagnosticReport",
			TextColumns:  []string{"conclusion", "text_div"},
			WeightMap:    map[string]string{"conclusion": "A", "text_div": "B"},
			Language:     "english",
			IndexName:    "idx_diagnosticreport_fulltext",
		},
		"AllergyIntolerance": {
			ResourceType: "AllergyIntolerance",
			TextColumns:  []string{"code_display", "text_div", "note"},
			WeightMap:    map[string]string{"code_display": "A", "text_div": "B", "note": "C"},
			Language:     "english",
			IndexName:    "idx_allergyintolerance_fulltext",
		},
		"Procedure": {
			ResourceType: "Procedure",
			TextColumns:  []string{"code_display", "text_div", "note"},
			WeightMap:    map[string]string{"code_display": "A", "text_div": "B", "note": "C"},
			Language:     "english",
			IndexName:    "idx_procedure_fulltext",
		},
		"Encounter": {
			ResourceType: "Encounter",
			TextColumns:  []string{"type_display", "text_div", "reason_display"},
			WeightMap:    map[string]string{"type_display": "A", "text_div": "B", "reason_display": "B"},
			Language:     "english",
			IndexName:    "idx_encounter_fulltext",
		},
	}
}

// ---------------------------------------------------------------------------
// Integration with SearchQuery
// ---------------------------------------------------------------------------

// ApplyFullTextSearch adds full-text search to an existing SearchQuery.
// paramName must be "_text" or "_content".
func (e *FullTextSearchEngine) ApplyFullTextSearch(query *SearchQuery, paramName string, paramValue string) error {
	if paramName != "_text" && paramName != "_content" {
		return fmt.Errorf("unsupported full-text search parameter: %s", paramName)
	}

	var ftQuery *FullTextQuery
	var err error

	switch paramName {
	case "_text":
		ftQuery, err = ParseTextSearchParam(paramValue)
	case "_content":
		ftQuery, err = ParseContentSearchParam(paramValue)
	}
	if err != nil {
		return fmt.Errorf("invalid full-text query: %w", err)
	}

	// Find the appropriate config. For ApplyFullTextSearch, we attempt to
	// determine the resource type from the table name. If no specific config
	// is found, we build a generic one using text_div.
	var config *FullTextConfig
	for _, cfg := range e.Configs {
		tableName := strings.ToLower(cfg.ResourceType) + "s"
		if tableName == query.table {
			config = cfg
			break
		}
	}

	if config == nil {
		// Fallback: use a generic config with text_div
		config = &FullTextConfig{
			TextColumns: []string{"text_div"},
			Language:    "english",
		}
	}

	// For _text, restrict to narrative columns (text_div)
	if paramName == "_text" {
		config = &FullTextConfig{
			TextColumns: []string{"text_div"},
			WeightMap:   config.WeightMap,
			Language:    config.Language,
		}
	}

	clause, args := FullTextSearchClause(config, ftQuery, query.idx)
	query.where += " AND " + clause
	query.args = append(query.args, args...)
	query.idx += len(args)

	// If ranking is requested, set the order by
	if ftQuery.UseRanking {
		rankClause := FullTextRankClause(config, ftQuery)
		if rankClause != "" {
			query.orderBy = rankClause
		}
	}

	return nil
}
