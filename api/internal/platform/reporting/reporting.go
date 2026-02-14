package reporting

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
)

// MeasureDefinition defines a reporting measure with its SQL query.
type MeasureDefinition struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	SQL         string   `json:"sql"`
	Parameters  []string `json:"parameters"`
}

// MeasureReport holds the results of evaluating a measure.
type MeasureReport struct {
	MeasureID   string                   `json:"measure_id"`
	MeasureName string                   `json:"measure_name"`
	GeneratedAt time.Time                `json:"generated_at"`
	Results     []map[string]interface{} `json:"results"`
	Parameters  map[string]string        `json:"parameters,omitempty"`
}

// PredefinedMeasures is the list of available reporting measures.
var PredefinedMeasures = []MeasureDefinition{
	{
		ID:          "patient-count",
		Name:        "Patient Count",
		Description: "Total number of patients in the system, optionally filtered by active status",
		SQL:         `SELECT COUNT(*) AS total, COALESCE(SUM(CASE WHEN active THEN 1 ELSE 0 END), 0) AS active_count FROM patient`,
		Parameters:  []string{},
	},
	{
		ID:          "encounter-volume-by-type",
		Name:        "Encounter Volume by Type",
		Description: "Number of encounters grouped by class/type",
		SQL:         `SELECT COALESCE(class, 'unknown') AS encounter_class, COUNT(*) AS total FROM encounter GROUP BY class ORDER BY total DESC`,
		Parameters:  []string{},
	},
	{
		ID:          "active-medication-orders",
		Name:        "Active Medication Orders",
		Description: "Count of medication requests by status",
		SQL:         `SELECT status, COUNT(*) AS total FROM medication_request GROUP BY status ORDER BY total DESC`,
		Parameters:  []string{},
	},
	{
		ID:          "diagnostic-report-summary",
		Name:        "Diagnostic Report Summary",
		Description: "Count of diagnostic reports by status",
		SQL:         `SELECT status, COUNT(*) AS total FROM diagnostic_report GROUP BY status ORDER BY total DESC`,
		Parameters:  []string{},
	},
}

// Handler provides HTTP handlers for the reporting API.
type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler creates a new reporting handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// RegisterRoutes registers the reporting API routes.
func (h *Handler) RegisterRoutes(api *echo.Group) {
	reportGroup := api.Group("/reports", auth.RequireRole("admin", "physician"))
	reportGroup.GET("/measures", h.ListMeasures)
	reportGroup.GET("/measures/:id/evaluate", h.EvaluateMeasure)
}

// ListMeasures returns all available measure definitions.
func (h *Handler) ListMeasures(c echo.Context) error {
	return c.JSON(http.StatusOK, PredefinedMeasures)
}

// EvaluateMeasure executes a measure's SQL and returns the results.
func (h *Handler) EvaluateMeasure(c echo.Context) error {
	measureID := c.Param("id")

	var measure *MeasureDefinition
	for i := range PredefinedMeasures {
		if PredefinedMeasures[i].ID == measureID {
			measure = &PredefinedMeasures[i]
			break
		}
	}
	if measure == nil {
		return echo.NewHTTPError(http.StatusNotFound, "measure not found")
	}

	// Collect parameters from query string
	params := map[string]string{}
	for _, p := range measure.Parameters {
		if v := c.QueryParam(p); v != "" {
			params[p] = v
		}
	}

	results, err := h.executeSQL(c.Request().Context(), measure.SQL)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("query failed: %v", err))
	}

	report := MeasureReport{
		MeasureID:   measure.ID,
		MeasureName: measure.Name,
		GeneratedAt: time.Now(),
		Results:     results,
		Parameters:  params,
	}

	return c.JSON(http.StatusOK, report)
}

// executeSQL runs a SQL query and returns results as a slice of maps.
func (h *Handler) executeSQL(ctx context.Context, sql string) ([]map[string]interface{}, error) {
	rows, err := h.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	var results []map[string]interface{}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}

		row := make(map[string]interface{}, len(fieldDescs))
		for i, fd := range fieldDescs {
			row[string(fd.Name)] = values[i]
		}
		results = append(results, row)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	return results, nil
}

// FindMeasure looks up a measure by ID, useful for testing.
func FindMeasure(id string) *MeasureDefinition {
	for i := range PredefinedMeasures {
		if PredefinedMeasures[i].ID == id {
			return &PredefinedMeasures[i]
		}
	}
	return nil
}

// ensure uuid is used (for potential parameterized queries in the future)
var _ = uuid.New
