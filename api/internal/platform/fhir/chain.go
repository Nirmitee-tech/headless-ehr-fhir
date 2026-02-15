package fhir

import (
	"context"
	"fmt"
	"strings"
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
	registry *IncludeRegistry
}

// NewChainResolver creates a new ChainResolver.
func NewChainResolver(registry *IncludeRegistry) *ChainResolver {
	return &ChainResolver{registry: registry}
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
