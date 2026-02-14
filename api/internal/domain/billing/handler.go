package billing

import (
	"net/http"

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
	// Read endpoints – admin, billing
	readGroup := api.Group("", auth.RequireRole("admin", "billing"))
	readGroup.GET("/coverages", h.ListCoverages)
	readGroup.GET("/coverages/:id", h.GetCoverage)
	readGroup.GET("/claims", h.ListClaims)
	readGroup.GET("/claims/:id", h.GetClaim)
	readGroup.GET("/claims/:id/diagnoses", h.GetClaimDiagnoses)
	readGroup.GET("/claims/:id/procedures", h.GetClaimProcedures)
	readGroup.GET("/claims/:id/items", h.GetClaimItems)
	readGroup.GET("/claim-responses", h.ListClaimResponses)
	readGroup.GET("/claim-responses/:id", h.GetClaimResponse)
	readGroup.GET("/invoices", h.ListInvoices)
	readGroup.GET("/invoices/:id", h.GetInvoice)
	readGroup.GET("/invoices/:id/line-items", h.GetInvoiceLineItems)

	// Write endpoints – admin, billing
	writeGroup := api.Group("", auth.RequireRole("admin", "billing"))
	writeGroup.POST("/coverages", h.CreateCoverage)
	writeGroup.PUT("/coverages/:id", h.UpdateCoverage)
	writeGroup.DELETE("/coverages/:id", h.DeleteCoverage)
	writeGroup.POST("/claims", h.CreateClaim)
	writeGroup.PUT("/claims/:id", h.UpdateClaim)
	writeGroup.DELETE("/claims/:id", h.DeleteClaim)
	writeGroup.POST("/claims/:id/diagnoses", h.AddClaimDiagnosis)
	writeGroup.POST("/claims/:id/procedures", h.AddClaimProcedure)
	writeGroup.POST("/claims/:id/items", h.AddClaimItem)
	writeGroup.POST("/claim-responses", h.CreateClaimResponse)
	writeGroup.POST("/invoices", h.CreateInvoice)
	writeGroup.PUT("/invoices/:id", h.UpdateInvoice)
	writeGroup.DELETE("/invoices/:id", h.DeleteInvoice)
	writeGroup.POST("/invoices/:id/line-items", h.AddInvoiceLineItem)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "billing"))
	fhirRead.GET("/Coverage", h.SearchCoveragesFHIR)
	fhirRead.GET("/Coverage/:id", h.GetCoverageFHIR)
	fhirRead.GET("/Claim", h.SearchClaimsFHIR)
	fhirRead.GET("/Claim/:id", h.GetClaimFHIR)
	fhirRead.GET("/ClaimResponse", h.SearchClaimResponsesFHIR)
	fhirRead.GET("/ClaimResponse/:id", h.GetClaimResponseFHIR)
	fhirRead.GET("/ExplanationOfBenefit", h.SearchEOBsFHIR)
	fhirRead.GET("/ExplanationOfBenefit/:id", h.GetEOBFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin", "billing"))
	fhirWrite.POST("/Coverage", h.CreateCoverageFHIR)
	fhirWrite.POST("/Claim", h.CreateClaimFHIR)
	fhirWrite.POST("/ClaimResponse", h.CreateClaimResponseFHIR)
}

// -- Coverage Handlers --

func (h *Handler) CreateCoverage(c echo.Context) error {
	var cov Coverage
	if err := c.Bind(&cov); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateCoverage(c.Request().Context(), &cov); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cov)
}

func (h *Handler) GetCoverage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cov, err := h.svc.GetCoverage(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "coverage not found")
	}
	return c.JSON(http.StatusOK, cov)
}

func (h *Handler) ListCoverages(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListCoveragesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchCoverages(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateCoverage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var cov Coverage
	if err := c.Bind(&cov); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cov.ID = id
	if err := h.svc.UpdateCoverage(c.Request().Context(), &cov); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, cov)
}

func (h *Handler) DeleteCoverage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteCoverage(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Claim Handlers --

func (h *Handler) CreateClaim(c echo.Context) error {
	var cl Claim
	if err := c.Bind(&cl); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateClaim(c.Request().Context(), &cl); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cl)
}

func (h *Handler) GetClaim(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cl, err := h.svc.GetClaim(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "claim not found")
	}
	return c.JSON(http.StatusOK, cl)
}

func (h *Handler) ListClaims(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListClaimsByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchClaims(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateClaim(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var cl Claim
	if err := c.Bind(&cl); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cl.ID = id
	if err := h.svc.UpdateClaim(c.Request().Context(), &cl); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, cl)
}

func (h *Handler) DeleteClaim(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteClaim(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddClaimDiagnosis(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var d ClaimDiagnosis
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	d.ClaimID = id
	if err := h.svc.AddClaimDiagnosis(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, d)
}

func (h *Handler) GetClaimDiagnoses(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetClaimDiagnoses(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

func (h *Handler) AddClaimProcedure(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p ClaimProcedure
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.ClaimID = id
	if err := h.svc.AddClaimProcedure(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetClaimProcedures(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetClaimProcedures(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

func (h *Handler) AddClaimItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var item ClaimItem
	if err := c.Bind(&item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	item.ClaimID = id
	if err := h.svc.AddClaimItem(c.Request().Context(), &item); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetClaimItems(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetClaimItems(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// -- ClaimResponse Handlers --

func (h *Handler) CreateClaimResponse(c echo.Context) error {
	var cr ClaimResponse
	if err := c.Bind(&cr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateClaimResponse(c.Request().Context(), &cr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cr)
}

func (h *Handler) GetClaimResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cr, err := h.svc.GetClaimResponse(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "claim response not found")
	}
	return c.JSON(http.StatusOK, cr)
}

func (h *Handler) ListClaimResponses(c echo.Context) error {
	pg := pagination.FromContext(c)
	if claimID := c.QueryParam("claim_id"); claimID != "" {
		cid, err := uuid.Parse(claimID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid claim_id")
		}
		items, total, err := h.svc.ListClaimResponsesByClaim(c.Request().Context(), cid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchClaimResponses(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

// -- Invoice Handlers --

func (h *Handler) CreateInvoice(c echo.Context) error {
	var inv Invoice
	if err := c.Bind(&inv); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateInvoice(c.Request().Context(), &inv); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, inv)
}

func (h *Handler) GetInvoice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	inv, err := h.svc.GetInvoice(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "invoice not found")
	}
	return c.JSON(http.StatusOK, inv)
}

func (h *Handler) ListInvoices(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListInvoicesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.SearchInvoices(c.Request().Context(), nil, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateInvoice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var inv Invoice
	if err := c.Bind(&inv); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	inv.ID = id
	if err := h.svc.UpdateInvoice(c.Request().Context(), &inv); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, inv)
}

func (h *Handler) DeleteInvoice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteInvoice(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddInvoiceLineItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var li InvoiceLineItem
	if err := c.Bind(&li); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	li.InvoiceID = id
	if err := h.svc.AddInvoiceLineItem(c.Request().Context(), &li); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, li)
}

func (h *Handler) GetInvoiceLineItems(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetInvoiceLineItems(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// -- FHIR Endpoints --

func (h *Handler) SearchCoveragesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "type"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchCoverages(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Coverage"))
}

func (h *Handler) GetCoverageFHIR(c echo.Context) error {
	cov, err := h.svc.GetCoverageByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Coverage", c.Param("id")))
	}
	return c.JSON(http.StatusOK, cov.ToFHIR())
}

func (h *Handler) CreateCoverageFHIR(c echo.Context) error {
	var cov Coverage
	if err := c.Bind(&cov); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateCoverage(c.Request().Context(), &cov); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Coverage/"+cov.FHIRID)
	return c.JSON(http.StatusCreated, cov.ToFHIR())
}

func (h *Handler) SearchClaimsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "use"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchClaims(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Claim"))
}

func (h *Handler) GetClaimFHIR(c echo.Context) error {
	cl, err := h.svc.GetClaimByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Claim", c.Param("id")))
	}
	return c.JSON(http.StatusOK, cl.ToFHIR())
}

func (h *Handler) CreateClaimFHIR(c echo.Context) error {
	var cl Claim
	if err := c.Bind(&cl); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateClaim(c.Request().Context(), &cl); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Claim/"+cl.FHIRID)
	return c.JSON(http.StatusCreated, cl.ToFHIR())
}

func (h *Handler) SearchClaimResponsesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"request", "outcome", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchClaimResponses(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ClaimResponse"))
}

func (h *Handler) GetClaimResponseFHIR(c echo.Context) error {
	cr, err := h.svc.GetClaimResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ClaimResponse", c.Param("id")))
	}
	return c.JSON(http.StatusOK, cr.ToFHIR())
}

func (h *Handler) CreateClaimResponseFHIR(c echo.Context) error {
	var cr ClaimResponse
	if err := c.Bind(&cr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateClaimResponse(c.Request().Context(), &cr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ClaimResponse/"+cr.FHIRID)
	return c.JSON(http.StatusCreated, cr.ToFHIR())
}

func (h *Handler) SearchEOBsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	// EOBs are read from the explanation_of_benefit table; for now, search by patient
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	// EOBs share the coverage search path; we query directly
	_ = pg
	_ = params
	// Return empty bundle since EOB has no dedicated service search yet
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(nil, 0, "/fhir/ExplanationOfBenefit"))
}

func (h *Handler) GetEOBFHIR(c echo.Context) error {
	// EOB read by FHIR ID is not yet fully wired; return not found for now
	return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ExplanationOfBenefit", c.Param("id")))
}
