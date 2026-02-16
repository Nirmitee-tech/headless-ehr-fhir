package fhir

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// EscapeFullTextQuery tests
// ---------------------------------------------------------------------------

func TestEscapeFullTextQuery_SimpleWord(t *testing.T) {
	got := EscapeFullTextQuery("hello")
	if got != "hello" {
		t.Errorf("EscapeFullTextQuery(%q) = %q, want %q", "hello", got, "hello")
	}
}

func TestEscapeFullTextQuery_SpecialChars(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello'world", "hello''world"},
		{"test\\value", "test\\\\value"},
		{"a:b", "a\\:b"},
		{"a&b", "a\\&b"},
		{"a|b", "a\\|b"},
		{"a!b", "a\\!b"},
		{"(test)", "\\(test\\)"},
		{"<script>", "\\<script\\>"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := EscapeFullTextQuery(tt.input)
			if got != tt.want {
				t.Errorf("EscapeFullTextQuery(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeFullTextQuery_Empty(t *testing.T) {
	got := EscapeFullTextQuery("")
	if got != "" {
		t.Errorf("EscapeFullTextQuery(%q) = %q, want empty", "", got)
	}
}

func TestEscapeFullTextQuery_SQLInjection(t *testing.T) {
	input := "'; DROP TABLE patients;--"
	got := EscapeFullTextQuery(input)
	// Single quotes must be doubled ('' is the PostgreSQL escape for a literal quote)
	if !strings.HasPrefix(got, "''") {
		t.Errorf("EscapeFullTextQuery should double single quotes in %q, got %q", input, got)
	}
}

func TestEscapeFullTextQuery_Unicode(t *testing.T) {
	got := EscapeFullTextQuery("fieber")
	if got != "fieber" {
		t.Errorf("EscapeFullTextQuery(%q) = %q, want %q", "fieber", got, "fieber")
	}
}

// ---------------------------------------------------------------------------
// SplitSearchTerms tests
// ---------------------------------------------------------------------------

func TestSplitSearchTerms_SimpleWords(t *testing.T) {
	got := SplitSearchTerms("hello world")
	if len(got) != 2 {
		t.Fatalf("SplitSearchTerms got %d terms, want 2", len(got))
	}
	if got[0] != "hello" || got[1] != "world" {
		t.Errorf("SplitSearchTerms = %v, want [hello world]", got)
	}
}

func TestSplitSearchTerms_QuotedPhrase(t *testing.T) {
	got := SplitSearchTerms(`"exact phrase" other`)
	if len(got) != 2 {
		t.Fatalf("SplitSearchTerms got %d terms, want 2", len(got))
	}
	if got[0] != "exact phrase" {
		t.Errorf("first term = %q, want %q", got[0], "exact phrase")
	}
	if got[1] != "other" {
		t.Errorf("second term = %q, want %q", got[1], "other")
	}
}

func TestSplitSearchTerms_MultipleQuotedPhrases(t *testing.T) {
	got := SplitSearchTerms(`"hello world" "foo bar"`)
	if len(got) != 2 {
		t.Fatalf("SplitSearchTerms got %d terms, want 2", len(got))
	}
	if got[0] != "hello world" {
		t.Errorf("first term = %q, want %q", got[0], "hello world")
	}
	if got[1] != "foo bar" {
		t.Errorf("second term = %q, want %q", got[1], "foo bar")
	}
}

func TestSplitSearchTerms_Empty(t *testing.T) {
	got := SplitSearchTerms("")
	if len(got) != 0 {
		t.Errorf("SplitSearchTerms(%q) = %v, want empty", "", got)
	}
}

func TestSplitSearchTerms_OnlySpaces(t *testing.T) {
	got := SplitSearchTerms("   ")
	if len(got) != 0 {
		t.Errorf("SplitSearchTerms(%q) = %v, want empty", "   ", got)
	}
}

func TestSplitSearchTerms_MixedQuotedAndUnquoted(t *testing.T) {
	got := SplitSearchTerms(`diabetes "type 2" mellitus`)
	if len(got) != 3 {
		t.Fatalf("SplitSearchTerms got %d terms, want 3", len(got))
	}
	if got[0] != "diabetes" {
		t.Errorf("term[0] = %q, want %q", got[0], "diabetes")
	}
	if got[1] != "type 2" {
		t.Errorf("term[1] = %q, want %q", got[1], "type 2")
	}
	if got[2] != "mellitus" {
		t.Errorf("term[2] = %q, want %q", got[2], "mellitus")
	}
}

func TestSplitSearchTerms_SingleWord(t *testing.T) {
	got := SplitSearchTerms("diabetes")
	if len(got) != 1 || got[0] != "diabetes" {
		t.Errorf("SplitSearchTerms = %v, want [diabetes]", got)
	}
}

// ---------------------------------------------------------------------------
// GenerateTSQuery tests
// ---------------------------------------------------------------------------

func TestGenerateTSQuery_SimpleWord(t *testing.T) {
	got := GenerateTSQuery("diabetes", "english")
	if got != "plainto_tsquery('english', 'diabetes')" {
		t.Errorf("GenerateTSQuery = %q, want plainto_tsquery('english', 'diabetes')", got)
	}
}

func TestGenerateTSQuery_MultipleWords(t *testing.T) {
	got := GenerateTSQuery("diabetes mellitus", "english")
	// Multi-word should produce a plainto_tsquery with AND semantics
	if got != "plainto_tsquery('english', 'diabetes mellitus')" {
		t.Errorf("GenerateTSQuery = %q, want plainto_tsquery with both words", got)
	}
}

func TestGenerateTSQuery_EmptyString(t *testing.T) {
	got := GenerateTSQuery("", "english")
	if got != "plainto_tsquery('english', '')" {
		t.Errorf("GenerateTSQuery(%q) = %q", "", got)
	}
}

func TestGenerateTSQuery_DefaultLanguage(t *testing.T) {
	got := GenerateTSQuery("test", "")
	if !strings.Contains(got, "'english'") {
		t.Errorf("GenerateTSQuery with empty language should default to english, got %q", got)
	}
}

func TestGenerateTSQuery_EscapesInput(t *testing.T) {
	got := GenerateTSQuery("test'injection", "english")
	if strings.Contains(got, "test'injection") {
		t.Errorf("GenerateTSQuery should escape single quotes, got %q", got)
	}
}

func TestGenerateTSQuery_SQLInjectionAttempt(t *testing.T) {
	got := GenerateTSQuery("'; DROP TABLE patients;--", "english")
	// Single quotes must be escaped (doubled) so the injection is neutralized
	// The escaped output should contain '' (doubled quote) not a bare single quote
	if !strings.Contains(got, "''") {
		t.Errorf("GenerateTSQuery should escape single quotes for SQL injection, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// ParseFullTextQuery tests
// ---------------------------------------------------------------------------

func TestParseFullTextQuery_SimpleWord(t *testing.T) {
	q, err := ParseFullTextQuery("diabetes", "english")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error: %v", err)
	}
	if q.RawQuery != "diabetes" {
		t.Errorf("RawQuery = %q, want %q", q.RawQuery, "diabetes")
	}
	if q.Language != "english" {
		t.Errorf("Language = %q, want %q", q.Language, "english")
	}
	if q.TSQuery == "" {
		t.Error("TSQuery should not be empty")
	}
}

func TestParseFullTextQuery_MultiWord(t *testing.T) {
	q, err := ParseFullTextQuery("diabetes mellitus", "english")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error: %v", err)
	}
	if q.RawQuery != "diabetes mellitus" {
		t.Errorf("RawQuery = %q, want %q", q.RawQuery, "diabetes mellitus")
	}
	if q.TSQuery == "" {
		t.Error("TSQuery should not be empty")
	}
}

func TestParseFullTextQuery_PhraseSearch(t *testing.T) {
	q, err := ParseFullTextQuery(`"type 2 diabetes"`, "english")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error: %v", err)
	}
	if !strings.Contains(q.TSQuery, "phraseto_tsquery") {
		t.Errorf("phrase search should use phraseto_tsquery, got %q", q.TSQuery)
	}
}

func TestParseFullTextQuery_PrefixMatch(t *testing.T) {
	q, err := ParseFullTextQuery("diab*", "english")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error: %v", err)
	}
	if !strings.Contains(q.TSQuery, ":*") {
		t.Errorf("prefix match should contain :*, got %q", q.TSQuery)
	}
}

func TestParseFullTextQuery_ANDOperator(t *testing.T) {
	q, err := ParseFullTextQuery("+diabetes +mellitus", "english")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error: %v", err)
	}
	if !strings.Contains(q.TSQuery, "&") {
		t.Errorf("AND operator should produce &, got %q", q.TSQuery)
	}
}

func TestParseFullTextQuery_OROperator(t *testing.T) {
	q, err := ParseFullTextQuery("diabetes|hypertension", "english")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error: %v", err)
	}
	if !strings.Contains(q.TSQuery, "|") {
		t.Errorf("OR operator should produce |, got %q", q.TSQuery)
	}
}

func TestParseFullTextQuery_NOTOperator(t *testing.T) {
	q, err := ParseFullTextQuery("diabetes -juvenile", "english")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error: %v", err)
	}
	if !strings.Contains(q.TSQuery, "!") {
		t.Errorf("NOT operator should produce !, got %q", q.TSQuery)
	}
}

func TestParseFullTextQuery_Empty(t *testing.T) {
	_, err := ParseFullTextQuery("", "english")
	if err == nil {
		t.Error("ParseFullTextQuery should return error for empty query")
	}
}

func TestParseFullTextQuery_OnlySpaces(t *testing.T) {
	_, err := ParseFullTextQuery("   ", "english")
	if err == nil {
		t.Error("ParseFullTextQuery should return error for whitespace-only query")
	}
}

func TestParseFullTextQuery_SpecialCharsOnly(t *testing.T) {
	_, err := ParseFullTextQuery("!@#$%^", "english")
	if err == nil {
		t.Error("ParseFullTextQuery should return error for special-chars-only query")
	}
}

func TestParseFullTextQuery_DefaultLanguage(t *testing.T) {
	q, err := ParseFullTextQuery("test", "")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error: %v", err)
	}
	if q.Language != "english" {
		t.Errorf("Language = %q, want %q (default)", q.Language, "english")
	}
}

func TestParseFullTextQuery_UseRanking(t *testing.T) {
	q, err := ParseFullTextQuery("diabetes", "english")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error: %v", err)
	}
	if !q.UseRanking {
		t.Error("UseRanking should default to true")
	}
}

func TestParseFullTextQuery_DefaultHighlightOpts(t *testing.T) {
	q, err := ParseFullTextQuery("diabetes", "english")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error: %v", err)
	}
	if q.HighlightOpts == nil {
		t.Fatal("HighlightOpts should not be nil")
	}
	if q.HighlightOpts.MaxWords <= 0 {
		t.Errorf("MaxWords = %d, want positive", q.HighlightOpts.MaxWords)
	}
	if q.HighlightOpts.StartSel == "" {
		t.Error("StartSel should not be empty")
	}
	if q.HighlightOpts.StopSel == "" {
		t.Error("StopSel should not be empty")
	}
}

// ---------------------------------------------------------------------------
// GenerateTSVector tests
// ---------------------------------------------------------------------------

func TestGenerateTSVector_SingleColumn(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"text_div"},
		Language:    "english",
	}
	got := GenerateTSVector(config)
	if !strings.Contains(got, "to_tsvector") {
		t.Errorf("GenerateTSVector should contain to_tsvector, got %q", got)
	}
	if !strings.Contains(got, "text_div") {
		t.Errorf("GenerateTSVector should contain text_div, got %q", got)
	}
	if !strings.Contains(got, "'english'") {
		t.Errorf("GenerateTSVector should contain language, got %q", got)
	}
}

func TestGenerateTSVector_MultipleColumns(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"family_name", "given_name", "text_div"},
		Language:    "english",
	}
	got := GenerateTSVector(config)
	if !strings.Contains(got, "family_name") {
		t.Errorf("GenerateTSVector should contain family_name, got %q", got)
	}
	if !strings.Contains(got, "given_name") {
		t.Errorf("GenerateTSVector should contain given_name, got %q", got)
	}
	if !strings.Contains(got, "text_div") {
		t.Errorf("GenerateTSVector should contain text_div, got %q", got)
	}
	// Multiple columns should be concatenated with ||
	if !strings.Contains(got, "||") {
		t.Errorf("GenerateTSVector should concatenate columns with ||, got %q", got)
	}
}

func TestGenerateTSVector_WeightedColumns(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"family_name", "text_div"},
		WeightMap:   map[string]string{"family_name": "A", "text_div": "B"},
		Language:    "english",
	}
	got := GenerateTSVector(config)
	if !strings.Contains(got, "setweight") {
		t.Errorf("GenerateTSVector with weights should contain setweight, got %q", got)
	}
	if !strings.Contains(got, "'A'") {
		t.Errorf("GenerateTSVector should contain weight A, got %q", got)
	}
	if !strings.Contains(got, "'B'") {
		t.Errorf("GenerateTSVector should contain weight B, got %q", got)
	}
}

func TestGenerateTSVector_DefaultLanguage(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"text_div"},
	}
	got := GenerateTSVector(config)
	if !strings.Contains(got, "'english'") {
		t.Errorf("GenerateTSVector should default to english, got %q", got)
	}
}

func TestGenerateTSVector_EmptyColumns(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{},
		Language:    "english",
	}
	got := GenerateTSVector(config)
	if got != "" {
		t.Errorf("GenerateTSVector with no columns should return empty, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// FullTextSearchClause tests
// ---------------------------------------------------------------------------

func TestFullTextSearchClause_SingleColumn(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"text_div"},
		Language:    "english",
	}
	query := &FullTextQuery{
		RawQuery: "diabetes",
		TSQuery:  "plainto_tsquery('english', 'diabetes')",
		Language: "english",
	}
	clause, args := FullTextSearchClause(config, query, 1)
	if clause == "" {
		t.Fatal("clause should not be empty")
	}
	if !strings.Contains(clause, "@@") {
		t.Errorf("clause should contain @@ operator, got %q", clause)
	}
	if !strings.Contains(clause, "to_tsvector") {
		t.Errorf("clause should contain to_tsvector, got %q", clause)
	}
	_ = args
}

func TestFullTextSearchClause_MultipleColumns(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"family_name", "given_name", "text_div"},
		Language:    "english",
	}
	query := &FullTextQuery{
		RawQuery: "smith",
		TSQuery:  "plainto_tsquery('english', 'smith')",
		Language: "english",
	}
	clause, _ := FullTextSearchClause(config, query, 1)
	if clause == "" {
		t.Fatal("clause should not be empty")
	}
	if !strings.Contains(clause, "@@") {
		t.Errorf("clause should contain @@ operator, got %q", clause)
	}
}

func TestFullTextSearchClause_WeightedColumns(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"family_name", "text_div"},
		WeightMap:   map[string]string{"family_name": "A", "text_div": "B"},
		Language:    "english",
	}
	query := &FullTextQuery{
		RawQuery: "smith",
		TSQuery:  "plainto_tsquery('english', 'smith')",
		Language: "english",
	}
	clause, _ := FullTextSearchClause(config, query, 1)
	if !strings.Contains(clause, "setweight") {
		t.Errorf("weighted clause should contain setweight, got %q", clause)
	}
}

func TestFullTextSearchClause_ParameterIndex(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"text_div"},
		Language:    "english",
	}
	query := &FullTextQuery{
		RawQuery: "diabetes",
		TSQuery:  "plainto_tsquery('english', 'diabetes')",
		Language: "english",
	}
	clause, args := FullTextSearchClause(config, query, 5)
	if !strings.Contains(clause, "$5") {
		t.Errorf("clause should use parameter index 5, got %q", clause)
	}
	if len(args) < 1 {
		t.Error("args should contain at least the search text")
	}
}

func TestFullTextSearchClause_WithRanking(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"text_div"},
		Language:    "english",
	}
	query := &FullTextQuery{
		RawQuery:   "diabetes",
		TSQuery:    "plainto_tsquery('english', 'diabetes')",
		Language:   "english",
		UseRanking: true,
		MinRank:    0.1,
	}
	clause, _ := FullTextSearchClause(config, query, 1)
	if clause == "" {
		t.Fatal("clause should not be empty")
	}
}

// ---------------------------------------------------------------------------
// FullTextRankClause tests
// ---------------------------------------------------------------------------

func TestFullTextRankClause_Basic(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"text_div"},
		Language:    "english",
	}
	query := &FullTextQuery{
		RawQuery:   "diabetes",
		TSQuery:    "plainto_tsquery('english', 'diabetes')",
		Language:   "english",
		UseRanking: true,
	}
	got := FullTextRankClause(config, query)
	if got == "" {
		t.Fatal("rank clause should not be empty")
	}
	if !strings.Contains(got, "ts_rank_cd") {
		t.Errorf("rank clause should contain ts_rank_cd, got %q", got)
	}
	if !strings.Contains(got, "DESC") {
		t.Errorf("rank clause should contain DESC for ordering, got %q", got)
	}
}

func TestFullTextRankClause_NoRanking(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"text_div"},
		Language:    "english",
	}
	query := &FullTextQuery{
		RawQuery:   "diabetes",
		TSQuery:    "plainto_tsquery('english', 'diabetes')",
		Language:   "english",
		UseRanking: false,
	}
	got := FullTextRankClause(config, query)
	if got != "" {
		t.Errorf("rank clause should be empty when UseRanking=false, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// CreateFullTextIndex tests
// ---------------------------------------------------------------------------

func TestCreateFullTextIndex_Basic(t *testing.T) {
	config := &FullTextConfig{
		ResourceType: "Patient",
		TextColumns:  []string{"text_div"},
		Language:     "english",
		IndexName:    "idx_patient_fulltext",
	}
	got := CreateFullTextIndex(config)
	if !strings.Contains(got, "CREATE INDEX") {
		t.Errorf("should contain CREATE INDEX, got %q", got)
	}
	if !strings.Contains(got, "idx_patient_fulltext") {
		t.Errorf("should contain index name, got %q", got)
	}
	if !strings.Contains(got, "GIN") || !strings.Contains(got, "gin") {
		// Case-insensitive check
		if !strings.Contains(strings.ToLower(got), "gin") {
			t.Errorf("should specify GIN index type, got %q", got)
		}
	}
	if !strings.Contains(got, "to_tsvector") {
		t.Errorf("should contain to_tsvector, got %q", got)
	}
}

func TestCreateFullTextIndex_MultipleColumns(t *testing.T) {
	config := &FullTextConfig{
		ResourceType: "Observation",
		TextColumns:  []string{"code_display", "value_string", "text_div"},
		Language:     "english",
		IndexName:    "idx_observation_fulltext",
	}
	got := CreateFullTextIndex(config)
	if !strings.Contains(got, "idx_observation_fulltext") {
		t.Errorf("should contain index name, got %q", got)
	}
	if !strings.Contains(got, "code_display") {
		t.Errorf("should contain code_display column, got %q", got)
	}
}

func TestCreateFullTextIndex_WeightedColumns(t *testing.T) {
	config := &FullTextConfig{
		ResourceType: "Patient",
		TextColumns:  []string{"family_name", "text_div"},
		WeightMap:    map[string]string{"family_name": "A", "text_div": "B"},
		Language:     "english",
		IndexName:    "idx_patient_fulltext",
	}
	got := CreateFullTextIndex(config)
	if !strings.Contains(got, "CREATE INDEX") {
		t.Errorf("should contain CREATE INDEX, got %q", got)
	}
}

func TestCreateFullTextIndex_DefaultIndexName(t *testing.T) {
	config := &FullTextConfig{
		ResourceType: "Condition",
		TextColumns:  []string{"text_div"},
		Language:     "english",
	}
	got := CreateFullTextIndex(config)
	if !strings.Contains(strings.ToLower(got), "condition") {
		t.Errorf("default index name should include resource type, got %q", got)
	}
}

func TestCreateFullTextIndex_IfNotExists(t *testing.T) {
	config := &FullTextConfig{
		ResourceType: "Patient",
		TextColumns:  []string{"text_div"},
		Language:     "english",
		IndexName:    "idx_patient_fulltext",
	}
	got := CreateFullTextIndex(config)
	if !strings.Contains(got, "IF NOT EXISTS") {
		t.Errorf("should contain IF NOT EXISTS, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// BuildPhraseQuery tests
// ---------------------------------------------------------------------------

func TestBuildPhraseQuery_Basic(t *testing.T) {
	got := BuildPhraseQuery([]string{"type", "2", "diabetes"}, "english")
	if !strings.Contains(got, "phraseto_tsquery") {
		t.Errorf("phrase query should use phraseto_tsquery, got %q", got)
	}
	if !strings.Contains(got, "type 2 diabetes") {
		t.Errorf("phrase query should contain combined phrase, got %q", got)
	}
}

func TestBuildPhraseQuery_SingleTerm(t *testing.T) {
	got := BuildPhraseQuery([]string{"diabetes"}, "english")
	if got == "" {
		t.Error("phrase query should not be empty for single term")
	}
}

func TestBuildPhraseQuery_EmptyTerms(t *testing.T) {
	got := BuildPhraseQuery([]string{}, "english")
	if got != "" {
		t.Errorf("phrase query should be empty for no terms, got %q", got)
	}
}

func TestBuildPhraseQuery_DefaultLanguage(t *testing.T) {
	got := BuildPhraseQuery([]string{"hello", "world"}, "")
	if !strings.Contains(got, "'english'") {
		t.Errorf("phrase query should default to english, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// BuildProximityQuery tests
// ---------------------------------------------------------------------------

func TestBuildProximityQuery_Basic(t *testing.T) {
	got := BuildProximityQuery([]string{"diabetes", "mellitus"}, 5, "english")
	if got == "" {
		t.Error("proximity query should not be empty")
	}
	// Proximity uses <N> operator in tsquery
	if !strings.Contains(got, "<") {
		t.Errorf("proximity query should contain distance operator, got %q", got)
	}
}

func TestBuildProximityQuery_EmptyTerms(t *testing.T) {
	got := BuildProximityQuery([]string{}, 5, "english")
	if got != "" {
		t.Errorf("proximity query should be empty for no terms, got %q", got)
	}
}

func TestBuildProximityQuery_SingleTerm(t *testing.T) {
	got := BuildProximityQuery([]string{"diabetes"}, 5, "english")
	// Single term proximity doesn't need distance; should still produce valid output
	if got == "" {
		t.Error("proximity query should not be empty for single term")
	}
}

func TestBuildProximityQuery_DefaultLanguage(t *testing.T) {
	// Single-term proximity uses plainto_tsquery with language
	got := BuildProximityQuery([]string{"test"}, 3, "")
	if !strings.Contains(got, "'english'") {
		t.Errorf("single-term proximity query should default to english, got %q", got)
	}
}

func TestBuildProximityQuery_MultiTermDistance(t *testing.T) {
	got := BuildProximityQuery([]string{"a", "b"}, 3, "english")
	if !strings.Contains(got, "<3>") {
		t.Errorf("proximity query should contain distance operator <3>, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// DefaultFullTextConfigs tests
// ---------------------------------------------------------------------------

func TestDefaultFullTextConfigs_NotEmpty(t *testing.T) {
	configs := DefaultFullTextConfigs()
	if len(configs) == 0 {
		t.Fatal("DefaultFullTextConfigs() returned empty map")
	}
}

func TestDefaultFullTextConfigs_ExpectedResourceTypes(t *testing.T) {
	configs := DefaultFullTextConfigs()
	expected := []string{
		"Patient",
		"Observation",
		"Condition",
		"MedicationRequest",
		"DiagnosticReport",
		"AllergyIntolerance",
		"Procedure",
		"Encounter",
	}
	for _, rt := range expected {
		if _, ok := configs[rt]; !ok {
			t.Errorf("DefaultFullTextConfigs missing resource type %q", rt)
		}
	}
}

func TestDefaultFullTextConfigs_PatientConfig(t *testing.T) {
	configs := DefaultFullTextConfigs()
	cfg, ok := configs["Patient"]
	if !ok {
		t.Fatal("missing Patient config")
	}
	if cfg.ResourceType != "Patient" {
		t.Errorf("ResourceType = %q, want %q", cfg.ResourceType, "Patient")
	}
	if len(cfg.TextColumns) < 2 {
		t.Errorf("Patient should have at least 2 text columns, got %d", len(cfg.TextColumns))
	}
	// Verify weight map contains expected columns
	if cfg.WeightMap == nil {
		t.Fatal("Patient WeightMap should not be nil")
	}
	if w, ok := cfg.WeightMap["family_name"]; !ok || w != "A" {
		t.Errorf("family_name weight = %q, want %q", w, "A")
	}
	if w, ok := cfg.WeightMap["text_div"]; !ok || w != "B" {
		t.Errorf("text_div weight = %q, want %q", w, "B")
	}
}

func TestDefaultFullTextConfigs_ObservationConfig(t *testing.T) {
	configs := DefaultFullTextConfigs()
	cfg, ok := configs["Observation"]
	if !ok {
		t.Fatal("missing Observation config")
	}
	if cfg.ResourceType != "Observation" {
		t.Errorf("ResourceType = %q, want %q", cfg.ResourceType, "Observation")
	}
	if len(cfg.TextColumns) < 3 {
		t.Errorf("Observation should have at least 3 text columns, got %d", len(cfg.TextColumns))
	}
	if w, ok := cfg.WeightMap["code_display"]; !ok || w != "A" {
		t.Errorf("code_display weight = %q, want %q", w, "A")
	}
}

func TestDefaultFullTextConfigs_ConditionConfig(t *testing.T) {
	configs := DefaultFullTextConfigs()
	cfg, ok := configs["Condition"]
	if !ok {
		t.Fatal("missing Condition config")
	}
	if w, ok := cfg.WeightMap["code_display"]; !ok || w != "A" {
		t.Errorf("code_display weight = %q, want %q", w, "A")
	}
}

func TestDefaultFullTextConfigs_AllHaveLanguage(t *testing.T) {
	configs := DefaultFullTextConfigs()
	for rt, cfg := range configs {
		if cfg.Language == "" {
			t.Errorf("config for %q has empty Language", rt)
		}
	}
}

func TestDefaultFullTextConfigs_AllHaveTextColumns(t *testing.T) {
	configs := DefaultFullTextConfigs()
	for rt, cfg := range configs {
		if len(cfg.TextColumns) == 0 {
			t.Errorf("config for %q has no TextColumns", rt)
		}
	}
}

func TestDefaultFullTextConfigs_AllHaveWeightMap(t *testing.T) {
	configs := DefaultFullTextConfigs()
	for rt, cfg := range configs {
		if cfg.WeightMap == nil || len(cfg.WeightMap) == 0 {
			t.Errorf("config for %q has empty WeightMap", rt)
		}
	}
}

func TestDefaultFullTextConfigs_EncounterConfig(t *testing.T) {
	configs := DefaultFullTextConfigs()
	cfg, ok := configs["Encounter"]
	if !ok {
		t.Fatal("missing Encounter config")
	}
	if w, ok := cfg.WeightMap["type_display"]; !ok || w != "A" {
		t.Errorf("type_display weight = %q, want %q", w, "A")
	}
	if w, ok := cfg.WeightMap["reason_display"]; !ok || w != "B" {
		t.Errorf("reason_display weight = %q, want %q", w, "B")
	}
}

func TestDefaultFullTextConfigs_DiagnosticReportConfig(t *testing.T) {
	configs := DefaultFullTextConfigs()
	cfg, ok := configs["DiagnosticReport"]
	if !ok {
		t.Fatal("missing DiagnosticReport config")
	}
	if w, ok := cfg.WeightMap["conclusion"]; !ok || w != "A" {
		t.Errorf("conclusion weight = %q, want %q", w, "A")
	}
}

// ---------------------------------------------------------------------------
// FullTextSearchEngine tests
// ---------------------------------------------------------------------------

func TestNewFullTextSearchEngine(t *testing.T) {
	engine := NewFullTextSearchEngine()
	if engine == nil {
		t.Fatal("NewFullTextSearchEngine should not return nil")
	}
	if engine.Configs == nil {
		t.Fatal("Configs map should not be nil")
	}
}

func TestFullTextSearchEngine_RegisterConfig(t *testing.T) {
	engine := NewFullTextSearchEngine()
	config := &FullTextConfig{
		ResourceType: "CustomResource",
		TextColumns:  []string{"text_div"},
		Language:     "english",
	}
	engine.RegisterConfig(config)
	if _, ok := engine.Configs["CustomResource"]; !ok {
		t.Error("RegisterConfig should add config to engine")
	}
}

func TestFullTextSearchEngine_RegisterConfig_Overwrites(t *testing.T) {
	engine := NewFullTextSearchEngine()
	config1 := &FullTextConfig{
		ResourceType: "Patient",
		TextColumns:  []string{"text_div"},
		Language:     "english",
	}
	config2 := &FullTextConfig{
		ResourceType: "Patient",
		TextColumns:  []string{"text_div", "given_name"},
		Language:     "english",
	}
	engine.RegisterConfig(config1)
	engine.RegisterConfig(config2)
	if len(engine.Configs["Patient"].TextColumns) != 2 {
		t.Error("RegisterConfig should overwrite existing config")
	}
}

// ---------------------------------------------------------------------------
// ApplyFullTextSearch tests
// ---------------------------------------------------------------------------

func TestApplyFullTextSearch_TextParam(t *testing.T) {
	engine := NewFullTextSearchEngine()
	engine.RegisterConfig(&FullTextConfig{
		ResourceType: "Patient",
		TextColumns:  []string{"text_div"},
		Language:     "english",
	})

	q := NewSearchQuery("patients", "id, resource")
	err := engine.ApplyFullTextSearch(q, "_text", "diabetes")
	if err != nil {
		t.Fatalf("ApplyFullTextSearch error: %v", err)
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "@@") {
		t.Errorf("SQL should contain @@ for full-text search, got: %s", sql)
	}
}

func TestApplyFullTextSearch_ContentParam(t *testing.T) {
	engine := NewFullTextSearchEngine()
	engine.RegisterConfig(&FullTextConfig{
		ResourceType: "Observation",
		TextColumns:  []string{"code_display", "value_string", "text_div"},
		Language:     "english",
	})

	q := NewSearchQuery("observations", "id, resource")
	err := engine.ApplyFullTextSearch(q, "_content", "blood pressure")
	if err != nil {
		t.Fatalf("ApplyFullTextSearch error: %v", err)
	}
	sql := q.CountSQL()
	if !strings.Contains(sql, "@@") {
		t.Errorf("SQL should contain @@ for full-text search, got: %s", sql)
	}
}

func TestApplyFullTextSearch_InvalidParam(t *testing.T) {
	engine := NewFullTextSearchEngine()
	q := NewSearchQuery("patients", "id")
	err := engine.ApplyFullTextSearch(q, "name", "diabetes")
	if err == nil {
		t.Error("ApplyFullTextSearch should return error for non-fulltext param")
	}
}

func TestApplyFullTextSearch_EmptyValue(t *testing.T) {
	engine := NewFullTextSearchEngine()
	engine.RegisterConfig(&FullTextConfig{
		ResourceType: "Patient",
		TextColumns:  []string{"text_div"},
		Language:     "english",
	})

	q := NewSearchQuery("patients", "id")
	err := engine.ApplyFullTextSearch(q, "_text", "")
	if err == nil {
		t.Error("ApplyFullTextSearch should return error for empty value")
	}
}

func TestApplyFullTextSearch_IdxAdvancement(t *testing.T) {
	engine := NewFullTextSearchEngine()
	engine.RegisterConfig(&FullTextConfig{
		ResourceType: "Patient",
		TextColumns:  []string{"text_div"},
		Language:     "english",
	})

	q := NewSearchQuery("patients", "id, resource")
	// Add a normal clause first
	q.Add("status = $1", "active")
	startIdx := q.Idx()

	err := engine.ApplyFullTextSearch(q, "_text", "diabetes")
	if err != nil {
		t.Fatalf("ApplyFullTextSearch error: %v", err)
	}
	if q.Idx() <= startIdx {
		t.Errorf("idx should advance after ApplyFullTextSearch, was %d, now %d", startIdx, q.Idx())
	}
}

// ---------------------------------------------------------------------------
// ParseTextSearchParam / ParseContentSearchParam tests
// ---------------------------------------------------------------------------

func TestParseTextSearchParam_Basic(t *testing.T) {
	q, err := ParseTextSearchParam("diabetes mellitus")
	if err != nil {
		t.Fatalf("ParseTextSearchParam error: %v", err)
	}
	if q.RawQuery != "diabetes mellitus" {
		t.Errorf("RawQuery = %q, want %q", q.RawQuery, "diabetes mellitus")
	}
}

func TestParseTextSearchParam_Empty(t *testing.T) {
	_, err := ParseTextSearchParam("")
	if err == nil {
		t.Error("ParseTextSearchParam should return error for empty")
	}
}

func TestParseContentSearchParam_Basic(t *testing.T) {
	q, err := ParseContentSearchParam("blood pressure")
	if err != nil {
		t.Fatalf("ParseContentSearchParam error: %v", err)
	}
	if q.RawQuery != "blood pressure" {
		t.Errorf("RawQuery = %q, want %q", q.RawQuery, "blood pressure")
	}
}

func TestParseContentSearchParam_Empty(t *testing.T) {
	_, err := ParseContentSearchParam("")
	if err == nil {
		t.Error("ParseContentSearchParam should return error for empty")
	}
}

func TestParseContentSearchParam_Phrase(t *testing.T) {
	q, err := ParseContentSearchParam(`"blood pressure"`)
	if err != nil {
		t.Fatalf("ParseContentSearchParam error: %v", err)
	}
	if !strings.Contains(q.TSQuery, "phraseto_tsquery") {
		t.Errorf("phrase should use phraseto_tsquery, got %q", q.TSQuery)
	}
}

// ---------------------------------------------------------------------------
// BuildHighlightSQL tests
// ---------------------------------------------------------------------------

func TestBuildHighlightSQL_Basic(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"text_div"},
		Language:    "english",
	}
	query := &FullTextQuery{
		RawQuery: "diabetes",
		TSQuery:  "plainto_tsquery('english', 'diabetes')",
		Language: "english",
		HighlightOpts: &HighlightOptions{
			MaxWords:    35,
			MinWords:    15,
			MaxFragments: 3,
			StartSel:    "<b>",
			StopSel:     "</b>",
		},
	}
	got := BuildHighlightSQL(config, query)
	if !strings.Contains(got, "ts_headline") {
		t.Errorf("highlight SQL should contain ts_headline, got %q", got)
	}
	if !strings.Contains(got, "'english'") {
		t.Errorf("highlight SQL should contain language, got %q", got)
	}
}

func TestBuildHighlightSQL_NilHighlightOpts(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"text_div"},
		Language:    "english",
	}
	query := &FullTextQuery{
		RawQuery:      "diabetes",
		TSQuery:       "plainto_tsquery('english', 'diabetes')",
		Language:       "english",
		HighlightOpts: nil,
	}
	got := BuildHighlightSQL(config, query)
	// Should still produce valid SQL with defaults
	if !strings.Contains(got, "ts_headline") {
		t.Errorf("highlight SQL should contain ts_headline even with nil opts, got %q", got)
	}
}

func TestBuildHighlightSQL_CustomSelectors(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"text_div"},
		Language:    "english",
	}
	query := &FullTextQuery{
		RawQuery: "test",
		TSQuery:  "plainto_tsquery('english', 'test')",
		Language: "english",
		HighlightOpts: &HighlightOptions{
			MaxWords:    50,
			MinWords:    10,
			MaxFragments: 2,
			StartSel:    "<em>",
			StopSel:     "</em>",
		},
	}
	got := BuildHighlightSQL(config, query)
	if !strings.Contains(got, "StartSel=<em>") {
		t.Errorf("highlight SQL should contain custom StartSel, got %q", got)
	}
	if !strings.Contains(got, "StopSel=</em>") {
		t.Errorf("highlight SQL should contain custom StopSel, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// FullTextResult type tests
// ---------------------------------------------------------------------------

func TestFullTextResult_Fields(t *testing.T) {
	result := FullTextResult{
		Rank:      0.75,
		Headline:  "<b>diabetes</b> mellitus type 2",
		MatchedOn: []string{"text_div", "code_display"},
	}
	if result.Rank != 0.75 {
		t.Errorf("Rank = %f, want 0.75", result.Rank)
	}
	if result.Headline == "" {
		t.Error("Headline should not be empty")
	}
	if len(result.MatchedOn) != 2 {
		t.Errorf("MatchedOn length = %d, want 2", len(result.MatchedOn))
	}
}

// ---------------------------------------------------------------------------
// FullTextConfig type tests
// ---------------------------------------------------------------------------

func TestFullTextConfig_Fields(t *testing.T) {
	config := FullTextConfig{
		ResourceType: "Patient",
		TextColumns:  []string{"family_name", "given_name"},
		WeightMap:    map[string]string{"family_name": "A"},
		Language:     "english",
		IndexName:    "idx_patient_ft",
	}
	if config.ResourceType != "Patient" {
		t.Errorf("ResourceType = %q, want %q", config.ResourceType, "Patient")
	}
	if len(config.TextColumns) != 2 {
		t.Errorf("TextColumns length = %d, want 2", len(config.TextColumns))
	}
}

// ---------------------------------------------------------------------------
// HighlightOptions type tests
// ---------------------------------------------------------------------------

func TestHighlightOptions_Fields(t *testing.T) {
	opts := HighlightOptions{
		MaxWords:    35,
		MinWords:    15,
		MaxFragments: 3,
		StartSel:    "<b>",
		StopSel:     "</b>",
	}
	if opts.MaxWords != 35 {
		t.Errorf("MaxWords = %d, want 35", opts.MaxWords)
	}
	if opts.MinWords != 15 {
		t.Errorf("MinWords = %d, want 15", opts.MinWords)
	}
	if opts.MaxFragments != 3 {
		t.Errorf("MaxFragments = %d, want 3", opts.MaxFragments)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestParseFullTextQuery_VeryLongQuery(t *testing.T) {
	longQuery := strings.Repeat("diabetes ", 100)
	q, err := ParseFullTextQuery(longQuery, "english")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error for long query: %v", err)
	}
	if q.TSQuery == "" {
		t.Error("TSQuery should not be empty for long query")
	}
}

func TestParseFullTextQuery_OnlyOperators(t *testing.T) {
	_, err := ParseFullTextQuery("+ - |", "english")
	if err == nil {
		t.Error("ParseFullTextQuery should return error for operator-only query")
	}
}

func TestParseFullTextQuery_NestedQuotes(t *testing.T) {
	q, err := ParseFullTextQuery(`"hello "world""`, "english")
	// Should not crash; behavior may vary
	if err != nil {
		// Acceptable to error on malformed quotes
		return
	}
	if q.TSQuery == "" {
		t.Error("TSQuery should not be empty")
	}
}

func TestEscapeFullTextQuery_AllSpecialChars(t *testing.T) {
	special := `&|!():*\`
	got := EscapeFullTextQuery(special)
	// Should not contain any unescaped special characters
	if got == special {
		t.Errorf("EscapeFullTextQuery should escape special chars, got %q", got)
	}
}

func TestGenerateTSVector_SingleColumnNoWeight(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{"text_div"},
		Language:    "english",
	}
	got := GenerateTSVector(config)
	if strings.Contains(got, "setweight") {
		t.Errorf("no WeightMap should not produce setweight, got %q", got)
	}
}

func TestFullTextSearchClause_EmptyConfig(t *testing.T) {
	config := &FullTextConfig{
		TextColumns: []string{},
		Language:    "english",
	}
	query := &FullTextQuery{
		TSQuery:  "plainto_tsquery('english', 'test')",
		Language: "english",
	}
	clause, args := FullTextSearchClause(config, query, 1)
	// Empty config should return a "no-match" clause
	if clause != "1=0" {
		t.Errorf("empty config should return 1=0, got %q", clause)
	}
	if len(args) != 0 {
		t.Errorf("empty config should return no args, got %d", len(args))
	}
}

func TestCreateFullTextIndex_EmptyColumns(t *testing.T) {
	config := &FullTextConfig{
		ResourceType: "Patient",
		TextColumns:  []string{},
		Language:     "english",
		IndexName:    "idx_patient_fulltext",
	}
	got := CreateFullTextIndex(config)
	if got != "" {
		t.Errorf("should return empty for no columns, got %q", got)
	}
}

func TestApplyFullTextSearch_ChainingWithOtherClauses(t *testing.T) {
	engine := NewFullTextSearchEngine()
	engine.RegisterConfig(&FullTextConfig{
		ResourceType: "Patient",
		TextColumns:  []string{"family_name", "text_div"},
		WeightMap:    map[string]string{"family_name": "A", "text_div": "B"},
		Language:     "english",
	})

	q := NewSearchQuery("patients", "id, resource")
	q.Add("active = $1", true)
	q.AddString("family_name", "Smith", "")

	err := engine.ApplyFullTextSearch(q, "_text", "diabetes")
	if err != nil {
		t.Fatalf("ApplyFullTextSearch error: %v", err)
	}

	sql := q.CountSQL()
	if !strings.Contains(sql, "active = $1") {
		t.Errorf("SQL should contain active clause, got: %s", sql)
	}
	if !strings.Contains(sql, "@@") {
		t.Errorf("SQL should contain full-text search, got: %s", sql)
	}
}

func TestGenerateTSQuery_LanguageVariant(t *testing.T) {
	got := GenerateTSQuery("test", "simple")
	if !strings.Contains(got, "'simple'") {
		t.Errorf("GenerateTSQuery should use specified language, got %q", got)
	}
}

func TestParseFullTextQuery_MixedOperatorsAndWords(t *testing.T) {
	q, err := ParseFullTextQuery("+diabetes -juvenile type", "english")
	if err != nil {
		t.Fatalf("ParseFullTextQuery error: %v", err)
	}
	if q.TSQuery == "" {
		t.Error("TSQuery should not be empty for mixed query")
	}
}

func TestFullTextSearchEngine_DefaultConfigs(t *testing.T) {
	engine := NewFullTextSearchEngine()
	// Engine should start with default configs
	if len(engine.Configs) < 8 {
		t.Errorf("engine should have at least 8 default configs, got %d", len(engine.Configs))
	}
}
