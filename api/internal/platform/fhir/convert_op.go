package fhir

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ConvertParams holds the conversion parameters for the $convert operation.
type ConvertParams struct {
	InputFormat  string // "json" or "xml"
	OutputFormat string // "json" or "xml"
}

// parseConvertParams extracts and normalises the _inputFormat and _outputFormat
// query parameters. Missing values default to "json".
func parseConvertParams(c echo.Context) ConvertParams {
	input := c.QueryParam("_inputFormat")
	output := c.QueryParam("_outputFormat")

	if input == "" {
		input = "json"
	}
	if output == "" {
		output = "json"
	}

	// Normalise common MIME types to short names.
	input = normalizeConvertFormat(input)
	output = normalizeConvertFormat(output)

	return ConvertParams{
		InputFormat:  input,
		OutputFormat: output,
	}
}

// normalizeConvertFormat maps MIME-type strings to their short names.
func normalizeConvertFormat(f string) string {
	switch normalizeFormat(f) {
	case "json", "application/json", "application/fhir+json":
		return "json"
	case "xml", "application/xml", "application/fhir+xml":
		return "xml"
	}
	return f
}

// ConvertHandler creates a handler for POST /fhir/$convert.
//
// The handler accepts a FHIR resource in the request body and converts between
// formats based on _inputFormat and _outputFormat query parameters.
//
// Because this server only supports JSON:
//   - If outputFormat is "xml", return 406 Not Acceptable.
//   - If inputFormat is "xml", return 415 Unsupported Media Type.
//   - For JSON->JSON, validate the JSON and return the resource (useful for
//     format normalisation / pretty-printing).
func ConvertHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		params := parseConvertParams(c)

		// Reject XML output requests.
		if params.OutputFormat == "xml" {
			return c.JSON(http.StatusNotAcceptable,
				operationOutcome("error", "not-supported",
					"XML output is not supported; this server only supports application/fhir+json"))
		}

		// Reject XML input.
		if params.InputFormat == "xml" {
			return c.JSON(http.StatusUnsupportedMediaType,
				operationOutcome("error", "not-supported",
					"XML input is not supported; this server only supports application/fhir+json"))
		}

		// Reject unknown formats.
		if params.InputFormat != "json" {
			return c.JSON(http.StatusUnsupportedMediaType,
				operationOutcome("error", "not-supported",
					"Unsupported input format: "+params.InputFormat))
		}
		if params.OutputFormat != "json" {
			return c.JSON(http.StatusNotAcceptable,
				operationOutcome("error", "not-supported",
					"Unsupported output format: "+params.OutputFormat))
		}

		// Read the request body.
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest,
				operationOutcome("error", "structure", "Failed to read request body"))
		}
		if len(body) == 0 {
			return c.JSON(http.StatusBadRequest,
				operationOutcome("error", "required", "Request body is empty"))
		}

		// Validate the JSON by parsing it.
		var resource map[string]interface{}
		if err := json.Unmarshal(body, &resource); err != nil {
			return c.JSON(http.StatusBadRequest,
				operationOutcome("error", "structure", "Invalid JSON: "+err.Error()))
		}

		// Validate that resourceType is present.
		if _, ok := resource["resourceType"]; !ok {
			return c.JSON(http.StatusBadRequest,
				operationOutcome("error", "structure", "Resource must contain a resourceType field"))
		}

		// Return the normalised JSON resource with the FHIR content type.
		c.Response().Header().Set(echo.HeaderContentType, FHIRContentType)
		return c.JSON(http.StatusOK, resource)
	}
}
