package fhir

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// StatsParams holds parameters for the $stats operation.
type StatsParams struct {
	Patient   string   // Patient reference
	Code      string   // Observation code (system|code)
	System    string   // Code system (optional)
	Period    string   // Date range (FHIR date format)
	Statistic []string // Requested statistics: count, min, max, mean, median, stddev, sum, 5-num
}

// StatsResult holds computed statistics.
type StatsResult struct {
	Code    string   `json:"code"`
	Subject string   `json:"subject"`
	Period  string   `json:"period,omitempty"`
	Count   int      `json:"count"`
	Min     *float64 `json:"min,omitempty"`
	Max     *float64 `json:"max,omitempty"`
	Mean    *float64 `json:"mean,omitempty"`
	Median  *float64 `json:"median,omitempty"`
	StdDev  *float64 `json:"stddev,omitempty"`
	Sum     *float64 `json:"sum,omitempty"`
}

// ParseStatsParams extracts $stats parameters from the request.
func ParseStatsParams(c echo.Context) StatsParams {
	params := StatsParams{
		Patient: c.QueryParam("patient"),
		Code:    c.QueryParam("code"),
		System:  c.QueryParam("system"),
		Period:  c.QueryParam("period"),
	}

	if statStr := c.QueryParam("statistic"); statStr != "" {
		params.Statistic = strings.Split(statStr, ",")
	}

	return params
}

// StatsExecutor is a function that executes the actual $stats computation.
type StatsExecutor func(ctx context.Context, params StatsParams) (*StatsResult, error)

// StatsHandler creates a handler for GET/POST /fhir/Observation/$stats.
// It accepts a function that executes the actual computation.
func StatsHandler(executor StatsExecutor) echo.HandlerFunc {
	return func(c echo.Context) error {
		params := ParseStatsParams(c)

		if params.Patient == "" {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeRequired, "patient parameter is required for $stats",
			))
		}

		if params.Code == "" {
			return c.JSON(http.StatusBadRequest, NewOperationOutcome(
				IssueSeverityError, IssueTypeRequired, "code parameter is required for $stats",
			))
		}

		result, err := executor(c.Request().Context(), params)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, NewOperationOutcome(
				IssueSeverityError, IssueTypeProcessing, "stats operation failed: "+err.Error(),
			))
		}

		parameters := buildStatsParameters(result)

		return c.JSON(http.StatusOK, parameters)
	}
}

// buildStatsParameters creates a FHIR Parameters resource from the stats result.
func buildStatsParameters(result *StatsResult) map[string]interface{} {
	params := []interface{}{
		map[string]interface{}{
			"name":        "code",
			"valueString": result.Code,
		},
		map[string]interface{}{
			"name":        "subject",
			"valueString": result.Subject,
		},
		map[string]interface{}{
			"name":         "count",
			"valueInteger": result.Count,
		},
	}

	if result.Period != "" {
		params = append(params, map[string]interface{}{
			"name":        "period",
			"valueString": result.Period,
		})
	}

	if result.Min != nil {
		params = append(params, map[string]interface{}{
			"name":         "min",
			"valueDecimal": *result.Min,
		})
	}

	if result.Max != nil {
		params = append(params, map[string]interface{}{
			"name":         "max",
			"valueDecimal": *result.Max,
		})
	}

	if result.Mean != nil {
		params = append(params, map[string]interface{}{
			"name":         "mean",
			"valueDecimal": *result.Mean,
		})
	}

	if result.Median != nil {
		params = append(params, map[string]interface{}{
			"name":         "median",
			"valueDecimal": *result.Median,
		})
	}

	if result.StdDev != nil {
		params = append(params, map[string]interface{}{
			"name":         "stddev",
			"valueDecimal": *result.StdDev,
		})
	}

	if result.Sum != nil {
		params = append(params, map[string]interface{}{
			"name":         "sum",
			"valueDecimal": *result.Sum,
		})
	}

	return map[string]interface{}{
		"resourceType": "Parameters",
		"parameter":    params,
	}
}
