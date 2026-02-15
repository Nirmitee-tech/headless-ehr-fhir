package openapi

import (
	"net/http"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/labstack/echo/v4"
)

// Generator builds an OpenAPI 3.0 spec from the CapabilityBuilder.
type Generator struct {
	capBuilder *fhir.CapabilityBuilder
	version    string
	baseURL    string
}

// NewGenerator creates a new OpenAPI spec generator.
func NewGenerator(capBuilder *fhir.CapabilityBuilder, version, baseURL string) *Generator {
	return &Generator{capBuilder: capBuilder, version: version, baseURL: baseURL}
}

// GenerateSpec produces the OpenAPI 3.0 spec as a map.
func (g *Generator) GenerateSpec() map[string]interface{} {
	cap := g.capBuilder.Build()

	paths := make(map[string]interface{})
	var resourceTypes []string

	// Extract resources from capability statement
	restArray, _ := cap["rest"].([]map[string]interface{})
	if len(restArray) > 0 {
		resources, _ := restArray[0]["resource"].([]map[string]interface{})
		for _, res := range resources {
			resType, _ := res["type"].(string)
			if resType == "" {
				continue
			}
			resourceTypes = append(resourceTypes, resType)

			// Extract search params for this resource
			searchParams := g.extractSearchParams(res)

			// Build search parameters for GET operations
			searchQueryParams := g.buildSearchParameters(searchParams)

			// Build request body for POST/PUT
			requestBody := g.buildRequestBody(resType)

			// Add search path
			searchPath := "/fhir/" + resType
			paths[searchPath] = map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Search " + resType,
					"operationId": "search" + resType,
					"tags":        []string{resType},
					"parameters":  searchQueryParams,
					"responses": map[string]interface{}{
						"200": g.buildResponseWithSchema("Search results Bundle", "#/components/schemas/Bundle"),
					},
				},
				"post": map[string]interface{}{
					"summary":     "Create " + resType,
					"operationId": "create" + resType,
					"tags":        []string{resType},
					"requestBody": requestBody,
					"responses": map[string]interface{}{
						"201": g.buildResponseWithSchema("Created", "#/components/schemas/"+resType),
					},
				},
			}

			// Add read path
			readPath := "/fhir/" + resType + "/{id}"
			paths[readPath] = map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Read " + resType,
					"operationId": "read" + resType,
					"tags":        []string{resType},
					"parameters": []map[string]interface{}{
						{"name": "id", "in": "path", "required": true, "schema": map[string]string{"type": "string"}},
					},
					"responses": map[string]interface{}{
						"200": g.buildResponseWithSchema("Success", "#/components/schemas/"+resType),
						"404": g.buildResponseWithSchema("Not Found", "#/components/schemas/OperationOutcome"),
					},
				},
				"put": map[string]interface{}{
					"summary":     "Update " + resType,
					"operationId": "update" + resType,
					"tags":        []string{resType},
					"requestBody": requestBody,
					"parameters": []map[string]interface{}{
						{"name": "id", "in": "path", "required": true, "schema": map[string]string{"type": "string"}},
					},
					"responses": map[string]interface{}{
						"200": g.buildResponseWithSchema("Updated", "#/components/schemas/"+resType),
					},
				},
				"delete": map[string]interface{}{
					"summary":     "Delete " + resType,
					"operationId": "delete" + resType,
					"tags":        []string{resType},
					"parameters": []map[string]interface{}{
						{"name": "id", "in": "path", "required": true, "schema": map[string]string{"type": "string"}},
					},
					"responses": map[string]interface{}{
						"204": map[string]interface{}{
							"description": "Deleted",
						},
					},
				},
			}
		}
	}

	spec := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "Headless EHR FHIR R4 API",
			"version":     g.version,
			"description": "FHIR R4 compliant EHR API",
		},
		"servers": []map[string]string{
			{"url": g.baseURL},
		},
		"paths": paths,
		"components": map[string]interface{}{
			"schemas": buildComponentSchemas(resourceTypes),
		},
	}

	return spec
}

// extractSearchParams extracts search parameter definitions from a capability
// statement resource entry.
func (g *Generator) extractSearchParams(res map[string]interface{}) []searchParamDef {
	var params []searchParamDef
	rawParams, _ := res["searchParam"].([]map[string]string)
	for _, sp := range rawParams {
		params = append(params, searchParamDef{
			Name: sp["name"],
			Type: sp["type"],
		})
	}
	return params
}

// searchParamDef holds a search parameter name and FHIR type.
type searchParamDef struct {
	Name string
	Type string
}

// buildSearchParameters builds the OpenAPI parameter array for a GET search
// operation, including both resource-specific params and common FHIR params.
func (g *Generator) buildSearchParameters(params []searchParamDef) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(params)+4)

	for _, p := range params {
		result = append(result, map[string]interface{}{
			"name":   p.Name,
			"in":     "query",
			"schema": fhirSearchParamSchema(p.Type),
		})
	}

	// Add common FHIR search parameters
	commonParams := []struct {
		name   string
		schema map[string]interface{}
		desc   string
	}{
		{"_count", map[string]interface{}{"type": "integer", "minimum": 0}, "Number of results per page"},
		{"_offset", map[string]interface{}{"type": "integer", "minimum": 0}, "Starting index for results"},
		{"_sort", map[string]interface{}{"type": "string"}, "Sort order (prefix with - for descending)"},
		{"_include", map[string]interface{}{"type": "string"}, "Include referenced resources"},
		{"_revinclude", map[string]interface{}{"type": "string"}, "Include resources that reference this resource"},
	}

	for _, cp := range commonParams {
		result = append(result, map[string]interface{}{
			"name":        cp.name,
			"in":          "query",
			"schema":      cp.schema,
			"description": cp.desc,
		})
	}

	return result
}

// fhirSearchParamSchema maps a FHIR search parameter type to an OpenAPI schema.
func fhirSearchParamSchema(fhirType string) map[string]interface{} {
	switch fhirType {
	case "date":
		return map[string]interface{}{"type": "string", "format": "date"}
	case "number", "quantity":
		return map[string]interface{}{"type": "string"}
	case "uri":
		return map[string]interface{}{"type": "string", "format": "uri"}
	default:
		// string, token, reference, composite, special all use string
		return map[string]interface{}{"type": "string"}
	}
}

// buildRequestBody creates the OpenAPI requestBody for POST/PUT operations.
func (g *Generator) buildRequestBody(resType string) map[string]interface{} {
	return map[string]interface{}{
		"required": true,
		"content": map[string]interface{}{
			"application/fhir+json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/" + resType,
				},
			},
		},
	}
}

// buildResponseWithSchema creates an OpenAPI response with content schema reference.
func (g *Generator) buildResponseWithSchema(description, schemaRef string) map[string]interface{} {
	return map[string]interface{}{
		"description": description,
		"content": map[string]interface{}{
			"application/fhir+json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": schemaRef,
				},
			},
		},
	}
}

// buildComponentSchemas creates the OpenAPI component schemas including core
// FHIR data types and resource-specific schemas.
func buildComponentSchemas(resourceTypes []string) map[string]interface{} {
	schemas := make(map[string]interface{})

	// Core data type schemas
	schemas["Meta"] = buildMetaSchema()
	schemas["Coding"] = buildCodingSchema()
	schemas["CodeableConcept"] = buildCodeableConceptSchema()
	schemas["Reference"] = buildReferenceSchema()
	schemas["Identifier"] = buildIdentifierSchema()
	schemas["HumanName"] = buildHumanNameSchema()
	schemas["Address"] = buildAddressSchema()
	schemas["ContactPoint"] = buildContactPointSchema()
	schemas["Period"] = buildPeriodSchema()
	schemas["Quantity"] = buildQuantitySchema()

	// Core resource schemas
	schemas["Bundle"] = buildBundleSchema()
	schemas["BundleEntry"] = buildBundleEntrySchema()
	schemas["OperationOutcome"] = buildOperationOutcomeSchema()

	// Resource schemas for each registered resource type
	for _, rt := range resourceTypes {
		if _, exists := schemas[rt]; !exists {
			schemas[rt] = buildResourceSchema(rt)
		}
	}

	return schemas
}

// ── Core data type schemas ──────────────────────────────────────────────

func buildMetaSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"versionId":   map[string]interface{}{"type": "string"},
			"lastUpdated": map[string]interface{}{"type": "string", "format": "date-time"},
			"profile": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string", "format": "uri"},
			},
		},
	}
}

func buildCodingSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"system":  map[string]interface{}{"type": "string", "format": "uri"},
			"code":    map[string]interface{}{"type": "string"},
			"display": map[string]interface{}{"type": "string"},
		},
	}
}

func buildCodeableConceptSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"coding": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"$ref": "#/components/schemas/Coding"},
			},
			"text": map[string]interface{}{"type": "string"},
		},
	}
}

func buildReferenceSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"reference": map[string]interface{}{"type": "string"},
			"type":      map[string]interface{}{"type": "string"},
			"display":   map[string]interface{}{"type": "string"},
		},
	}
}

func buildIdentifierSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"use":    map[string]interface{}{"type": "string", "enum": []string{"usual", "official", "temp", "secondary", "old"}},
			"type":   map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
			"system": map[string]interface{}{"type": "string", "format": "uri"},
			"value":  map[string]interface{}{"type": "string"},
			"period": map[string]interface{}{"$ref": "#/components/schemas/Period"},
		},
	}
}

func buildHumanNameSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"use":    map[string]interface{}{"type": "string", "enum": []string{"usual", "official", "temp", "nickname", "anonymous", "old", "maiden"}},
			"family": map[string]interface{}{"type": "string"},
			"given": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"prefix": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"suffix": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
		},
	}
}

func buildAddressSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"use":  map[string]interface{}{"type": "string", "enum": []string{"home", "work", "temp", "old", "billing"}},
			"type": map[string]interface{}{"type": "string", "enum": []string{"postal", "physical", "both"}},
			"line": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"city":       map[string]interface{}{"type": "string"},
			"district":   map[string]interface{}{"type": "string"},
			"state":      map[string]interface{}{"type": "string"},
			"postalCode": map[string]interface{}{"type": "string"},
			"country":    map[string]interface{}{"type": "string"},
		},
	}
}

func buildContactPointSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"system": map[string]interface{}{"type": "string", "enum": []string{"phone", "fax", "email", "pager", "url", "sms", "other"}},
			"value":  map[string]interface{}{"type": "string"},
			"use":    map[string]interface{}{"type": "string", "enum": []string{"home", "work", "temp", "old", "mobile"}},
			"rank":   map[string]interface{}{"type": "integer", "minimum": 1},
		},
	}
}

func buildPeriodSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"start": map[string]interface{}{"type": "string", "format": "date-time"},
			"end":   map[string]interface{}{"type": "string", "format": "date-time"},
		},
	}
}

func buildQuantitySchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"value":      map[string]interface{}{"type": "number"},
			"comparator": map[string]interface{}{"type": "string", "enum": []string{"<", "<=", ">=", ">"}},
			"unit":       map[string]interface{}{"type": "string"},
			"system":     map[string]interface{}{"type": "string", "format": "uri"},
			"code":       map[string]interface{}{"type": "string"},
		},
	}
}

// ── Core resource schemas ───────────────────────────────────────────────

func buildBundleSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"resourceType": map[string]interface{}{"type": "string", "enum": []string{"Bundle"}},
			"id":           map[string]interface{}{"type": "string"},
			"type": map[string]interface{}{
				"type": "string",
				"enum": []string{"searchset", "batch", "transaction", "history", "document", "message", "collection", "batch-response", "transaction-response"},
			},
			"total": map[string]interface{}{"type": "integer", "minimum": 0},
			"link": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"relation": map[string]interface{}{"type": "string"},
						"url":      map[string]interface{}{"type": "string", "format": "uri"},
					},
				},
			},
			"entry": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"$ref": "#/components/schemas/BundleEntry"},
			},
		},
	}
}

func buildBundleEntrySchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"fullUrl":  map[string]interface{}{"type": "string", "format": "uri"},
			"resource": map[string]interface{}{"type": "object", "description": "The FHIR resource"},
			"search": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"mode": map[string]interface{}{"type": "string", "enum": []string{"match", "include", "outcome"}},
				},
			},
		},
	}
}

func buildOperationOutcomeSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"resourceType": map[string]interface{}{"type": "string", "enum": []string{"OperationOutcome"}},
			"issue": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"severity": map[string]interface{}{
							"type": "string",
							"enum": []string{"fatal", "error", "warning", "information"},
						},
						"code":        map[string]interface{}{"type": "string"},
						"diagnostics": map[string]interface{}{"type": "string"},
						"details":     map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
						"expression": map[string]interface{}{
							"type":  "array",
							"items": map[string]interface{}{"type": "string"},
						},
					},
					"required": []string{"severity", "code"},
				},
			},
		},
		"required": []string{"resourceType", "issue"},
	}
}

// ── Resource-specific schemas ───────────────────────────────────────────

// baseResourceProperties returns the properties common to all FHIR resources.
func baseResourceProperties(resourceType string) map[string]interface{} {
	return map[string]interface{}{
		"resourceType": map[string]interface{}{"type": "string", "enum": []string{resourceType}},
		"id":           map[string]interface{}{"type": "string", "format": "uuid"},
		"meta":         map[string]interface{}{"$ref": "#/components/schemas/Meta"},
	}
}

// mergeProperties merges additional properties into a base map.
func mergeProperties(base, extra map[string]interface{}) map[string]interface{} {
	for k, v := range extra {
		base[k] = v
	}
	return base
}

// resourceSchemaDefinitions returns the detailed property definitions for known
// FHIR resource types. Resources not listed here get only the base properties.
var resourceSchemaDefinitions = map[string]map[string]interface{}{
	"Patient": {
		"active": map[string]interface{}{"type": "boolean"},
		"name": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/HumanName"},
		},
		"telecom": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/ContactPoint"},
		},
		"gender":    map[string]interface{}{"type": "string", "enum": []string{"male", "female", "other", "unknown"}},
		"birthDate": map[string]interface{}{"type": "string", "format": "date"},
		"address": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Address"},
		},
		"maritalStatus": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
		"deceasedBoolean":  map[string]interface{}{"type": "boolean"},
		"deceasedDateTime": map[string]interface{}{"type": "string", "format": "date-time"},
		"communication": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"language":  map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"preferred": map[string]interface{}{"type": "boolean"},
				},
			},
		},
	},
	"Practitioner": {
		"active": map[string]interface{}{"type": "boolean"},
		"name": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/HumanName"},
		},
		"telecom": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/ContactPoint"},
		},
		"gender":    map[string]interface{}{"type": "string", "enum": []string{"male", "female", "other", "unknown"}},
		"birthDate": map[string]interface{}{"type": "string", "format": "date"},
		"address": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Address"},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
		"qualification": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"identifier": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
					},
					"code":   map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"period": map[string]interface{}{"$ref": "#/components/schemas/Period"},
					"issuer": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
				},
			},
		},
	},
	"Encounter": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"planned", "arrived", "triaged", "in-progress", "onleave", "finished", "cancelled", "entered-in-error", "unknown"},
		},
		"class": map[string]interface{}{"$ref": "#/components/schemas/Coding"},
		"type": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"subject": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"participant": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					},
					"period":     map[string]interface{}{"$ref": "#/components/schemas/Period"},
					"individual": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
				},
			},
		},
		"period": map[string]interface{}{"$ref": "#/components/schemas/Period"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
		"reasonCode": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"serviceProvider": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
	},
	"Observation": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"registered", "preliminary", "final", "amended", "corrected", "cancelled", "entered-in-error", "unknown"},
		},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"code":    map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"subject": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter":            map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"effectiveDateTime":    map[string]interface{}{"type": "string", "format": "date-time"},
		"effectivePeriod":      map[string]interface{}{"$ref": "#/components/schemas/Period"},
		"issued":               map[string]interface{}{"type": "string", "format": "date-time"},
		"valueQuantity":        map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
		"valueCodeableConcept": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"valueString":          map[string]interface{}{"type": "string"},
		"valueBoolean":         map[string]interface{}{"type": "boolean"},
		"valueInteger":         map[string]interface{}{"type": "integer"},
		"valueDateTime":        map[string]interface{}{"type": "string", "format": "date-time"},
		"interpretation": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
		"performer": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"referenceRange": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"low":  map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
					"high": map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
					"type": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"text": map[string]interface{}{"type": "string"},
				},
			},
		},
	},
	"Condition": {
		"clinicalStatus":      map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"verificationStatus":  map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"severity":        map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"code":            map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"subject":         map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter":       map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"onsetDateTime":   map[string]interface{}{"type": "string", "format": "date-time"},
		"abatementDateTime": map[string]interface{}{"type": "string", "format": "date-time"},
		"recordedDate":    map[string]interface{}{"type": "string", "format": "date-time"},
		"recorder":        map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"asserter":        map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
		"note": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{"type": "string"},
				},
			},
		},
	},
	"MedicationRequest": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"active", "on-hold", "cancelled", "completed", "entered-in-error", "stopped", "draft", "unknown"},
		},
		"intent": map[string]interface{}{
			"type": "string",
			"enum": []string{"proposal", "plan", "order", "original-order", "reflex-order", "filler-order", "instance-order", "option"},
		},
		"medicationCodeableConcept": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"medicationReference":       map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"subject":                   map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter":                 map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"authoredOn":                map[string]interface{}{"type": "string", "format": "date-time"},
		"requester":                 map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"reasonCode": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"dosageInstruction": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text":     map[string]interface{}{"type": "string"},
					"sequence": map[string]interface{}{"type": "integer"},
					"timing": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"repeat": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"frequency":  map[string]interface{}{"type": "integer"},
									"period":     map[string]interface{}{"type": "number"},
									"periodUnit": map[string]interface{}{"type": "string"},
								},
							},
						},
					},
					"route": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"doseAndRate": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"type":         map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
								"doseQuantity": map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
							},
						},
					},
				},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Organization": {
		"active": map[string]interface{}{"type": "boolean"},
		"type": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"name": map[string]interface{}{"type": "string"},
		"alias": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"type": "string"},
		},
		"telecom": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/ContactPoint"},
		},
		"address": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Address"},
		},
		"partOf": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Location": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"active", "suspended", "inactive"},
		},
		"name":        map[string]interface{}{"type": "string"},
		"description": map[string]interface{}{"type": "string"},
		"mode": map[string]interface{}{
			"type": "string",
			"enum": []string{"instance", "kind"},
		},
		"type": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"telecom": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/ContactPoint"},
		},
		"address":            map[string]interface{}{"$ref": "#/components/schemas/Address"},
		"managingOrganization": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"AllergyIntolerance": {
		"clinicalStatus":      map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"verificationStatus":  map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"type": map[string]interface{}{
			"type": "string",
			"enum": []string{"allergy", "intolerance"},
		},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"type": "string", "enum": []string{"food", "medication", "environment", "biologic"}},
		},
		"criticality": map[string]interface{}{
			"type": "string",
			"enum": []string{"low", "high", "unable-to-assess"},
		},
		"code":     map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"patient":  map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter":    map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"onsetDateTime": map[string]interface{}{"type": "string", "format": "date-time"},
		"recordedDate":  map[string]interface{}{"type": "string", "format": "date-time"},
		"recorder":      map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"asserter":      map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"reaction": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"substance": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"manifestation": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					},
					"severity": map[string]interface{}{
						"type": "string",
						"enum": []string{"mild", "moderate", "severe"},
					},
				},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Procedure": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"preparation", "in-progress", "not-done", "on-hold", "stopped", "completed", "entered-in-error", "unknown"},
		},
		"category": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"code":     map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"subject":  map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter":         map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"performedDateTime": map[string]interface{}{"type": "string", "format": "date-time"},
		"performedPeriod":   map[string]interface{}{"$ref": "#/components/schemas/Period"},
		"recorder":          map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"asserter":          map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"performer": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"function": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"actor":    map[string]interface{}{"$ref": "#/components/schemas/Reference"},
				},
			},
		},
		"reasonCode": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"bodySite": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
		"note": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{"type": "string"},
				},
			},
		},
	},
	"Medication": {
		"code":   map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"status": map[string]interface{}{"type": "string", "enum": []string{"active", "inactive", "entered-in-error"}},
		"manufacturer": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"form":         map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"amount":       map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"MedicationAdministration": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"in-progress", "not-done", "on-hold", "completed", "entered-in-error", "stopped", "unknown"},
		},
		"medicationCodeableConcept": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"medicationReference":       map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"subject":                   map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"context":                   map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"effectiveDateTime":         map[string]interface{}{"type": "string", "format": "date-time"},
		"effectivePeriod":           map[string]interface{}{"$ref": "#/components/schemas/Period"},
		"performer": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"function": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"actor":    map[string]interface{}{"$ref": "#/components/schemas/Reference"},
				},
			},
		},
		"dosage": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text":  map[string]interface{}{"type": "string"},
				"route": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
				"dose":  map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"MedicationDispense": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"preparation", "in-progress", "cancelled", "on-hold", "completed", "entered-in-error", "stopped", "declined", "unknown"},
		},
		"medicationCodeableConcept": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"medicationReference":       map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"subject":                   map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"context":                   map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"performer": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"function": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"actor":    map[string]interface{}{"$ref": "#/components/schemas/Reference"},
				},
			},
		},
		"quantity":     map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
		"daysSupply":   map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
		"whenPrepared": map[string]interface{}{"type": "string", "format": "date-time"},
		"whenHandedOver": map[string]interface{}{"type": "string", "format": "date-time"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"ServiceRequest": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"draft", "active", "on-hold", "revoked", "completed", "entered-in-error", "unknown"},
		},
		"intent": map[string]interface{}{
			"type": "string",
			"enum": []string{"proposal", "plan", "directive", "order", "original-order", "reflex-order", "filler-order", "instance-order", "option"},
		},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"priority": map[string]interface{}{
			"type": "string",
			"enum": []string{"routine", "urgent", "asap", "stat"},
		},
		"code":           map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"subject":        map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter":      map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"authoredOn":     map[string]interface{}{"type": "string", "format": "date-time"},
		"requester":      map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"performer": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"reasonCode": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
		"note": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{"type": "string"},
				},
			},
		},
	},
	"DiagnosticReport": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"registered", "partial", "preliminary", "final", "amended", "corrected", "appended", "cancelled", "entered-in-error", "unknown"},
		},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"code":      map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"subject":   map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"effectiveDateTime": map[string]interface{}{"type": "string", "format": "date-time"},
		"effectivePeriod":   map[string]interface{}{"$ref": "#/components/schemas/Period"},
		"issued":            map[string]interface{}{"type": "string", "format": "date-time"},
		"performer": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"result": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"conclusion":     map[string]interface{}{"type": "string"},
		"conclusionCode": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"ImagingStudy": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"registered", "available", "cancelled", "entered-in-error", "unknown"},
		},
		"subject":     map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter":   map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"started":     map[string]interface{}{"type": "string", "format": "date-time"},
		"referrer":    map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"numberOfSeries":    map[string]interface{}{"type": "integer"},
		"numberOfInstances": map[string]interface{}{"type": "integer"},
		"description":       map[string]interface{}{"type": "string"},
		"modality": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Coding"},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Specimen": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"available", "unavailable", "unsatisfactory", "entered-in-error"},
		},
		"type":    map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"subject": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"receivedTime": map[string]interface{}{"type": "string", "format": "date-time"},
		"collection": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collector":         map[string]interface{}{"$ref": "#/components/schemas/Reference"},
				"collectedDateTime": map[string]interface{}{"type": "string", "format": "date-time"},
				"quantity":          map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
				"bodySite":          map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
		"note": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{"type": "string"},
				},
			},
		},
	},
	"Appointment": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"proposed", "pending", "booked", "arrived", "fulfilled", "cancelled", "noshow", "entered-in-error", "checked-in", "waitlist"},
		},
		"serviceCategory": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"serviceType": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"appointmentType": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"reasonCode": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"priority":    map[string]interface{}{"type": "integer"},
		"description": map[string]interface{}{"type": "string"},
		"start":       map[string]interface{}{"type": "string", "format": "date-time"},
		"end":         map[string]interface{}{"type": "string", "format": "date-time"},
		"minutesDuration": map[string]interface{}{"type": "integer"},
		"participant": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					},
					"actor":    map[string]interface{}{"$ref": "#/components/schemas/Reference"},
					"required": map[string]interface{}{"type": "string", "enum": []string{"required", "optional", "information-only"}},
					"status":   map[string]interface{}{"type": "string", "enum": []string{"accepted", "declined", "tentative", "needs-action"}},
				},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Schedule": {
		"active":         map[string]interface{}{"type": "boolean"},
		"serviceCategory": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"serviceType": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"actor": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"planningHorizon": map[string]interface{}{"$ref": "#/components/schemas/Period"},
		"comment":         map[string]interface{}{"type": "string"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Slot": {
		"serviceCategory": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"serviceType": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"schedule": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"busy", "free", "busy-unavailable", "busy-tentative", "entered-in-error"},
		},
		"start":   map[string]interface{}{"type": "string", "format": "date-time"},
		"end":     map[string]interface{}{"type": "string", "format": "date-time"},
		"comment": map[string]interface{}{"type": "string"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Coverage": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"active", "cancelled", "draft", "entered-in-error"},
		},
		"type":         map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"subscriber":   map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"subscriberId": map[string]interface{}{"type": "string"},
		"beneficiary":  map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"dependent":    map[string]interface{}{"type": "string"},
		"relationship": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"period":       map[string]interface{}{"$ref": "#/components/schemas/Period"},
		"payor": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"order": map[string]interface{}{"type": "integer"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Claim": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"active", "cancelled", "draft", "entered-in-error"},
		},
		"type":     map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"use":      map[string]interface{}{"type": "string", "enum": []string{"claim", "preauthorization", "predetermination"}},
		"patient":  map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"created":  map[string]interface{}{"type": "string", "format": "date-time"},
		"provider": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"priority": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"insurance": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sequence": map[string]interface{}{"type": "integer"},
					"focal":    map[string]interface{}{"type": "boolean"},
					"coverage": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
				},
			},
		},
		"total": map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Consent": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"draft", "proposed", "active", "rejected", "inactive", "entered-in-error"},
		},
		"scope":    map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"patient":  map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"dateTime": map[string]interface{}{"type": "string", "format": "date-time"},
		"performer": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"organization": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"policyRule": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"provision": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"type":   map[string]interface{}{"type": "string", "enum": []string{"deny", "permit"}},
				"period": map[string]interface{}{"$ref": "#/components/schemas/Period"},
				"action": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
				},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"DocumentReference": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"current", "superseded", "entered-in-error"},
		},
		"docStatus": map[string]interface{}{
			"type": "string",
			"enum": []string{"preliminary", "final", "amended", "entered-in-error"},
		},
		"type":    map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"subject":    map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"date":       map[string]interface{}{"type": "string", "format": "date-time"},
		"author": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"custodian":   map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"description": map[string]interface{}{"type": "string"},
		"content": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"attachment": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"contentType": map[string]interface{}{"type": "string"},
							"url":         map[string]interface{}{"type": "string", "format": "uri"},
							"title":       map[string]interface{}{"type": "string"},
							"creation":    map[string]interface{}{"type": "string", "format": "date-time"},
						},
					},
					"format": map[string]interface{}{"$ref": "#/components/schemas/Coding"},
				},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Composition": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"preliminary", "final", "amended", "entered-in-error"},
		},
		"type":    map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"subject": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"date":    map[string]interface{}{"type": "string", "format": "date-time"},
		"author": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"title":        map[string]interface{}{"type": "string"},
		"confidentiality": map[string]interface{}{"type": "string"},
		"custodian":    map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"section": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{"type": "string"},
					"code":  map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"text": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"status": map[string]interface{}{"type": "string"},
							"div":    map[string]interface{}{"type": "string"},
						},
					},
					"entry": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
					},
				},
			},
		},
		"identifier": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
	},
	"Communication": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"preparation", "in-progress", "not-done", "on-hold", "stopped", "completed", "entered-in-error", "unknown"},
		},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"priority": map[string]interface{}{
			"type": "string",
			"enum": []string{"routine", "urgent", "asap", "stat"},
		},
		"subject":   map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"sent":      map[string]interface{}{"type": "string", "format": "date-time"},
		"received":  map[string]interface{}{"type": "string", "format": "date-time"},
		"sender":    map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"recipient": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"payload": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"contentString":    map[string]interface{}{"type": "string"},
					"contentReference": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
				},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"ResearchStudy": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"active", "administratively-completed", "approved", "closed-to-accrual", "closed-to-accrual-and-intervention", "completed", "disapproved", "in-review", "temporarily-closed-to-accrual", "temporarily-closed-to-accrual-and-intervention", "withdrawn"},
		},
		"title":       map[string]interface{}{"type": "string"},
		"description": map[string]interface{}{"type": "string"},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"focus": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"condition": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"period":          map[string]interface{}{"$ref": "#/components/schemas/Period"},
		"sponsor":         map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"principalInvestigator": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Questionnaire": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"draft", "active", "retired", "unknown"},
		},
		"name":          map[string]interface{}{"type": "string"},
		"title":         map[string]interface{}{"type": "string"},
		"date":          map[string]interface{}{"type": "string", "format": "date-time"},
		"publisher":     map[string]interface{}{"type": "string"},
		"description":   map[string]interface{}{"type": "string"},
		"subjectType": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"type": "string"},
		},
		"item": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"linkId":   map[string]interface{}{"type": "string"},
					"text":     map[string]interface{}{"type": "string"},
					"type":     map[string]interface{}{"type": "string"},
					"required": map[string]interface{}{"type": "boolean"},
					"repeats":  map[string]interface{}{"type": "boolean"},
				},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"QuestionnaireResponse": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"in-progress", "completed", "amended", "entered-in-error", "stopped"},
		},
		"questionnaire": map[string]interface{}{"type": "string"},
		"subject":       map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter":     map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"authored":      map[string]interface{}{"type": "string", "format": "date-time"},
		"author":        map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"source":        map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"item": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"linkId": map[string]interface{}{"type": "string"},
					"text":   map[string]interface{}{"type": "string"},
					"answer": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"valueString":  map[string]interface{}{"type": "string"},
								"valueBoolean": map[string]interface{}{"type": "boolean"},
								"valueInteger": map[string]interface{}{"type": "integer"},
								"valueDate":    map[string]interface{}{"type": "string", "format": "date"},
								"valueCoding":  map[string]interface{}{"$ref": "#/components/schemas/Coding"},
							},
						},
					},
				},
			},
		},
		"identifier": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
	},
	"Immunization": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"completed", "entered-in-error", "not-done"},
		},
		"vaccineCode":    map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"patient":        map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter":      map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"occurrenceDateTime": map[string]interface{}{"type": "string", "format": "date-time"},
		"occurrenceString":   map[string]interface{}{"type": "string"},
		"recorded":          map[string]interface{}{"type": "string", "format": "date-time"},
		"primarySource":     map[string]interface{}{"type": "boolean"},
		"manufacturer":      map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"lotNumber":         map[string]interface{}{"type": "string"},
		"expirationDate":    map[string]interface{}{"type": "string", "format": "date"},
		"site":              map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"route":             map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"doseQuantity":      map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
		"performer": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"function": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"actor":    map[string]interface{}{"$ref": "#/components/schemas/Reference"},
				},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"ImmunizationRecommendation": {
		"patient": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"date":    map[string]interface{}{"type": "string", "format": "date-time"},
		"recommendation": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"vaccineCode": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					},
					"forecastStatus":  map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"forecastReason": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					},
					"dateCriterion": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"code":  map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
								"value": map[string]interface{}{"type": "string", "format": "date-time"},
							},
						},
					},
					"doseNumberPositiveInt": map[string]interface{}{"type": "integer"},
					"seriesDosesPositiveInt": map[string]interface{}{"type": "integer"},
				},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"CarePlan": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"draft", "active", "on-hold", "revoked", "completed", "entered-in-error", "unknown"},
		},
		"intent": map[string]interface{}{
			"type": "string",
			"enum": []string{"proposal", "plan", "order", "option"},
		},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"title":       map[string]interface{}{"type": "string"},
		"description": map[string]interface{}{"type": "string"},
		"subject":     map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"encounter":   map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"period":      map[string]interface{}{"$ref": "#/components/schemas/Period"},
		"created":     map[string]interface{}{"type": "string", "format": "date-time"},
		"author":      map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"careTeam": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"goal": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"activity": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"reference": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
					"detail": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"kind":        map[string]interface{}{"type": "string"},
							"code":        map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
							"status":      map[string]interface{}{"type": "string"},
							"description": map[string]interface{}{"type": "string"},
						},
					},
				},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Goal": {
		"lifecycleStatus": map[string]interface{}{
			"type": "string",
			"enum": []string{"proposed", "planned", "accepted", "active", "on-hold", "completed", "cancelled", "entered-in-error", "rejected"},
		},
		"achievementStatus": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"category": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"priority":    map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"description": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"subject":     map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"startDate":   map[string]interface{}{"type": "string", "format": "date"},
		"target": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"measure":        map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"detailQuantity": map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
					"detailString":   map[string]interface{}{"type": "string"},
					"dueDate":        map[string]interface{}{"type": "string", "format": "date"},
				},
			},
		},
		"statusDate":   map[string]interface{}{"type": "string", "format": "date"},
		"statusReason": map[string]interface{}{"type": "string"},
		"expressedBy":  map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
		"note": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{"type": "string"},
				},
			},
		},
	},
	"FamilyMemberHistory": {
		"status": map[string]interface{}{
			"type": "string",
			"enum": []string{"partial", "completed", "entered-in-error", "health-unknown"},
		},
		"patient":      map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"date":         map[string]interface{}{"type": "string", "format": "date-time"},
		"name":         map[string]interface{}{"type": "string"},
		"relationship": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"sex":          map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"bornDate":     map[string]interface{}{"type": "string", "format": "date"},
		"deceasedBoolean":  map[string]interface{}{"type": "boolean"},
		"deceasedDate":     map[string]interface{}{"type": "string", "format": "date"},
		"condition": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code":        map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"outcome":     map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"onsetAge":    map[string]interface{}{"$ref": "#/components/schemas/Quantity"},
					"onsetString": map[string]interface{}{"type": "string"},
				},
			},
		},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
		"note": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{"type": "string"},
				},
			},
		},
	},
	"RelatedPerson": {
		"active":  map[string]interface{}{"type": "boolean"},
		"patient": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"relationship": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"name": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/HumanName"},
		},
		"telecom": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/ContactPoint"},
		},
		"gender":    map[string]interface{}{"type": "string", "enum": []string{"male", "female", "other", "unknown"}},
		"birthDate": map[string]interface{}{"type": "string", "format": "date"},
		"address": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Address"},
		},
		"period": map[string]interface{}{"$ref": "#/components/schemas/Period"},
		"identifier": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Identifier"},
		},
	},
	"Provenance": {
		"target": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		},
		"occurredDateTime": map[string]interface{}{"type": "string", "format": "date-time"},
		"occurredPeriod":   map[string]interface{}{"$ref": "#/components/schemas/Period"},
		"recorded":         map[string]interface{}{"type": "string", "format": "date-time"},
		"policy": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"type": "string", "format": "uri"},
		},
		"location": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
		"reason": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		},
		"activity": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
		"agent": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					"role": map[string]interface{}{
						"type":  "array",
						"items": map[string]interface{}{"$ref": "#/components/schemas/CodeableConcept"},
					},
					"who":        map[string]interface{}{"$ref": "#/components/schemas/Reference"},
					"onBehalfOf": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
				},
			},
		},
		"entity": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"role": map[string]interface{}{
						"type": "string",
						"enum": []string{"derivation", "revision", "quotation", "source", "removal"},
					},
					"what": map[string]interface{}{"$ref": "#/components/schemas/Reference"},
				},
			},
		},
	},
}

// buildResourceSchema builds the OpenAPI schema for a FHIR resource type.
// If a detailed definition exists, it merges it with the base properties.
// Otherwise, it returns a schema with just resourceType, id, and meta.
func buildResourceSchema(resourceType string) map[string]interface{} {
	props := baseResourceProperties(resourceType)

	if extra, ok := resourceSchemaDefinitions[resourceType]; ok {
		mergeProperties(props, extra)
	}

	return map[string]interface{}{
		"type":       "object",
		"properties": props,
	}
}

// ── Swagger UI ──────────────────────────────────────────────────────────

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Headless EHR FHIR R4 API - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" >
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "/api/openapi.json",
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIBundle.SwaggerUIStandalonePreset
      ],
      layout: "BaseLayout"
    })
  </script>
</body>
</html>`

// RegisterRoutes registers the OpenAPI endpoints.
func (g *Generator) RegisterRoutes(apiGroup *echo.Group) {
	apiGroup.GET("/openapi.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, g.GenerateSpec())
	})
	apiGroup.GET("/docs", func(c echo.Context) error {
		return c.HTML(http.StatusOK, swaggerUIHTML)
	})
}
