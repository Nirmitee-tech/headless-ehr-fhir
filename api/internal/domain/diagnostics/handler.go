package diagnostics

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/pkg/pagination"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(api *echo.Group, fhirGroup *echo.Group) {
	// Read endpoints – admin, physician, nurse, lab_tech, radiologist
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "lab_tech", "radiologist"))
	readGroup.GET("/service-requests", h.ListServiceRequests)
	readGroup.GET("/service-requests/:id", h.GetServiceRequest)
	readGroup.GET("/specimens", h.ListSpecimens)
	readGroup.GET("/specimens/:id", h.GetSpecimen)
	readGroup.GET("/diagnostic-reports", h.ListDiagnosticReports)
	readGroup.GET("/diagnostic-reports/:id", h.GetDiagnosticReport)
	readGroup.GET("/diagnostic-reports/:id/results", h.GetResults)
	readGroup.GET("/service-requests/:id/status-history", h.GetServiceRequestStatusHistory)
	readGroup.GET("/imaging-studies", h.ListImagingStudies)
	readGroup.GET("/imaging-studies/:id", h.GetImagingStudy)

	// Write endpoints – admin, physician, lab_tech, radiologist
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "lab_tech", "radiologist"))
	writeGroup.POST("/service-requests", h.CreateServiceRequest)
	writeGroup.PUT("/service-requests/:id", h.UpdateServiceRequest)
	writeGroup.DELETE("/service-requests/:id", h.DeleteServiceRequest)
	writeGroup.POST("/specimens", h.CreateSpecimen)
	writeGroup.PUT("/specimens/:id", h.UpdateSpecimen)
	writeGroup.DELETE("/specimens/:id", h.DeleteSpecimen)
	writeGroup.POST("/diagnostic-reports", h.CreateDiagnosticReport)
	writeGroup.PUT("/diagnostic-reports/:id", h.UpdateDiagnosticReport)
	writeGroup.DELETE("/diagnostic-reports/:id", h.DeleteDiagnosticReport)
	writeGroup.POST("/diagnostic-reports/:id/results", h.AddResult)
	writeGroup.DELETE("/diagnostic-reports/:id/results/:observationId", h.RemoveResult)
	writeGroup.POST("/imaging-studies", h.CreateImagingStudy)
	writeGroup.PUT("/imaging-studies/:id", h.UpdateImagingStudy)
	writeGroup.DELETE("/imaging-studies/:id", h.DeleteImagingStudy)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse", "lab_tech", "radiologist"))
	fhirRead.GET("/ServiceRequest", h.SearchServiceRequestsFHIR)
	fhirRead.GET("/ServiceRequest/:id", h.GetServiceRequestFHIR)
	fhirRead.GET("/DiagnosticReport", h.SearchDiagnosticReportsFHIR)
	fhirRead.GET("/DiagnosticReport/:id", h.GetDiagnosticReportFHIR)
	fhirRead.GET("/Specimen", h.SearchSpecimensFHIR)
	fhirRead.GET("/Specimen/:id", h.GetSpecimenFHIR)
	fhirRead.GET("/ImagingStudy", h.SearchImagingStudiesFHIR)
	fhirRead.GET("/ImagingStudy/:id", h.GetImagingStudyFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin", "physician", "lab_tech", "radiologist"))
	fhirWrite.POST("/ServiceRequest", h.CreateServiceRequestFHIR)
	fhirWrite.PUT("/ServiceRequest/:id", h.UpdateServiceRequestFHIR)
	fhirWrite.DELETE("/ServiceRequest/:id", h.DeleteServiceRequestFHIR)
	fhirWrite.PATCH("/ServiceRequest/:id", h.PatchServiceRequestFHIR)
	fhirWrite.POST("/DiagnosticReport", h.CreateDiagnosticReportFHIR)
	fhirWrite.PUT("/DiagnosticReport/:id", h.UpdateDiagnosticReportFHIR)
	fhirWrite.DELETE("/DiagnosticReport/:id", h.DeleteDiagnosticReportFHIR)
	fhirWrite.PATCH("/DiagnosticReport/:id", h.PatchDiagnosticReportFHIR)
	fhirWrite.POST("/Specimen", h.CreateSpecimenFHIR)
	fhirWrite.PUT("/Specimen/:id", h.UpdateSpecimenFHIR)
	fhirWrite.DELETE("/Specimen/:id", h.DeleteSpecimenFHIR)
	fhirWrite.PATCH("/Specimen/:id", h.PatchSpecimenFHIR)
	fhirWrite.POST("/ImagingStudy", h.CreateImagingStudyFHIR)
	fhirWrite.PUT("/ImagingStudy/:id", h.UpdateImagingStudyFHIR)
	fhirWrite.DELETE("/ImagingStudy/:id", h.DeleteImagingStudyFHIR)
	fhirWrite.PATCH("/ImagingStudy/:id", h.PatchImagingStudyFHIR)

	// FHIR POST _search endpoints
	fhirRead.POST("/ServiceRequest/_search", h.SearchServiceRequestsFHIR)
	fhirRead.POST("/DiagnosticReport/_search", h.SearchDiagnosticReportsFHIR)
	fhirRead.POST("/Specimen/_search", h.SearchSpecimensFHIR)
	fhirRead.POST("/ImagingStudy/_search", h.SearchImagingStudiesFHIR)

	// FHIR vread and history endpoints
	fhirRead.GET("/ServiceRequest/:id/_history/:vid", h.VreadServiceRequestFHIR)
	fhirRead.GET("/ServiceRequest/:id/_history", h.HistoryServiceRequestFHIR)
	fhirRead.GET("/DiagnosticReport/:id/_history/:vid", h.VreadDiagnosticReportFHIR)
	fhirRead.GET("/DiagnosticReport/:id/_history", h.HistoryDiagnosticReportFHIR)
	fhirRead.GET("/ImagingStudy/:id/_history/:vid", h.VreadImagingStudyFHIR)
	fhirRead.GET("/ImagingStudy/:id/_history", h.HistoryImagingStudyFHIR)
	fhirRead.GET("/Specimen/:id/_history/:vid", h.VreadSpecimenFHIR)
	fhirRead.GET("/Specimen/:id/_history", h.HistorySpecimenFHIR)
}

// -- ServiceRequest Handlers --

func (h *Handler) CreateServiceRequest(c echo.Context) error {
	var sr ServiceRequest
	if err := c.Bind(&sr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateServiceRequest(c.Request().Context(), &sr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sr)
}

func (h *Handler) GetServiceRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sr, err := h.svc.GetServiceRequest(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "service request not found")
	}
	return c.JSON(http.StatusOK, sr)
}

func (h *Handler) ListServiceRequests(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListServiceRequestsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchServiceRequests(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateServiceRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sr ServiceRequest
	if err := c.Bind(&sr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sr.ID = id
	if err := h.svc.UpdateServiceRequest(c.Request().Context(), &sr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sr)
}

func (h *Handler) DeleteServiceRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteServiceRequest(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Status History Handlers --

func (h *Handler) GetServiceRequestStatusHistory(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	history, err := h.svc.GetStatusHistory(c.Request().Context(), "ServiceRequest", id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, history)
}

// -- Specimen Handlers --

func (h *Handler) CreateSpecimen(c echo.Context) error {
	var sp Specimen
	if err := c.Bind(&sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateSpecimen(c.Request().Context(), &sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, sp)
}

func (h *Handler) GetSpecimen(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	sp, err := h.svc.GetSpecimen(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "specimen not found")
	}
	return c.JSON(http.StatusOK, sp)
}

func (h *Handler) ListSpecimens(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListSpecimensByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchSpecimens(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateSpecimen(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var sp Specimen
	if err := c.Bind(&sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sp.ID = id
	if err := h.svc.UpdateSpecimen(c.Request().Context(), &sp); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sp)
}

func (h *Handler) DeleteSpecimen(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteSpecimen(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- DiagnosticReport Handlers --

func (h *Handler) CreateDiagnosticReport(c echo.Context) error {
	var dr DiagnosticReport
	if err := c.Bind(&dr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateDiagnosticReport(c.Request().Context(), &dr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, dr)
}

func (h *Handler) GetDiagnosticReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	dr, err := h.svc.GetDiagnosticReport(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "diagnostic report not found")
	}
	return c.JSON(http.StatusOK, dr)
}

func (h *Handler) ListDiagnosticReports(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListDiagnosticReportsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchDiagnosticReports(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateDiagnosticReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var dr DiagnosticReport
	if err := c.Bind(&dr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	dr.ID = id
	if err := h.svc.UpdateDiagnosticReport(c.Request().Context(), &dr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, dr)
}

func (h *Handler) DeleteDiagnosticReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteDiagnosticReport(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddResult(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var body struct {
		ObservationID uuid.UUID `json:"observation_id"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.AddDiagnosticReportResult(c.Request().Context(), id, body.ObservationID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"diagnostic_report_id": id,
		"observation_id":       body.ObservationID,
	})
}

func (h *Handler) GetResults(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ids, err := h.svc.GetDiagnosticReportResults(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, ids)
}

func (h *Handler) RemoveResult(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	obsID, err := uuid.Parse(c.Param("observationId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid observation id")
	}
	if err := h.svc.RemoveDiagnosticReportResult(c.Request().Context(), id, obsID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- ImagingStudy Handlers --

func (h *Handler) CreateImagingStudy(c echo.Context) error {
	var is ImagingStudy
	if err := c.Bind(&is); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateImagingStudy(c.Request().Context(), &is); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, is)
}

func (h *Handler) GetImagingStudy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	is, err := h.svc.GetImagingStudy(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "imaging study not found")
	}
	return c.JSON(http.StatusOK, is)
}

func (h *Handler) ListImagingStudies(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListImagingStudiesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchImagingStudies(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateImagingStudy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var is ImagingStudy
	if err := c.Bind(&is); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	is.ID = id
	if err := h.svc.UpdateImagingStudy(c.Request().Context(), &is); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, is)
}

func (h *Handler) DeleteImagingStudy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteImagingStudy(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Endpoints --

func (h *Handler) SearchServiceRequestsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "category", "code", "intent"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchServiceRequests(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ServiceRequest"))
}

func (h *Handler) GetServiceRequestFHIR(c echo.Context) error {
	sr, err := h.svc.GetServiceRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ServiceRequest", c.Param("id")))
	}
	return c.JSON(http.StatusOK, sr.ToFHIR())
}

func (h *Handler) CreateServiceRequestFHIR(c echo.Context) error {
	var sr ServiceRequest
	if err := c.Bind(&sr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateServiceRequest(c.Request().Context(), &sr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ServiceRequest/"+sr.FHIRID)
	return c.JSON(http.StatusCreated, sr.ToFHIR())
}

func (h *Handler) SearchDiagnosticReportsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "category", "code"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchDiagnosticReports(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/DiagnosticReport"))
}

func (h *Handler) GetDiagnosticReportFHIR(c echo.Context) error {
	dr, err := h.svc.GetDiagnosticReportByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DiagnosticReport", c.Param("id")))
	}
	return c.JSON(http.StatusOK, dr.ToFHIR())
}

func (h *Handler) CreateDiagnosticReportFHIR(c echo.Context) error {
	var dr DiagnosticReport
	if err := c.Bind(&dr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateDiagnosticReport(c.Request().Context(), &dr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/DiagnosticReport/"+dr.FHIRID)
	return c.JSON(http.StatusCreated, dr.ToFHIR())
}

func (h *Handler) SearchSpecimensFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "type"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchSpecimens(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Specimen"))
}

func (h *Handler) GetSpecimenFHIR(c echo.Context) error {
	sp, err := h.svc.GetSpecimenByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Specimen", c.Param("id")))
	}
	return c.JSON(http.StatusOK, sp.ToFHIR())
}

func (h *Handler) CreateSpecimenFHIR(c echo.Context) error {
	var sp Specimen
	if err := c.Bind(&sp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateSpecimen(c.Request().Context(), &sp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Specimen/"+sp.FHIRID)
	return c.JSON(http.StatusCreated, sp.ToFHIR())
}

func (h *Handler) SearchImagingStudiesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "modality"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchImagingStudies(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ImagingStudy"))
}

func (h *Handler) GetImagingStudyFHIR(c echo.Context) error {
	is, err := h.svc.GetImagingStudyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImagingStudy", c.Param("id")))
	}
	return c.JSON(http.StatusOK, is.ToFHIR())
}

func (h *Handler) CreateImagingStudyFHIR(c echo.Context) error {
	var is ImagingStudy
	if err := c.Bind(&is); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateImagingStudy(c.Request().Context(), &is); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ImagingStudy/"+is.FHIRID)
	return c.JSON(http.StatusCreated, is.ToFHIR())
}

// -- FHIR Update Endpoints --

func (h *Handler) UpdateServiceRequestFHIR(c echo.Context) error {
	var sr ServiceRequest
	if err := c.Bind(&sr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetServiceRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ServiceRequest", c.Param("id")))
	}
	sr.ID = existing.ID
	sr.FHIRID = existing.FHIRID
	if err := h.svc.UpdateServiceRequest(c.Request().Context(), &sr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, sr.ToFHIR())
}

func (h *Handler) UpdateDiagnosticReportFHIR(c echo.Context) error {
	var dr DiagnosticReport
	if err := c.Bind(&dr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetDiagnosticReportByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DiagnosticReport", c.Param("id")))
	}
	dr.ID = existing.ID
	dr.FHIRID = existing.FHIRID
	if err := h.svc.UpdateDiagnosticReport(c.Request().Context(), &dr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, dr.ToFHIR())
}

func (h *Handler) UpdateSpecimenFHIR(c echo.Context) error {
	var sp Specimen
	if err := c.Bind(&sp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetSpecimenByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Specimen", c.Param("id")))
	}
	sp.ID = existing.ID
	sp.FHIRID = existing.FHIRID
	if err := h.svc.UpdateSpecimen(c.Request().Context(), &sp); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, sp.ToFHIR())
}

func (h *Handler) UpdateImagingStudyFHIR(c echo.Context) error {
	var is ImagingStudy
	if err := c.Bind(&is); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetImagingStudyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImagingStudy", c.Param("id")))
	}
	is.ID = existing.ID
	is.FHIRID = existing.FHIRID
	if err := h.svc.UpdateImagingStudy(c.Request().Context(), &is); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, is.ToFHIR())
}

// -- FHIR Delete Endpoints --

func (h *Handler) DeleteServiceRequestFHIR(c echo.Context) error {
	existing, err := h.svc.GetServiceRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ServiceRequest", c.Param("id")))
	}
	if err := h.svc.DeleteServiceRequest(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteDiagnosticReportFHIR(c echo.Context) error {
	existing, err := h.svc.GetDiagnosticReportByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DiagnosticReport", c.Param("id")))
	}
	if err := h.svc.DeleteDiagnosticReport(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteSpecimenFHIR(c echo.Context) error {
	existing, err := h.svc.GetSpecimenByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Specimen", c.Param("id")))
	}
	if err := h.svc.DeleteSpecimen(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteImagingStudyFHIR(c echo.Context) error {
	existing, err := h.svc.GetImagingStudyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImagingStudy", c.Param("id")))
	}
	if err := h.svc.DeleteImagingStudy(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR PATCH Endpoints --

func (h *Handler) PatchServiceRequestFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetServiceRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ServiceRequest", c.Param("id")))
	}
	currentResource := existing.ToFHIR()
	var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") {
		ops, err := fhir.ParseJSONPatch(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyJSONPatch(currentResource, ops)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else if strings.Contains(contentType, "merge-patch+json") {
		var mp map[string]interface{}
		if err := json.Unmarshal(body, &mp); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mp)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}
	_ = patched
	if v, ok := patched["status"].(string); ok {
		existing.Status = v
	}
	if err := h.svc.UpdateServiceRequest(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

func (h *Handler) PatchDiagnosticReportFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetDiagnosticReportByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DiagnosticReport", c.Param("id")))
	}
	currentResource := existing.ToFHIR()
	var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") {
		ops, err := fhir.ParseJSONPatch(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyJSONPatch(currentResource, ops)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else if strings.Contains(contentType, "merge-patch+json") {
		var mp map[string]interface{}
		if err := json.Unmarshal(body, &mp); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mp)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}
	_ = patched
	if v, ok := patched["status"].(string); ok {
		existing.Status = v
	}
	if err := h.svc.UpdateDiagnosticReport(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

func (h *Handler) PatchSpecimenFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetSpecimenByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Specimen", c.Param("id")))
	}
	currentResource := existing.ToFHIR()
	var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") {
		ops, err := fhir.ParseJSONPatch(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyJSONPatch(currentResource, ops)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else if strings.Contains(contentType, "merge-patch+json") {
		var mp map[string]interface{}
		if err := json.Unmarshal(body, &mp); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mp)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}
	_ = patched
	if v, ok := patched["status"].(string); ok {
		existing.Status = v
	}
	if err := h.svc.UpdateSpecimen(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

func (h *Handler) PatchImagingStudyFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	existing, err := h.svc.GetImagingStudyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImagingStudy", c.Param("id")))
	}
	currentResource := existing.ToFHIR()
	var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") {
		ops, err := fhir.ParseJSONPatch(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyJSONPatch(currentResource, ops)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else if strings.Contains(contentType, "merge-patch+json") {
		var mp map[string]interface{}
		if err := json.Unmarshal(body, &mp); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mp)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}
	_ = patched
	if v, ok := patched["status"].(string); ok {
		existing.Status = v
	}
	if err := h.svc.UpdateImagingStudy(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

// -- FHIR vread and history endpoints --

func (h *Handler) VreadServiceRequestFHIR(c echo.Context) error {
	sr, err := h.svc.GetServiceRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ServiceRequest", c.Param("id")))
	}
	result := sr.ToFHIR()
	fhir.SetVersionHeaders(c, 1, sr.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryServiceRequestFHIR(c echo.Context) error {
	sr, err := h.svc.GetServiceRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ServiceRequest", c.Param("id")))
	}
	result := sr.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ServiceRequest", ResourceID: sr.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: sr.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadDiagnosticReportFHIR(c echo.Context) error {
	dr, err := h.svc.GetDiagnosticReportByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DiagnosticReport", c.Param("id")))
	}
	result := dr.ToFHIR()
	fhir.SetVersionHeaders(c, 1, dr.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryDiagnosticReportFHIR(c echo.Context) error {
	dr, err := h.svc.GetDiagnosticReportByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("DiagnosticReport", c.Param("id")))
	}
	result := dr.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "DiagnosticReport", ResourceID: dr.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: dr.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadImagingStudyFHIR(c echo.Context) error {
	is, err := h.svc.GetImagingStudyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImagingStudy", c.Param("id")))
	}
	result := is.ToFHIR()
	fhir.SetVersionHeaders(c, 1, is.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryImagingStudyFHIR(c echo.Context) error {
	is, err := h.svc.GetImagingStudyByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ImagingStudy", c.Param("id")))
	}
	result := is.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ImagingStudy", ResourceID: is.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: is.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

func (h *Handler) VreadSpecimenFHIR(c echo.Context) error {
	sp, err := h.svc.GetSpecimenByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Specimen", c.Param("id")))
	}
	result := sp.ToFHIR()
	fhir.SetVersionHeaders(c, 1, sp.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistorySpecimenFHIR(c echo.Context) error {
	sp, err := h.svc.GetSpecimenByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Specimen", c.Param("id")))
	}
	result := sp.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Specimen", ResourceID: sp.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: sp.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}
