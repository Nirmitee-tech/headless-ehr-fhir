package obstetrics

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/pkg/pagination"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(api *echo.Group) {
	// Read endpoints – admin, physician, nurse
	readGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	readGroup.GET("/pregnancies", h.ListPregnancies)
	readGroup.GET("/pregnancies/:id", h.GetPregnancy)
	readGroup.GET("/pregnancies/:id/prenatal-visits", h.ListPrenatalVisits)
	readGroup.GET("/prenatal-visits/:id", h.GetPrenatalVisit)
	readGroup.GET("/labor-records", h.ListLaborRecords)
	readGroup.GET("/labor-records/:id", h.GetLaborRecord)
	readGroup.GET("/labor-records/:id/cervical-exams", h.GetCervicalExams)
	readGroup.GET("/labor-records/:id/fetal-monitoring", h.GetFetalMonitoring)
	readGroup.GET("/deliveries/:id", h.GetDelivery)
	readGroup.GET("/newborns", h.ListNewborns)
	readGroup.GET("/newborns/:id", h.GetNewborn)
	readGroup.GET("/postpartum-records", h.ListPostpartumRecords)
	readGroup.GET("/postpartum-records/:id", h.GetPostpartum)

	// Write endpoints – admin, physician, nurse
	writeGroup := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	writeGroup.POST("/pregnancies", h.CreatePregnancy)
	writeGroup.PUT("/pregnancies/:id", h.UpdatePregnancy)
	writeGroup.DELETE("/pregnancies/:id", h.DeletePregnancy)
	writeGroup.POST("/pregnancies/:id/prenatal-visits", h.CreatePrenatalVisit)
	writeGroup.PUT("/prenatal-visits/:id", h.UpdatePrenatalVisit)
	writeGroup.DELETE("/prenatal-visits/:id", h.DeletePrenatalVisit)
	writeGroup.POST("/labor-records", h.CreateLaborRecord)
	writeGroup.PUT("/labor-records/:id", h.UpdateLaborRecord)
	writeGroup.DELETE("/labor-records/:id", h.DeleteLaborRecord)
	writeGroup.POST("/labor-records/:id/cervical-exams", h.AddCervicalExam)
	writeGroup.POST("/labor-records/:id/fetal-monitoring", h.AddFetalMonitoring)
	writeGroup.POST("/deliveries", h.CreateDelivery)
	writeGroup.PUT("/deliveries/:id", h.UpdateDelivery)
	writeGroup.DELETE("/deliveries/:id", h.DeleteDelivery)
	writeGroup.POST("/newborns", h.CreateNewborn)
	writeGroup.PUT("/newborns/:id", h.UpdateNewborn)
	writeGroup.DELETE("/newborns/:id", h.DeleteNewborn)
	writeGroup.POST("/postpartum-records", h.CreatePostpartum)
	writeGroup.PUT("/postpartum-records/:id", h.UpdatePostpartum)
	writeGroup.DELETE("/postpartum-records/:id", h.DeletePostpartum)
}

// -- Pregnancy Handlers --

func (h *Handler) CreatePregnancy(c echo.Context) error {
	var p Pregnancy
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePregnancy(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetPregnancy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := h.svc.GetPregnancy(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "pregnancy not found")
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) ListPregnancies(c echo.Context) error {
	pg := pagination.FromContext(c)
	if patientID := c.QueryParam("patient_id"); patientID != "" {
		pid, err := uuid.Parse(patientID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		items, total, err := h.svc.ListPregnanciesByPatient(c.Request().Context(), pid, pg.Limit, pg.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
	}
	items, total, err := h.svc.ListPregnancies(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePregnancy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p Pregnancy
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.ID = id
	if err := h.svc.UpdatePregnancy(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) DeletePregnancy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePregnancy(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Prenatal Visit Handlers --

func (h *Handler) CreatePrenatalVisit(c echo.Context) error {
	pregID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pregnancy id")
	}
	var v PrenatalVisit
	if err := c.Bind(&v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	v.PregnancyID = pregID
	if err := h.svc.CreatePrenatalVisit(c.Request().Context(), &v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, v)
}

func (h *Handler) GetPrenatalVisit(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	v, err := h.svc.GetPrenatalVisit(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "prenatal visit not found")
	}
	return c.JSON(http.StatusOK, v)
}

func (h *Handler) ListPrenatalVisits(c echo.Context) error {
	pregID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pregnancy id")
	}
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListPrenatalVisitsByPregnancy(c.Request().Context(), pregID, pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePrenatalVisit(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var v PrenatalVisit
	if err := c.Bind(&v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	v.ID = id
	if err := h.svc.UpdatePrenatalVisit(c.Request().Context(), &v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, v)
}

func (h *Handler) DeletePrenatalVisit(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePrenatalVisit(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Labor Record Handlers --

func (h *Handler) CreateLaborRecord(c echo.Context) error {
	var l LaborRecord
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateLaborRecord(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, l)
}

func (h *Handler) GetLaborRecord(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	l, err := h.svc.GetLaborRecord(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "labor record not found")
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) ListLaborRecords(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListLaborRecords(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateLaborRecord(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var l LaborRecord
	if err := c.Bind(&l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	l.ID = id
	if err := h.svc.UpdateLaborRecord(c.Request().Context(), &l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, l)
}

func (h *Handler) DeleteLaborRecord(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteLaborRecord(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AddCervicalExam(c echo.Context) error {
	laborID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var e LaborCervicalExam
	if err := c.Bind(&e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	e.LaborRecordID = laborID
	if err := h.svc.AddCervicalExam(c.Request().Context(), &e); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, e)
}

func (h *Handler) GetCervicalExams(c echo.Context) error {
	laborID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	exams, err := h.svc.GetCervicalExams(c.Request().Context(), laborID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, exams)
}

func (h *Handler) AddFetalMonitoring(c echo.Context) error {
	laborID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var f FetalMonitoring
	if err := c.Bind(&f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	f.LaborRecordID = laborID
	if err := h.svc.AddFetalMonitoring(c.Request().Context(), &f); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, f)
}

func (h *Handler) GetFetalMonitoring(c echo.Context) error {
	laborID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	items, err := h.svc.GetFetalMonitoring(c.Request().Context(), laborID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, items)
}

// -- Delivery Handlers --

func (h *Handler) CreateDelivery(c echo.Context) error {
	var d DeliveryRecord
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateDelivery(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, d)
}

func (h *Handler) GetDelivery(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	d, err := h.svc.GetDelivery(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "delivery record not found")
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) UpdateDelivery(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var d DeliveryRecord
	if err := c.Bind(&d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	d.ID = id
	if err := h.svc.UpdateDelivery(c.Request().Context(), &d); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, d)
}

func (h *Handler) DeleteDelivery(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteDelivery(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Newborn Handlers --

func (h *Handler) CreateNewborn(c echo.Context) error {
	var n NewbornRecord
	if err := c.Bind(&n); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreateNewborn(c.Request().Context(), &n); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, n)
}

func (h *Handler) GetNewborn(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	n, err := h.svc.GetNewborn(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "newborn record not found")
	}
	return c.JSON(http.StatusOK, n)
}

func (h *Handler) ListNewborns(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListNewborns(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdateNewborn(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var n NewbornRecord
	if err := c.Bind(&n); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	n.ID = id
	if err := h.svc.UpdateNewborn(c.Request().Context(), &n); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, n)
}

func (h *Handler) DeleteNewborn(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteNewborn(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// -- Postpartum Handlers --

func (h *Handler) CreatePostpartum(c echo.Context) error {
	var p PostpartumRecord
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.svc.CreatePostpartum(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetPostpartum(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := h.svc.GetPostpartum(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "postpartum record not found")
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) ListPostpartumRecords(c echo.Context) error {
	pg := pagination.FromContext(c)
	items, total, err := h.svc.ListPostpartumRecords(c.Request().Context(), pg.Limit, pg.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, pagination.NewResponse(items, total, pg.Limit, pg.Offset))
}

func (h *Handler) UpdatePostpartum(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p PostpartumRecord
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p.ID = id
	if err := h.svc.UpdatePostpartum(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Handler) DeletePostpartum(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeletePostpartum(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
