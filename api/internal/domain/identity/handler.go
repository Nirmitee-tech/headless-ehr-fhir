package identity

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
	// Read endpoints – admin, physician, nurse, registrar, pharmacist, lab_tech
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar", "pharmacist", "lab_tech"))
	readGroup.GET("/patients", h.ListPatients)
	readGroup.GET("/patients/:id", h.GetPatient)
	readGroup.GET("/patients/:id/contacts", h.GetPatientContacts)
	readGroup.GET("/patients/:id/identifiers", h.GetPatientIdentifiers)
	readGroup.GET("/practitioners", h.ListPractitioners)
	readGroup.GET("/practitioners/:id", h.GetPractitioner)
	readGroup.GET("/practitioners/:id/roles", h.GetPractitionerRoles)

	// Write endpoints – admin, registrar
	writeGroup := api.Group("", auth.RequireRole("admin", "registrar"))
	writeGroup.POST("/patients", h.CreatePatient)
	writeGroup.PUT("/patients/:id", h.UpdatePatient)
	writeGroup.DELETE("/patients/:id", h.DeletePatient)
	writeGroup.POST("/patients/:id/contacts", h.AddPatientContact)
	writeGroup.DELETE("/patients/:id/contacts/:contact_id", h.RemovePatientContact)
	writeGroup.POST("/patients/:id/identifiers", h.AddPatientIdentifier)
	writeGroup.POST("/practitioners", h.CreatePractitioner)
	writeGroup.PUT("/practitioners/:id", h.UpdatePractitioner)
	writeGroup.DELETE("/practitioners/:id", h.DeletePractitioner)
	writeGroup.POST("/practitioners/:id/roles", h.AddPractitionerRole)

	// FHIR read endpoints
	fhirRead := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse", "registrar", "pharmacist", "lab_tech"))
	fhirRead.GET("/Patient", h.SearchPatientsFHIR)
	fhirRead.GET("/Patient/:id", h.GetPatientFHIR)
	fhirRead.GET("/Practitioner", h.SearchPractitionersFHIR)
	fhirRead.GET("/Practitioner/:id", h.GetPractitionerFHIR)

	// FHIR write endpoints
	fhirWrite := fhirGroup.Group("", auth.RequireRole("admin", "registrar"))
	fhirWrite.POST("/Patient", h.CreatePatientFHIR)
	fhirWrite.PUT("/Patient/:id", h.UpdatePatientFHIR)
	fhirWrite.DELETE("/Patient/:id", h.DeletePatientFHIR)
	fhirWrite.POST("/Practitioner", h.CreatePractitionerFHIR)
	fhirWrite.PUT("/Practitioner/:id", h.UpdatePractitionerFHIR)
	fhirWrite.DELETE("/Practitioner/:id", h.DeletePractitionerFHIR)
}

// -- Patient Operational Handlers --

func (h *Handler) CreatePatient(c echo.Context) error {
	var p Patient
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePatient(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetPatient(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := h.svc.GetPatient(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "patient not found")
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) ListPatients(c echo.Context) error {
	pg := pagination.FromContext(c)
	patients, total, err := h.svc.ListPatients(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(patients, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePatient(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p Patient
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.ID = id
	if err := h.svc.UpdatePatient(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) DeletePatient(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePatient(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddPatientContact(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var contact PatientContact
	if err := c.Bind(&contact); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	contact.PatientID = id
	if err := h.svc.AddPatientContact(c.Request().Context(), &contact); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, contact)
}

func (h *Handler) GetPatientContacts(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	contacts, err := h.svc.GetPatientContacts(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, contacts)
}

func (h *Handler) RemovePatientContact(c echo.Context) error {
	contactID, err := uuid.Parse(c.Param("contact_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid contact_id")
	}
	if err := h.svc.RemovePatientContact(c.Request().Context(), contactID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddPatientIdentifier(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var ident PatientIdentifier
	if err := c.Bind(&ident); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ident.PatientID = id
	if err := h.svc.AddPatientIdentifier(c.Request().Context(), &ident); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, ident)
}

func (h *Handler) GetPatientIdentifiers(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	idents, err := h.svc.GetPatientIdentifiers(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, idents)
}

// -- Practitioner Operational Handlers --

func (h *Handler) CreatePractitioner(c echo.Context) error {
	var p Practitioner
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePractitioner(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetPractitioner(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := h.svc.GetPractitioner(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "practitioner not found")
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) ListPractitioners(c echo.Context) error {
	pg := pagination.FromContext(c)
	practs, total, err := h.svc.ListPractitioners(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(practs, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePractitioner(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p Practitioner
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.ID = id
	if err := h.svc.UpdatePractitioner(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) DeletePractitioner(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePractitioner(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddPractitionerRole(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var role PractitionerRole
	if err := c.Bind(&role); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	role.PractitionerID = id
	if err := h.svc.AddPractitionerRole(c.Request().Context(), &role); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, role)
}

func (h *Handler) GetPractitionerRoles(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	roles, err := h.svc.GetPractitionerRoles(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, roles)
}

// -- FHIR Patient Handlers --

func (h *Handler) SearchPatientsFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, key := range []string{"name", "family", "given", "birthdate", "gender", "identifier"} {
		if v := c.QueryParam(key); v != "" {
			params[key] = v
		}
	}

	patients, total, err := h.svc.SearchPatients(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}

	resources := make([]interface{}, len(patients))
	for i, p := range patients {
		resources[i] = p.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Patient"))
}

func (h *Handler) GetPatientFHIR(c echo.Context) error {
	p, err := h.svc.GetPatientByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Patient", c.Param("id")))
	}
	return c.JSON(http.StatusOK, p.ToFHIR())
}

func (h *Handler) CreatePatientFHIR(c echo.Context) error {
	var p Patient
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreatePatient(c.Request().Context(), &p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Patient/"+p.FHIRID)
	return c.JSON(http.StatusCreated, p.ToFHIR())
}

func (h *Handler) UpdatePatientFHIR(c echo.Context) error {
	existing, err := h.svc.GetPatientByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Patient", c.Param("id")))
	}
	var p Patient
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	p.ID = existing.ID
	p.FHIRID = existing.FHIRID
	if err := h.svc.UpdatePatient(c.Request().Context(), &p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, p.ToFHIR())
}

func (h *Handler) DeletePatientFHIR(c echo.Context) error {
	p, err := h.svc.GetPatientByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Patient", c.Param("id")))
	}
	if err := h.svc.DeletePatient(c.Request().Context(), p.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// -- FHIR Practitioner Handlers --

func (h *Handler) SearchPractitionersFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, key := range []string{"name", "family", "identifier"} {
		if v := c.QueryParam(key); v != "" {
			params[key] = v
		}
	}

	practs, total, err := h.svc.SearchPractitioners(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}

	resources := make([]interface{}, len(practs))
	for i, p := range practs {
		resources[i] = p.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/Practitioner"))
}

func (h *Handler) GetPractitionerFHIR(c echo.Context) error {
	p, err := h.svc.GetPractitionerByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Practitioner", c.Param("id")))
	}
	return c.JSON(http.StatusOK, p.ToFHIR())
}

func (h *Handler) CreatePractitionerFHIR(c echo.Context) error {
	var p Practitioner
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.CreatePractitioner(c.Request().Context(), &p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/Practitioner/"+p.FHIRID)
	return c.JSON(http.StatusCreated, p.ToFHIR())
}

func (h *Handler) UpdatePractitionerFHIR(c echo.Context) error {
	existing, err := h.svc.GetPractitionerByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Practitioner", c.Param("id")))
	}
	var p Practitioner
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	p.ID = existing.ID
	p.FHIRID = existing.FHIRID
	if err := h.svc.UpdatePractitioner(c.Request().Context(), &p); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, p.ToFHIR())
}

func (h *Handler) DeletePractitionerFHIR(c echo.Context) error {
	p, err := h.svc.GetPractitionerByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("Practitioner", c.Param("id")))
	}
	if err := h.svc.DeletePractitioner(c.Request().Context(), p.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}
