package financial

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
	// REST read endpoints – admin, billing
	readGroup := api.Group("", auth.RequireRole("admin", "billing"))
	readGroup.GET("/accounts", h.ListAccounts)
	readGroup.GET("/accounts/:id", h.GetAccount)
	readGroup.GET("/insurance-plans", h.ListInsurancePlans)
	readGroup.GET("/insurance-plans/:id", h.GetInsurancePlan)
	readGroup.GET("/payment-notices", h.ListPaymentNotices)
	readGroup.GET("/payment-notices/:id", h.GetPaymentNotice)
	readGroup.GET("/payment-reconciliations", h.ListPaymentReconciliations)
	readGroup.GET("/payment-reconciliations/:id", h.GetPaymentReconciliation)
	readGroup.GET("/charge-items", h.ListChargeItems)
	readGroup.GET("/charge-items/:id", h.GetChargeItem)
	readGroup.GET("/charge-item-definitions", h.ListChargeItemDefinitions)
	readGroup.GET("/charge-item-definitions/:id", h.GetChargeItemDefinition)
	readGroup.GET("/contracts", h.ListContracts)
	readGroup.GET("/contracts/:id", h.GetContract)
	readGroup.GET("/enrollment-requests", h.ListEnrollmentRequests)
	readGroup.GET("/enrollment-requests/:id", h.GetEnrollmentRequest)
	readGroup.GET("/enrollment-responses", h.ListEnrollmentResponses)
	readGroup.GET("/enrollment-responses/:id", h.GetEnrollmentResponse)

	// REST write endpoints – admin, billing
	writeGroup := api.Group("", auth.RequireRole("admin", "billing"))
	writeGroup.POST("/accounts", h.CreateAccount)
	writeGroup.PUT("/accounts/:id", h.UpdateAccount)
	writeGroup.DELETE("/accounts/:id", h.DeleteAccount)
	writeGroup.POST("/insurance-plans", h.CreateInsurancePlan)
	writeGroup.PUT("/insurance-plans/:id", h.UpdateInsurancePlan)
	writeGroup.DELETE("/insurance-plans/:id", h.DeleteInsurancePlan)
	writeGroup.POST("/payment-notices", h.CreatePaymentNotice)
	writeGroup.PUT("/payment-notices/:id", h.UpdatePaymentNotice)
	writeGroup.DELETE("/payment-notices/:id", h.DeletePaymentNotice)
	writeGroup.POST("/payment-reconciliations", h.CreatePaymentReconciliation)
	writeGroup.PUT("/payment-reconciliations/:id", h.UpdatePaymentReconciliation)
	writeGroup.DELETE("/payment-reconciliations/:id", h.DeletePaymentReconciliation)
	writeGroup.POST("/charge-items", h.CreateChargeItem)
	writeGroup.PUT("/charge-items/:id", h.UpdateChargeItem)
	writeGroup.DELETE("/charge-items/:id", h.DeleteChargeItem)
	writeGroup.POST("/charge-item-definitions", h.CreateChargeItemDefinition)
	writeGroup.PUT("/charge-item-definitions/:id", h.UpdateChargeItemDefinition)
	writeGroup.DELETE("/charge-item-definitions/:id", h.DeleteChargeItemDefinition)
	writeGroup.POST("/contracts", h.CreateContract)
	writeGroup.PUT("/contracts/:id", h.UpdateContract)
	writeGroup.DELETE("/contracts/:id", h.DeleteContract)
	writeGroup.POST("/enrollment-requests", h.CreateEnrollmentRequest)
	writeGroup.PUT("/enrollment-requests/:id", h.UpdateEnrollmentRequest)
	writeGroup.DELETE("/enrollment-requests/:id", h.DeleteEnrollmentRequest)
	writeGroup.POST("/enrollment-responses", h.CreateEnrollmentResponse)
	writeGroup.PUT("/enrollment-responses/:id", h.UpdateEnrollmentResponse)
	writeGroup.DELETE("/enrollment-responses/:id", h.DeleteEnrollmentResponse)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "billing"))
	fhirRead.GET("/Account", h.SearchAccountsFHIR)
	fhirRead.GET("/Account/:id", h.GetAccountFHIR)
	fhirRead.GET("/InsurancePlan", h.SearchInsurancePlansFHIR)
	fhirRead.GET("/InsurancePlan/:id", h.GetInsurancePlanFHIR)
	fhirRead.GET("/PaymentNotice", h.SearchPaymentNoticesFHIR)
	fhirRead.GET("/PaymentNotice/:id", h.GetPaymentNoticeFHIR)
	fhirRead.GET("/PaymentReconciliation", h.SearchPaymentReconciliationsFHIR)
	fhirRead.GET("/PaymentReconciliation/:id", h.GetPaymentReconciliationFHIR)
	fhirRead.GET("/ChargeItem", h.SearchChargeItemsFHIR)
	fhirRead.GET("/ChargeItem/:id", h.GetChargeItemFHIR)
	fhirRead.GET("/ChargeItemDefinition", h.SearchChargeItemDefinitionsFHIR)
	fhirRead.GET("/ChargeItemDefinition/:id", h.GetChargeItemDefinitionFHIR)
	fhirRead.GET("/Contract", h.SearchContractsFHIR)
	fhirRead.GET("/Contract/:id", h.GetContractFHIR)
	fhirRead.GET("/EnrollmentRequest", h.SearchEnrollmentRequestsFHIR)
	fhirRead.GET("/EnrollmentRequest/:id", h.GetEnrollmentRequestFHIR)
	fhirRead.GET("/EnrollmentResponse", h.SearchEnrollmentResponsesFHIR)
	fhirRead.GET("/EnrollmentResponse/:id", h.GetEnrollmentResponseFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin", "billing"))
	fhirWrite.POST("/Account", h.CreateAccountFHIR)
	fhirWrite.PUT("/Account/:id", h.UpdateAccountFHIR)
	fhirWrite.DELETE("/Account/:id", h.DeleteAccountFHIR)
	fhirWrite.PATCH("/Account/:id", h.PatchAccountFHIR)
	fhirWrite.POST("/InsurancePlan", h.CreateInsurancePlanFHIR)
	fhirWrite.PUT("/InsurancePlan/:id", h.UpdateInsurancePlanFHIR)
	fhirWrite.DELETE("/InsurancePlan/:id", h.DeleteInsurancePlanFHIR)
	fhirWrite.PATCH("/InsurancePlan/:id", h.PatchInsurancePlanFHIR)
	fhirWrite.POST("/PaymentNotice", h.CreatePaymentNoticeFHIR)
	fhirWrite.PUT("/PaymentNotice/:id", h.UpdatePaymentNoticeFHIR)
	fhirWrite.DELETE("/PaymentNotice/:id", h.DeletePaymentNoticeFHIR)
	fhirWrite.PATCH("/PaymentNotice/:id", h.PatchPaymentNoticeFHIR)
	fhirWrite.POST("/PaymentReconciliation", h.CreatePaymentReconciliationFHIR)
	fhirWrite.PUT("/PaymentReconciliation/:id", h.UpdatePaymentReconciliationFHIR)
	fhirWrite.DELETE("/PaymentReconciliation/:id", h.DeletePaymentReconciliationFHIR)
	fhirWrite.PATCH("/PaymentReconciliation/:id", h.PatchPaymentReconciliationFHIR)
	fhirWrite.POST("/ChargeItem", h.CreateChargeItemFHIR)
	fhirWrite.PUT("/ChargeItem/:id", h.UpdateChargeItemFHIR)
	fhirWrite.DELETE("/ChargeItem/:id", h.DeleteChargeItemFHIR)
	fhirWrite.PATCH("/ChargeItem/:id", h.PatchChargeItemFHIR)
	fhirWrite.POST("/ChargeItemDefinition", h.CreateChargeItemDefinitionFHIR)
	fhirWrite.PUT("/ChargeItemDefinition/:id", h.UpdateChargeItemDefinitionFHIR)
	fhirWrite.DELETE("/ChargeItemDefinition/:id", h.DeleteChargeItemDefinitionFHIR)
	fhirWrite.PATCH("/ChargeItemDefinition/:id", h.PatchChargeItemDefinitionFHIR)
	fhirWrite.POST("/Contract", h.CreateContractFHIR)
	fhirWrite.PUT("/Contract/:id", h.UpdateContractFHIR)
	fhirWrite.DELETE("/Contract/:id", h.DeleteContractFHIR)
	fhirWrite.PATCH("/Contract/:id", h.PatchContractFHIR)
	fhirWrite.POST("/EnrollmentRequest", h.CreateEnrollmentRequestFHIR)
	fhirWrite.PUT("/EnrollmentRequest/:id", h.UpdateEnrollmentRequestFHIR)
	fhirWrite.DELETE("/EnrollmentRequest/:id", h.DeleteEnrollmentRequestFHIR)
	fhirWrite.PATCH("/EnrollmentRequest/:id", h.PatchEnrollmentRequestFHIR)
	fhirWrite.POST("/EnrollmentResponse", h.CreateEnrollmentResponseFHIR)
	fhirWrite.PUT("/EnrollmentResponse/:id", h.UpdateEnrollmentResponseFHIR)
	fhirWrite.DELETE("/EnrollmentResponse/:id", h.DeleteEnrollmentResponseFHIR)
	fhirWrite.PATCH("/EnrollmentResponse/:id", h.PatchEnrollmentResponseFHIR)

	// FHIR POST _search endpoints
	fhirRead.POST("/Account/_search", h.SearchAccountsFHIR)
	fhirRead.POST("/InsurancePlan/_search", h.SearchInsurancePlansFHIR)
	fhirRead.POST("/PaymentNotice/_search", h.SearchPaymentNoticesFHIR)
	fhirRead.POST("/PaymentReconciliation/_search", h.SearchPaymentReconciliationsFHIR)
	fhirRead.POST("/ChargeItem/_search", h.SearchChargeItemsFHIR)
	fhirRead.POST("/ChargeItemDefinition/_search", h.SearchChargeItemDefinitionsFHIR)
	fhirRead.POST("/Contract/_search", h.SearchContractsFHIR)
	fhirRead.POST("/EnrollmentRequest/_search", h.SearchEnrollmentRequestsFHIR)
	fhirRead.POST("/EnrollmentResponse/_search", h.SearchEnrollmentResponsesFHIR)

	// FHIR vread and history endpoints
	fhirRead.GET("/Account/:id/_history/:vid", h.VreadAccountFHIR)
	fhirRead.GET("/Account/:id/_history", h.HistoryAccountFHIR)
	fhirRead.GET("/InsurancePlan/:id/_history/:vid", h.VreadInsurancePlanFHIR)
	fhirRead.GET("/InsurancePlan/:id/_history", h.HistoryInsurancePlanFHIR)
	fhirRead.GET("/PaymentNotice/:id/_history/:vid", h.VreadPaymentNoticeFHIR)
	fhirRead.GET("/PaymentNotice/:id/_history", h.HistoryPaymentNoticeFHIR)
	fhirRead.GET("/PaymentReconciliation/:id/_history/:vid", h.VreadPaymentReconciliationFHIR)
	fhirRead.GET("/PaymentReconciliation/:id/_history", h.HistoryPaymentReconciliationFHIR)
	fhirRead.GET("/ChargeItem/:id/_history/:vid", h.VreadChargeItemFHIR)
	fhirRead.GET("/ChargeItem/:id/_history", h.HistoryChargeItemFHIR)
	fhirRead.GET("/ChargeItemDefinition/:id/_history/:vid", h.VreadChargeItemDefinitionFHIR)
	fhirRead.GET("/ChargeItemDefinition/:id/_history", h.HistoryChargeItemDefinitionFHIR)
	fhirRead.GET("/Contract/:id/_history/:vid", h.VreadContractFHIR)
	fhirRead.GET("/Contract/:id/_history", h.HistoryContractFHIR)
	fhirRead.GET("/EnrollmentRequest/:id/_history/:vid", h.VreadEnrollmentRequestFHIR)
	fhirRead.GET("/EnrollmentRequest/:id/_history", h.HistoryEnrollmentRequestFHIR)
	fhirRead.GET("/EnrollmentResponse/:id/_history/:vid", h.VreadEnrollmentResponseFHIR)
	fhirRead.GET("/EnrollmentResponse/:id/_history", h.HistoryEnrollmentResponseFHIR)
}

// ===== REST Account Handlers =====

func (h *Handler) CreateAccount(c echo.Context) error {
	var a Account
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateAccount(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) GetAccount(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.GetAccount(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "account not found")
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) ListAccounts(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListAccounts(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateAccount(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var a Account
	if err := c.Bind(&a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	a.ID = id
	if err := h.svc.UpdateAccount(c.Request().Context(), &a); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) DeleteAccount(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteAccount(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ===== REST InsurancePlan Handlers =====

func (h *Handler) CreateInsurancePlan(c echo.Context) error {
	var ip InsurancePlan
	if err := c.Bind(&ip); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateInsurancePlan(c.Request().Context(), &ip); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ip)
}

func (h *Handler) GetInsurancePlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ip, err := h.svc.GetInsurancePlan(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "insurance plan not found")
	}
	return c.JSON(http.StatusOK, ip)
}

func (h *Handler) ListInsurancePlans(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListInsurancePlans(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateInsurancePlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ip InsurancePlan
	if err := c.Bind(&ip); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ip.ID = id
	if err := h.svc.UpdateInsurancePlan(c.Request().Context(), &ip); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ip)
}

func (h *Handler) DeleteInsurancePlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteInsurancePlan(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ===== REST PaymentNotice Handlers =====

func (h *Handler) CreatePaymentNotice(c echo.Context) error {
	var pn PaymentNotice
	if err := c.Bind(&pn); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePaymentNotice(c.Request().Context(), &pn); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, pn)
}

func (h *Handler) GetPaymentNotice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	pn, err := h.svc.GetPaymentNotice(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "payment notice not found")
	}
	return c.JSON(http.StatusOK, pn)
}

func (h *Handler) ListPaymentNotices(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListPaymentNotices(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePaymentNotice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var pn PaymentNotice
	if err := c.Bind(&pn); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	pn.ID = id
	if err := h.svc.UpdatePaymentNotice(c.Request().Context(), &pn); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, pn)
}

func (h *Handler) DeletePaymentNotice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePaymentNotice(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ===== REST PaymentReconciliation Handlers =====

func (h *Handler) CreatePaymentReconciliation(c echo.Context) error {
	var pr PaymentReconciliation
	if err := c.Bind(&pr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePaymentReconciliation(c.Request().Context(), &pr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, pr)
}

func (h *Handler) GetPaymentReconciliation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	pr, err := h.svc.GetPaymentReconciliation(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "payment reconciliation not found")
	}
	return c.JSON(http.StatusOK, pr)
}

func (h *Handler) ListPaymentReconciliations(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListPaymentReconciliations(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePaymentReconciliation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var pr PaymentReconciliation
	if err := c.Bind(&pr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	pr.ID = id
	if err := h.svc.UpdatePaymentReconciliation(c.Request().Context(), &pr); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, pr)
}

func (h *Handler) DeletePaymentReconciliation(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePaymentReconciliation(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ===== REST ChargeItem Handlers =====

func (h *Handler) CreateChargeItem(c echo.Context) error {
	var ci ChargeItem
	if err := c.Bind(&ci); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateChargeItem(c.Request().Context(), &ci); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ci)
}

func (h *Handler) GetChargeItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ci, err := h.svc.GetChargeItem(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "charge item not found")
	}
	return c.JSON(http.StatusOK, ci)
}

func (h *Handler) ListChargeItems(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListChargeItems(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateChargeItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ci ChargeItem
	if err := c.Bind(&ci); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ci.ID = id
	if err := h.svc.UpdateChargeItem(c.Request().Context(), &ci); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ci)
}

func (h *Handler) DeleteChargeItem(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteChargeItem(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ===== REST ChargeItemDefinition Handlers =====

func (h *Handler) CreateChargeItemDefinition(c echo.Context) error {
	var cd ChargeItemDefinition
	if err := c.Bind(&cd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateChargeItemDefinition(c.Request().Context(), &cd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, cd)
}

func (h *Handler) GetChargeItemDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cd, err := h.svc.GetChargeItemDefinition(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "charge item definition not found")
	}
	return c.JSON(http.StatusOK, cd)
}

func (h *Handler) ListChargeItemDefinitions(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListChargeItemDefinitions(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateChargeItemDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var cd ChargeItemDefinition
	if err := c.Bind(&cd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cd.ID = id
	if err := h.svc.UpdateChargeItemDefinition(c.Request().Context(), &cd); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, cd)
}

func (h *Handler) DeleteChargeItemDefinition(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteChargeItemDefinition(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ===== REST Contract Handlers =====

func (h *Handler) CreateContract(c echo.Context) error {
	var ct Contract
	if err := c.Bind(&ct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateContract(c.Request().Context(), &ct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ct)
}

func (h *Handler) GetContract(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ct, err := h.svc.GetContract(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "contract not found")
	}
	return c.JSON(http.StatusOK, ct)
}

func (h *Handler) ListContracts(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListContracts(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateContract(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ct Contract
	if err := c.Bind(&ct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ct.ID = id
	if err := h.svc.UpdateContract(c.Request().Context(), &ct); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, ct)
}

func (h *Handler) DeleteContract(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteContract(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ===== REST EnrollmentRequest Handlers =====

func (h *Handler) CreateEnrollmentRequest(c echo.Context) error {
	var er EnrollmentRequest
	if err := c.Bind(&er); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateEnrollmentRequest(c.Request().Context(), &er); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, er)
}

func (h *Handler) GetEnrollmentRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	er, err := h.svc.GetEnrollmentRequest(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "enrollment request not found")
	}
	return c.JSON(http.StatusOK, er)
}

func (h *Handler) ListEnrollmentRequests(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListEnrollmentRequests(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateEnrollmentRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var er EnrollmentRequest
	if err := c.Bind(&er); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	er.ID = id
	if err := h.svc.UpdateEnrollmentRequest(c.Request().Context(), &er); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, er)
}

func (h *Handler) DeleteEnrollmentRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteEnrollmentRequest(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ===== REST EnrollmentResponse Handlers =====

func (h *Handler) CreateEnrollmentResponse(c echo.Context) error {
	var er EnrollmentResponse
	if err := c.Bind(&er); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateEnrollmentResponse(c.Request().Context(), &er); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, er)
}

func (h *Handler) GetEnrollmentResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	er, err := h.svc.GetEnrollmentResponse(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "enrollment response not found")
	}
	return c.JSON(http.StatusOK, er)
}

func (h *Handler) ListEnrollmentResponses(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListEnrollmentResponses(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateEnrollmentResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var er EnrollmentResponse
	if err := c.Bind(&er); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	er.ID = id
	if err := h.svc.UpdateEnrollmentResponse(c.Request().Context(), &er); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, er)
}

func (h *Handler) DeleteEnrollmentResponse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteEnrollmentResponse(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ===== FHIR Account Endpoints =====

func (h *Handler) SearchAccountsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "name", "patient"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchAccounts(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Account"))
}

func (h *Handler) GetAccountFHIR(c echo.Context) error {
	a, err := h.svc.GetAccountByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Account", c.Param("id")))
	}
	return c.JSON(http.StatusOK, a.ToFHIR())
}

func (h *Handler) CreateAccountFHIR(c echo.Context) error {
	var a Account
	if err := c.Bind(&a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateAccount(c.Request().Context(), &a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Account/"+a.FHIRID)
	return c.JSON(http.StatusCreated, a.ToFHIR())
}

func (h *Handler) UpdateAccountFHIR(c echo.Context) error {
	var a Account
	if err := c.Bind(&a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetAccountByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Account", c.Param("id")))
	}
	a.ID = existing.ID
	a.FHIRID = existing.FHIRID
	if err := h.svc.UpdateAccount(c.Request().Context(), &a); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, a.ToFHIR())
}

func (h *Handler) DeleteAccountFHIR(c echo.Context) error {
	existing, err := h.svc.GetAccountByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Account", c.Param("id")))
	}
	if err := h.svc.DeleteAccount(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchAccountFHIR(c echo.Context) error {
	return h.handlePatch(c, "Account", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetAccountByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Account", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateAccount(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadAccountFHIR(c echo.Context) error {
	a, err := h.svc.GetAccountByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Account", c.Param("id")))
	}
	result := a.ToFHIR()
	fhir.SetVersionHeaders(c, 1, a.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryAccountFHIR(c echo.Context) error {
	a, err := h.svc.GetAccountByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Account", c.Param("id")))
	}
	result := a.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Account", ResourceID: a.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: a.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ===== FHIR InsurancePlan Endpoints =====

func (h *Handler) SearchInsurancePlansFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "name", "type"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchInsurancePlans(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/InsurancePlan"))
}

func (h *Handler) GetInsurancePlanFHIR(c echo.Context) error {
	ip, err := h.svc.GetInsurancePlanByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("InsurancePlan", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ip.ToFHIR())
}

func (h *Handler) CreateInsurancePlanFHIR(c echo.Context) error {
	var ip InsurancePlan
	if err := c.Bind(&ip); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateInsurancePlan(c.Request().Context(), &ip); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/InsurancePlan/"+ip.FHIRID)
	return c.JSON(http.StatusCreated, ip.ToFHIR())
}

func (h *Handler) UpdateInsurancePlanFHIR(c echo.Context) error {
	var ip InsurancePlan
	if err := c.Bind(&ip); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetInsurancePlanByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("InsurancePlan", c.Param("id")))
	}
	ip.ID = existing.ID
	ip.FHIRID = existing.FHIRID
	if err := h.svc.UpdateInsurancePlan(c.Request().Context(), &ip); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ip.ToFHIR())
}

func (h *Handler) DeleteInsurancePlanFHIR(c echo.Context) error {
	existing, err := h.svc.GetInsurancePlanByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("InsurancePlan", c.Param("id")))
	}
	if err := h.svc.DeleteInsurancePlan(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchInsurancePlanFHIR(c echo.Context) error {
	return h.handlePatch(c, "InsurancePlan", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetInsurancePlanByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("InsurancePlan", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateInsurancePlan(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadInsurancePlanFHIR(c echo.Context) error {
	ip, err := h.svc.GetInsurancePlanByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("InsurancePlan", c.Param("id")))
	}
	result := ip.ToFHIR()
	fhir.SetVersionHeaders(c, 1, ip.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryInsurancePlanFHIR(c echo.Context) error {
	ip, err := h.svc.GetInsurancePlanByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("InsurancePlan", c.Param("id")))
	}
	result := ip.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "InsurancePlan", ResourceID: ip.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: ip.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ===== FHIR PaymentNotice Endpoints =====

func (h *Handler) SearchPaymentNoticesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "provider"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchPaymentNotices(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/PaymentNotice"))
}

func (h *Handler) GetPaymentNoticeFHIR(c echo.Context) error {
	pn, err := h.svc.GetPaymentNoticeByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentNotice", c.Param("id")))
	}
	return c.JSON(http.StatusOK, pn.ToFHIR())
}

func (h *Handler) CreatePaymentNoticeFHIR(c echo.Context) error {
	var pn PaymentNotice
	if err := c.Bind(&pn); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreatePaymentNotice(c.Request().Context(), &pn); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/PaymentNotice/"+pn.FHIRID)
	return c.JSON(http.StatusCreated, pn.ToFHIR())
}

func (h *Handler) UpdatePaymentNoticeFHIR(c echo.Context) error {
	var pn PaymentNotice
	if err := c.Bind(&pn); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetPaymentNoticeByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentNotice", c.Param("id")))
	}
	pn.ID = existing.ID
	pn.FHIRID = existing.FHIRID
	if err := h.svc.UpdatePaymentNotice(c.Request().Context(), &pn); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, pn.ToFHIR())
}

func (h *Handler) DeletePaymentNoticeFHIR(c echo.Context) error {
	existing, err := h.svc.GetPaymentNoticeByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentNotice", c.Param("id")))
	}
	if err := h.svc.DeletePaymentNotice(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchPaymentNoticeFHIR(c echo.Context) error {
	return h.handlePatch(c, "PaymentNotice", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetPaymentNoticeByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentNotice", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdatePaymentNotice(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadPaymentNoticeFHIR(c echo.Context) error {
	pn, err := h.svc.GetPaymentNoticeByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentNotice", c.Param("id")))
	}
	result := pn.ToFHIR()
	fhir.SetVersionHeaders(c, 1, pn.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryPaymentNoticeFHIR(c echo.Context) error {
	pn, err := h.svc.GetPaymentNoticeByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentNotice", c.Param("id")))
	}
	result := pn.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "PaymentNotice", ResourceID: pn.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: pn.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ===== FHIR PaymentReconciliation Endpoints =====

func (h *Handler) SearchPaymentReconciliationsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "outcome"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchPaymentReconciliations(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/PaymentReconciliation"))
}

func (h *Handler) GetPaymentReconciliationFHIR(c echo.Context) error {
	pr, err := h.svc.GetPaymentReconciliationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentReconciliation", c.Param("id")))
	}
	return c.JSON(http.StatusOK, pr.ToFHIR())
}

func (h *Handler) CreatePaymentReconciliationFHIR(c echo.Context) error {
	var pr PaymentReconciliation
	if err := c.Bind(&pr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreatePaymentReconciliation(c.Request().Context(), &pr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/PaymentReconciliation/"+pr.FHIRID)
	return c.JSON(http.StatusCreated, pr.ToFHIR())
}

func (h *Handler) UpdatePaymentReconciliationFHIR(c echo.Context) error {
	var pr PaymentReconciliation
	if err := c.Bind(&pr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetPaymentReconciliationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentReconciliation", c.Param("id")))
	}
	pr.ID = existing.ID
	pr.FHIRID = existing.FHIRID
	if err := h.svc.UpdatePaymentReconciliation(c.Request().Context(), &pr); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, pr.ToFHIR())
}

func (h *Handler) DeletePaymentReconciliationFHIR(c echo.Context) error {
	existing, err := h.svc.GetPaymentReconciliationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentReconciliation", c.Param("id")))
	}
	if err := h.svc.DeletePaymentReconciliation(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchPaymentReconciliationFHIR(c echo.Context) error {
	return h.handlePatch(c, "PaymentReconciliation", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetPaymentReconciliationByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentReconciliation", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdatePaymentReconciliation(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadPaymentReconciliationFHIR(c echo.Context) error {
	pr, err := h.svc.GetPaymentReconciliationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentReconciliation", c.Param("id")))
	}
	result := pr.ToFHIR()
	fhir.SetVersionHeaders(c, 1, pr.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryPaymentReconciliationFHIR(c echo.Context) error {
	pr, err := h.svc.GetPaymentReconciliationByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("PaymentReconciliation", c.Param("id")))
	}
	result := pr.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "PaymentReconciliation", ResourceID: pr.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: pr.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ===== FHIR ChargeItem Endpoints =====

func (h *Handler) SearchChargeItemsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status", "code"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchChargeItems(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ChargeItem"))
}

func (h *Handler) GetChargeItemFHIR(c echo.Context) error {
	ci, err := h.svc.GetChargeItemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItem", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ci.ToFHIR())
}

func (h *Handler) CreateChargeItemFHIR(c echo.Context) error {
	var ci ChargeItem
	if err := c.Bind(&ci); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateChargeItem(c.Request().Context(), &ci); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ChargeItem/"+ci.FHIRID)
	return c.JSON(http.StatusCreated, ci.ToFHIR())
}

func (h *Handler) UpdateChargeItemFHIR(c echo.Context) error {
	var ci ChargeItem
	if err := c.Bind(&ci); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetChargeItemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItem", c.Param("id")))
	}
	ci.ID = existing.ID
	ci.FHIRID = existing.FHIRID
	if err := h.svc.UpdateChargeItem(c.Request().Context(), &ci); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ci.ToFHIR())
}

func (h *Handler) DeleteChargeItemFHIR(c echo.Context) error {
	existing, err := h.svc.GetChargeItemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItem", c.Param("id")))
	}
	if err := h.svc.DeleteChargeItem(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchChargeItemFHIR(c echo.Context) error {
	return h.handlePatch(c, "ChargeItem", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetChargeItemByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItem", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateChargeItem(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadChargeItemFHIR(c echo.Context) error {
	ci, err := h.svc.GetChargeItemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItem", c.Param("id")))
	}
	result := ci.ToFHIR()
	fhir.SetVersionHeaders(c, 1, ci.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryChargeItemFHIR(c echo.Context) error {
	ci, err := h.svc.GetChargeItemByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItem", c.Param("id")))
	}
	result := ci.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ChargeItem", ResourceID: ci.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: ci.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ===== FHIR ChargeItemDefinition Endpoints =====

func (h *Handler) SearchChargeItemDefinitionsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "title", "code"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchChargeItemDefinitions(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/ChargeItemDefinition"))
}

func (h *Handler) GetChargeItemDefinitionFHIR(c echo.Context) error {
	cd, err := h.svc.GetChargeItemDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItemDefinition", c.Param("id")))
	}
	return c.JSON(http.StatusOK, cd.ToFHIR())
}

func (h *Handler) CreateChargeItemDefinitionFHIR(c echo.Context) error {
	var cd ChargeItemDefinition
	if err := c.Bind(&cd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateChargeItemDefinition(c.Request().Context(), &cd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/ChargeItemDefinition/"+cd.FHIRID)
	return c.JSON(http.StatusCreated, cd.ToFHIR())
}

func (h *Handler) UpdateChargeItemDefinitionFHIR(c echo.Context) error {
	var cd ChargeItemDefinition
	if err := c.Bind(&cd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetChargeItemDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItemDefinition", c.Param("id")))
	}
	cd.ID = existing.ID
	cd.FHIRID = existing.FHIRID
	if err := h.svc.UpdateChargeItemDefinition(c.Request().Context(), &cd); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, cd.ToFHIR())
}

func (h *Handler) DeleteChargeItemDefinitionFHIR(c echo.Context) error {
	existing, err := h.svc.GetChargeItemDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItemDefinition", c.Param("id")))
	}
	if err := h.svc.DeleteChargeItemDefinition(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchChargeItemDefinitionFHIR(c echo.Context) error {
	return h.handlePatch(c, "ChargeItemDefinition", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetChargeItemDefinitionByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItemDefinition", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateChargeItemDefinition(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadChargeItemDefinitionFHIR(c echo.Context) error {
	cd, err := h.svc.GetChargeItemDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItemDefinition", c.Param("id")))
	}
	result := cd.ToFHIR()
	fhir.SetVersionHeaders(c, 1, cd.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryChargeItemDefinitionFHIR(c echo.Context) error {
	cd, err := h.svc.GetChargeItemDefinitionByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("ChargeItemDefinition", c.Param("id")))
	}
	result := cd.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "ChargeItemDefinition", ResourceID: cd.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: cd.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ===== FHIR Contract Endpoints =====

func (h *Handler) SearchContractsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "patient"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchContracts(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Contract"))
}

func (h *Handler) GetContractFHIR(c echo.Context) error {
	ct, err := h.svc.GetContractByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Contract", c.Param("id")))
	}
	return c.JSON(http.StatusOK, ct.ToFHIR())
}

func (h *Handler) CreateContractFHIR(c echo.Context) error {
	var ct Contract
	if err := c.Bind(&ct); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateContract(c.Request().Context(), &ct); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Contract/"+ct.FHIRID)
	return c.JSON(http.StatusCreated, ct.ToFHIR())
}

func (h *Handler) UpdateContractFHIR(c echo.Context) error {
	var ct Contract
	if err := c.Bind(&ct); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetContractByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Contract", c.Param("id")))
	}
	ct.ID = existing.ID
	ct.FHIRID = existing.FHIRID
	if err := h.svc.UpdateContract(c.Request().Context(), &ct); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, ct.ToFHIR())
}

func (h *Handler) DeleteContractFHIR(c echo.Context) error {
	existing, err := h.svc.GetContractByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Contract", c.Param("id")))
	}
	if err := h.svc.DeleteContract(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchContractFHIR(c echo.Context) error {
	return h.handlePatch(c, "Contract", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetContractByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Contract", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateContract(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadContractFHIR(c echo.Context) error {
	ct, err := h.svc.GetContractByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Contract", c.Param("id")))
	}
	result := ct.ToFHIR()
	fhir.SetVersionHeaders(c, 1, ct.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryContractFHIR(c echo.Context) error {
	ct, err := h.svc.GetContractByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Contract", c.Param("id")))
	}
	result := ct.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "Contract", ResourceID: ct.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: ct.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ===== FHIR EnrollmentRequest Endpoints =====

func (h *Handler) SearchEnrollmentRequestsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "patient"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchEnrollmentRequests(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/EnrollmentRequest"))
}

func (h *Handler) GetEnrollmentRequestFHIR(c echo.Context) error {
	er, err := h.svc.GetEnrollmentRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentRequest", c.Param("id")))
	}
	return c.JSON(http.StatusOK, er.ToFHIR())
}

func (h *Handler) CreateEnrollmentRequestFHIR(c echo.Context) error {
	var er EnrollmentRequest
	if err := c.Bind(&er); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateEnrollmentRequest(c.Request().Context(), &er); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/EnrollmentRequest/"+er.FHIRID)
	return c.JSON(http.StatusCreated, er.ToFHIR())
}

func (h *Handler) UpdateEnrollmentRequestFHIR(c echo.Context) error {
	var er EnrollmentRequest
	if err := c.Bind(&er); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetEnrollmentRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentRequest", c.Param("id")))
	}
	er.ID = existing.ID
	er.FHIRID = existing.FHIRID
	if err := h.svc.UpdateEnrollmentRequest(c.Request().Context(), &er); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, er.ToFHIR())
}

func (h *Handler) DeleteEnrollmentRequestFHIR(c echo.Context) error {
	existing, err := h.svc.GetEnrollmentRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentRequest", c.Param("id")))
	}
	if err := h.svc.DeleteEnrollmentRequest(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchEnrollmentRequestFHIR(c echo.Context) error {
	return h.handlePatch(c, "EnrollmentRequest", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetEnrollmentRequestByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentRequest", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateEnrollmentRequest(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadEnrollmentRequestFHIR(c echo.Context) error {
	er, err := h.svc.GetEnrollmentRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentRequest", c.Param("id")))
	}
	result := er.ToFHIR()
	fhir.SetVersionHeaders(c, 1, er.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryEnrollmentRequestFHIR(c echo.Context) error {
	er, err := h.svc.GetEnrollmentRequestByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentRequest", c.Param("id")))
	}
	result := er.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "EnrollmentRequest", ResourceID: er.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: er.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ===== FHIR EnrollmentResponse Endpoints =====

func (h *Handler) SearchEnrollmentResponsesFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"status", "request"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.SearchEnrollmentResponses(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/EnrollmentResponse"))
}

func (h *Handler) GetEnrollmentResponseFHIR(c echo.Context) error {
	er, err := h.svc.GetEnrollmentResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentResponse", c.Param("id")))
	}
	return c.JSON(http.StatusOK, er.ToFHIR())
}

func (h *Handler) CreateEnrollmentResponseFHIR(c echo.Context) error {
	var er EnrollmentResponse
	if err := c.Bind(&er); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreateEnrollmentResponse(c.Request().Context(), &er); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/EnrollmentResponse/"+er.FHIRID)
	return c.JSON(http.StatusCreated, er.ToFHIR())
}

func (h *Handler) UpdateEnrollmentResponseFHIR(c echo.Context) error {
	var er EnrollmentResponse
	if err := c.Bind(&er); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetEnrollmentResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentResponse", c.Param("id")))
	}
	er.ID = existing.ID
	er.FHIRID = existing.FHIRID
	if err := h.svc.UpdateEnrollmentResponse(c.Request().Context(), &er); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, er.ToFHIR())
}

func (h *Handler) DeleteEnrollmentResponseFHIR(c echo.Context) error {
	existing, err := h.svc.GetEnrollmentResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentResponse", c.Param("id")))
	}
	if err := h.svc.DeleteEnrollmentResponse(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) PatchEnrollmentResponseFHIR(c echo.Context) error {
	return h.handlePatch(c, "EnrollmentResponse", c.Param("id"), func(ctx echo.Context, resource map[string]interface{}) error {
		existing, err := h.svc.GetEnrollmentResponseByFHIRID(ctx.Request().Context(), ctx.Param("id"))
		if err != nil {
			return ctx.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentResponse", ctx.Param("id")))
		}
		if v, ok := resource["status"].(string); ok {
			existing.Status = v
		}
		if err := h.svc.UpdateEnrollmentResponse(ctx.Request().Context(), existing); err != nil {
			return ctx.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		return ctx.JSON(http.StatusOK, existing.ToFHIR())
	})
}

func (h *Handler) VreadEnrollmentResponseFHIR(c echo.Context) error {
	er, err := h.svc.GetEnrollmentResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentResponse", c.Param("id")))
	}
	result := er.ToFHIR()
	fhir.SetVersionHeaders(c, 1, er.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) HistoryEnrollmentResponseFHIR(c echo.Context) error {
	er, err := h.svc.GetEnrollmentResponseByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("EnrollmentResponse", c.Param("id")))
	}
	result := er.ToFHIR()
	raw, _ := json.Marshal(result)
	entry := &fhir.HistoryEntry{
		ResourceType: "EnrollmentResponse", ResourceID: er.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: er.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}

// ===== handlePatch =====

func (h *Handler) handlePatch(c echo.Context, resourceType, fhirID string, applyFn func(echo.Context, map[string]interface{}) error) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	var currentResource map[string]interface{}
	switch resourceType {
	case "Account":
		existing, err := h.svc.GetAccountByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "InsurancePlan":
		existing, err := h.svc.GetInsurancePlanByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "PaymentNotice":
		existing, err := h.svc.GetPaymentNoticeByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "PaymentReconciliation":
		existing, err := h.svc.GetPaymentReconciliationByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "ChargeItem":
		existing, err := h.svc.GetChargeItemByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "ChargeItemDefinition":
		existing, err := h.svc.GetChargeItemDefinitionByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "Contract":
		existing, err := h.svc.GetContractByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "EnrollmentRequest":
		existing, err := h.svc.GetEnrollmentRequestByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	case "EnrollmentResponse":
		existing, err := h.svc.GetEnrollmentResponseByFHIRID(c.Request().Context(), fhirID)
		if err != nil {
			return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome(resourceType, fhirID))
		}
		currentResource = existing.ToFHIR()
	default:
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("unsupported resource type for PATCH"))
	}

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
		var mergePatch map[string]interface{}
		if err := json.Unmarshal(body, &mergePatch); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mergePatch)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}

	return applyFn(c, patched)
}
