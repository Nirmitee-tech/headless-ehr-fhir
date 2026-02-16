package fhir

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ChainedParam represents a parsed chained search parameter.
// Example: "subject:Patient.name=John" -> SourceParam="subject", TargetType="Patient", TargetParam="name", Value="John"
type ChainedParam struct {
	SourceParam string // The reference search parameter on the source resource
	TargetType  string // The target resource type
	TargetParam string // The search parameter on the target resource
	Value       string // The search value
}

// HasParam represents a parsed _has search parameter.
// Example: "_has:Observation:subject:code=1234" -> TargetType="Observation", TargetParam="subject", SearchParam="code", Value="1234"
type HasParam struct {
	TargetType  string // The resource type that has a reference to the current resource
	TargetParam string // The reference search parameter on the target resource
	SearchParam string // The search parameter to filter on the target resource
	Value       string // The value to match
	Modifier    string // Optional chained _has modifier
}

// ChainedSearchConfig defines the database mapping for a single chain segment.
// It describes how a reference parameter on a source resource maps to a target
// table and which column on the target should be searched.
type ChainedSearchConfig struct {
	ReferenceColumn string          // The column in the source table that holds the reference (e.g. "practitioner_id")
	TargetTable     string          // The table being chained to (e.g. "practitioners")
	TargetColumn    string          // The column in the target table to search (e.g. "name")
	TargetType      SearchParamType // The type of the target search parameter
	TargetSysColumn string          // For token params on the target (e.g. "identifier_system")
}

// chainRegistryKey uniquely identifies a chain configuration entry.
type chainRegistryKey struct {
	sourceResource string
	refParam       string
	targetParam    string
}

// ChainRegistry maps chain paths to their database configurations.
// It is safe for concurrent reads and writes.
type ChainRegistry struct {
	mu      sync.RWMutex
	configs map[chainRegistryKey]ChainedSearchConfig
}

// NewChainRegistry creates a new empty ChainRegistry.
func NewChainRegistry() *ChainRegistry {
	return &ChainRegistry{
		configs: make(map[chainRegistryKey]ChainedSearchConfig),
	}
}

// Register adds a chain configuration mapping.
// sourceResource is the FHIR resource type being searched (e.g. "Patient").
// refParam is the reference search parameter name (e.g. "general-practitioner").
// targetParam is the search parameter on the target resource (e.g. "name").
// config describes the database columns and tables involved.
func (r *ChainRegistry) Register(sourceResource, refParam, targetParam string, config ChainedSearchConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := chainRegistryKey{
		sourceResource: sourceResource,
		refParam:       refParam,
		targetParam:    targetParam,
	}
	r.configs[key] = config
}

// Resolve looks up the chain configuration for a given source resource and chain path.
// The chainPath has the format "refParam.targetParam" (e.g. "general-practitioner.name").
func (r *ChainRegistry) Resolve(sourceResource, chainPath string) (*ChainedSearchConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dotIdx := strings.Index(chainPath, ".")
	if dotIdx < 0 {
		return nil, fmt.Errorf("invalid chain path %q: missing dot separator", chainPath)
	}

	refParam := chainPath[:dotIdx]
	targetParam := chainPath[dotIdx+1:]

	// Strip optional :Type modifier from refParam.
	if colonIdx := strings.Index(refParam, ":"); colonIdx >= 0 {
		refParam = refParam[:colonIdx]
	}

	key := chainRegistryKey{
		sourceResource: sourceResource,
		refParam:       refParam,
		targetParam:    targetParam,
	}

	config, ok := r.configs[key]
	if !ok {
		return nil, fmt.Errorf("unknown chain path %q for resource %q", chainPath, sourceResource)
	}

	return &config, nil
}

// MaxChainDepth is the maximum number of chain levels allowed per the FHIR specification.
const MaxChainDepth = 3

// ParseChainedParam parses a chained search parameter.
// Format: "param:ResourceType.targetParam" or "param.targetParam" (when type is unambiguous)
func ParseChainedParam(paramName string) (*ChainedParam, bool) {
	// Look for pattern: something.something (with optional :Type in between)
	dotIdx := strings.Index(paramName, ".")
	if dotIdx < 0 {
		return nil, false
	}

	sourceAndType := paramName[:dotIdx]
	targetParam := paramName[dotIdx+1:]

	if targetParam == "" {
		return nil, false
	}

	// Check for :Type modifier
	parts := strings.SplitN(sourceAndType, ":", 2)
	result := &ChainedParam{
		SourceParam: parts[0],
		TargetParam: targetParam,
	}

	if len(parts) == 2 {
		result.TargetType = parts[1]
	}

	return result, true
}

// ParseHasParam parses a _has search parameter value.
// Format: "_has:ResourceType:referenceParam:searchParam=value"
func ParseHasParam(paramName string) (*HasParam, bool) {
	if !strings.HasPrefix(paramName, "_has:") {
		return nil, false
	}

	rest := strings.TrimPrefix(paramName, "_has:")
	parts := strings.SplitN(rest, ":", 3)
	if len(parts) != 3 {
		return nil, false
	}

	return &HasParam{
		TargetType:  parts[0],
		TargetParam: parts[1],
		SearchParam: parts[2],
	}, true
}

// ChainResolver resolves chained search parameters by looking up referenced resources.
type ChainResolver struct {
	registry      *IncludeRegistry
	chainRegistry *ChainRegistry
}

// NewChainResolver creates a new ChainResolver.
func NewChainResolver(registry *IncludeRegistry) *ChainResolver {
	return &ChainResolver{registry: registry}
}

// NewChainResolverWithRegistry creates a ChainResolver that uses a ChainRegistry
// for SQL-based chain resolution instead of in-memory ID collection.
func NewChainResolverWithRegistry(includeReg *IncludeRegistry, chainReg *ChainRegistry) *ChainResolver {
	return &ChainResolver{
		registry:      includeReg,
		chainRegistry: chainReg,
	}
}

// ResolveChainedParam resolves a chained parameter to a set of IDs.
// It searches the target resource type and returns the IDs of matching resources,
// which can then be used in an IN clause for the source resource's reference column.
func (cr *ChainResolver) ResolveChainedParam(ctx context.Context, chain *ChainedParam) ([]string, error) {
	if cr.registry == nil {
		return nil, fmt.Errorf("no include registry configured")
	}

	// This is a simplified implementation that returns the concept.
	// Full implementation would execute a search on the target resource
	// and collect the matching IDs.
	// For now, return an empty list to indicate no matches,
	// which effectively makes chained params a no-op until
	// full search infrastructure is wired.
	return nil, fmt.Errorf("chained parameter resolution requires search execution on %s", chain.TargetType)
}

// ResolveHasParam resolves a _has parameter by searching the target resource type
// for resources that reference the current resource type.
func (cr *ChainResolver) ResolveHasParam(ctx context.Context, has *HasParam, currentResourceType string) ([]string, error) {
	if cr.registry == nil {
		return nil, fmt.Errorf("no include registry configured")
	}

	return nil, fmt.Errorf("_has parameter resolution requires search execution on %s", has.TargetType)
}

// BuildChainedINClause generates an IN clause for chained parameter resolution.
// The ids parameter contains the resolved IDs from the target resource search.
func BuildChainedINClause(column string, ids []string, argIdx int) (string, []interface{}, int) {
	if len(ids) == 0 {
		return "1=0", nil, argIdx // No matches
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", argIdx+i)
		args[i] = id
	}

	clause := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
	return clause, args, argIdx + len(ids)
}

// ChainedSearchClause generates a SQL subquery clause for a chained search parameter.
// It produces a "column IN (SELECT id FROM target_table WHERE condition)" clause.
// The config describes the chain mapping, targetValue is the search value, and
// startIdx is the current positional parameter index.
// Returns the SQL clause, bind arguments, and the next available parameter index.
func ChainedSearchClause(config ChainedSearchConfig, targetValue string, startIdx int) (string, []interface{}, int) {
	var innerClause string
	var args []interface{}
	nextIdx := startIdx

	switch config.TargetType {
	case SearchParamString:
		innerClause = fmt.Sprintf("%s ILIKE $%d", config.TargetColumn, nextIdx)
		args = append(args, targetValue+"%")
		nextIdx++
	case SearchParamToken:
		if config.TargetSysColumn != "" && strings.Contains(targetValue, "|") {
			// system|code token search on target
			tClause, tArgs, tNext := TokenSearchClause(config.TargetSysColumn, config.TargetColumn, targetValue, nextIdx)
			innerClause = tClause
			args = append(args, tArgs...)
			nextIdx = tNext
		} else {
			innerClause = fmt.Sprintf("%s = $%d", config.TargetColumn, nextIdx)
			args = append(args, targetValue)
			nextIdx++
		}
	case SearchParamDate:
		dClause, dArgs, dNext := DateSearchClause(config.TargetColumn, targetValue, nextIdx)
		innerClause = dClause
		args = append(args, dArgs...)
		nextIdx = dNext
	case SearchParamReference:
		rClause, rArgs, rNext := ReferenceSearchClause(config.TargetColumn, targetValue, nextIdx)
		innerClause = rClause
		args = append(args, rArgs...)
		nextIdx = rNext
	case SearchParamNumber, SearchParamQuantity:
		nClause, nArgs, nNext := NumberSearchClause(config.TargetColumn, targetValue, nextIdx)
		innerClause = nClause
		args = append(args, nArgs...)
		nextIdx = nNext
	default:
		// URI or unknown: exact match
		innerClause = fmt.Sprintf("%s = $%d", config.TargetColumn, nextIdx)
		args = append(args, targetValue)
		nextIdx++
	}

	sql := fmt.Sprintf("%s IN (SELECT id FROM %s WHERE %s)",
		config.ReferenceColumn, config.TargetTable, innerClause)

	return sql, args, nextIdx
}

// MultiLevelChainedSearchClause generates a nested subquery for multi-level chains.
// configs is a slice of ChainedSearchConfig entries from outermost to innermost chain.
// targetValue is the value to match against the innermost chain target.
// startIdx is the starting positional parameter index.
// Returns the SQL clause, bind arguments, and the next available parameter index.
//
// For a 2-level chain like Patient?general-practitioner.organization.name=Acme:
//
//	configs[0]: Patient -> Practitioner (general-practitioner)
//	configs[1]: Practitioner -> Organization (organization)
//
// Produces:
//
//	practitioner_id IN (SELECT id FROM practitioners WHERE
//	  organization_id IN (SELECT id FROM organizations WHERE name ILIKE $1))
func MultiLevelChainedSearchClause(configs []ChainedSearchConfig, targetValue string, startIdx int) (string, []interface{}, int, error) {
	if len(configs) == 0 {
		return "", nil, startIdx, fmt.Errorf("no chain configs provided")
	}
	if len(configs) > MaxChainDepth {
		return "", nil, startIdx, fmt.Errorf("chain depth %d exceeds maximum of %d", len(configs), MaxChainDepth)
	}

	// Single level: delegate to the simple clause builder.
	if len(configs) == 1 {
		clause, args, nextIdx := ChainedSearchClause(configs[0], targetValue, startIdx)
		return clause, args, nextIdx, nil
	}

	// Multi-level: build from the inside out.
	// The innermost config generates the leaf condition.
	innermost := configs[len(configs)-1]

	// Build the innermost WHERE condition.
	var innerWhere string
	var args []interface{}
	nextIdx := startIdx

	switch innermost.TargetType {
	case SearchParamString:
		innerWhere = fmt.Sprintf("%s ILIKE $%d", innermost.TargetColumn, nextIdx)
		args = append(args, targetValue+"%")
		nextIdx++
	case SearchParamToken:
		if innermost.TargetSysColumn != "" && strings.Contains(targetValue, "|") {
			tClause, tArgs, tNext := TokenSearchClause(innermost.TargetSysColumn, innermost.TargetColumn, targetValue, nextIdx)
			innerWhere = tClause
			args = append(args, tArgs...)
			nextIdx = tNext
		} else {
			innerWhere = fmt.Sprintf("%s = $%d", innermost.TargetColumn, nextIdx)
			args = append(args, targetValue)
			nextIdx++
		}
	case SearchParamDate:
		dClause, dArgs, dNext := DateSearchClause(innermost.TargetColumn, targetValue, nextIdx)
		innerWhere = dClause
		args = append(args, dArgs...)
		nextIdx = dNext
	default:
		innerWhere = fmt.Sprintf("%s = $%d", innermost.TargetColumn, nextIdx)
		args = append(args, targetValue)
		nextIdx++
	}

	// Wrap with inner subquery: ref_col IN (SELECT id FROM target WHERE condition)
	subquery := fmt.Sprintf("%s IN (SELECT id FROM %s WHERE %s)",
		innermost.ReferenceColumn, innermost.TargetTable, innerWhere)

	// Wrap each remaining config from inside out (going from second-to-last to first).
	for i := len(configs) - 2; i >= 0; i-- {
		cfg := configs[i]
		// The current subquery becomes the WHERE of the next outer level.
		// For the outermost level (i==0), the reference column is on the source table directly.
		if i == 0 {
			subquery = fmt.Sprintf("%s IN (SELECT id FROM %s WHERE %s)",
				cfg.ReferenceColumn, cfg.TargetTable, subquery)
		} else {
			subquery = fmt.Sprintf("%s IN (SELECT id FROM %s WHERE %s)",
				cfg.ReferenceColumn, cfg.TargetTable, subquery)
		}
	}

	return subquery, args, nextIdx, nil
}

// ReverseChainClause generates a SQL subquery for reverse chaining (_has).
// It produces an "id IN (SELECT ref_col FROM target_table WHERE search_condition)" clause.
//
// sourceTable is the table of the current resource (e.g. "patients").
// sourceIdCol is the id column name on the source table (e.g. "id").
// targetTable is the table of the referencing resource (e.g. "observations").
// targetRefCol is the column in the target table that references the source (e.g. "patient_id").
// targetSearchCol is the column in the target table to filter on (e.g. "code").
// targetType is the search parameter type for the target search column.
// value is the search value.
// startIdx is the current positional parameter index.
//
// Returns the SQL clause, bind arguments, and the next available parameter index.
func ReverseChainClause(sourceTable, sourceIdCol, targetTable, targetRefCol, targetSearchCol string, targetType SearchParamType, value string, startIdx int) (string, []interface{}, int) {
	var innerClause string
	var args []interface{}
	nextIdx := startIdx

	switch targetType {
	case SearchParamString:
		innerClause = fmt.Sprintf("%s ILIKE $%d", targetSearchCol, nextIdx)
		args = append(args, value+"%")
		nextIdx++
	case SearchParamToken:
		innerClause = fmt.Sprintf("%s = $%d", targetSearchCol, nextIdx)
		args = append(args, value)
		nextIdx++
	case SearchParamDate:
		dClause, dArgs, dNext := DateSearchClause(targetSearchCol, value, nextIdx)
		innerClause = dClause
		args = append(args, dArgs...)
		nextIdx = dNext
	case SearchParamNumber, SearchParamQuantity:
		nClause, nArgs, nNext := NumberSearchClause(targetSearchCol, value, nextIdx)
		innerClause = nClause
		args = append(args, nArgs...)
		nextIdx = nNext
	default:
		innerClause = fmt.Sprintf("%s = $%d", targetSearchCol, nextIdx)
		args = append(args, value)
		nextIdx++
	}

	sql := fmt.Sprintf("%s IN (SELECT %s FROM %s WHERE %s)",
		sourceIdCol, targetRefCol, targetTable, innerClause)

	return sql, args, nextIdx
}

// ReverseChainTokenClause generates a SQL subquery for reverse chaining (_has) with
// token-type search parameters that support system|code format.
func ReverseChainTokenClause(sourceIdCol, targetTable, targetRefCol, targetSysCol, targetCodeCol, value string, startIdx int) (string, []interface{}, int) {
	tClause, tArgs, nextIdx := TokenSearchClause(targetSysCol, targetCodeCol, value, startIdx)

	sql := fmt.Sprintf("%s IN (SELECT %s FROM %s WHERE %s)",
		sourceIdCol, targetRefCol, targetTable, tClause)

	return sql, tArgs, nextIdx
}

// DefaultChainRegistry returns a ChainRegistry pre-configured with common FHIR chain paths.
// This covers the most frequently used chained searches across standard FHIR resources.
func DefaultChainRegistry() *ChainRegistry {
	r := NewChainRegistry()

	// Patient.general-practitioner -> Practitioner
	r.Register("Patient", "general-practitioner", "name", ChainedSearchConfig{
		ReferenceColumn: "practitioner_id",
		TargetTable:     "practitioners",
		TargetColumn:    "name",
		TargetType:      SearchParamString,
	})
	r.Register("Patient", "general-practitioner", "identifier", ChainedSearchConfig{
		ReferenceColumn: "practitioner_id",
		TargetTable:     "practitioners",
		TargetColumn:    "identifier_value",
		TargetType:      SearchParamToken,
		TargetSysColumn: "identifier_system",
	})

	// Observation.patient -> Patient
	r.Register("Observation", "patient", "name", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "last_name",
		TargetType:      SearchParamString,
	})
	r.Register("Observation", "patient", "identifier", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "identifier_value",
		TargetType:      SearchParamToken,
		TargetSysColumn: "identifier_system",
	})
	r.Register("Observation", "patient", "birthdate", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "birth_date",
		TargetType:      SearchParamDate,
	})

	// Observation.encounter -> Encounter
	r.Register("Observation", "encounter", "date", ChainedSearchConfig{
		ReferenceColumn: "encounter_id",
		TargetTable:     "encounters",
		TargetColumn:    "period_start",
		TargetType:      SearchParamDate,
	})
	r.Register("Observation", "encounter", "status", ChainedSearchConfig{
		ReferenceColumn: "encounter_id",
		TargetTable:     "encounters",
		TargetColumn:    "status",
		TargetType:      SearchParamToken,
	})

	// Encounter.patient -> Patient
	r.Register("Encounter", "patient", "name", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "last_name",
		TargetType:      SearchParamString,
	})
	r.Register("Encounter", "patient", "identifier", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "identifier_value",
		TargetType:      SearchParamToken,
		TargetSysColumn: "identifier_system",
	})

	// MedicationRequest.patient -> Patient
	r.Register("MedicationRequest", "patient", "name", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "last_name",
		TargetType:      SearchParamString,
	})
	r.Register("MedicationRequest", "patient", "identifier", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "identifier_value",
		TargetType:      SearchParamToken,
		TargetSysColumn: "identifier_system",
	})

	// Condition.patient -> Patient
	r.Register("Condition", "patient", "name", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "last_name",
		TargetType:      SearchParamString,
	})
	r.Register("Condition", "patient", "identifier", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "identifier_value",
		TargetType:      SearchParamToken,
		TargetSysColumn: "identifier_system",
	})

	// Procedure.patient -> Patient
	r.Register("Procedure", "patient", "name", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "last_name",
		TargetType:      SearchParamString,
	})
	r.Register("Procedure", "patient", "identifier", ChainedSearchConfig{
		ReferenceColumn: "patient_id",
		TargetTable:     "patients",
		TargetColumn:    "identifier_value",
		TargetType:      SearchParamToken,
		TargetSysColumn: "identifier_system",
	})

	return r
}
