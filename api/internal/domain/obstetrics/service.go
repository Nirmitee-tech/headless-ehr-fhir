package obstetrics

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	pregnancies    PregnancyRepository
	prenatalVisits PrenatalVisitRepository
	labors         LaborRepository
	deliveries     DeliveryRepository
	newborns       NewbornRepository
	postpartum     PostpartumRepository
}

func NewService(
	pregnancies PregnancyRepository,
	prenatalVisits PrenatalVisitRepository,
	labors LaborRepository,
	deliveries DeliveryRepository,
	newborns NewbornRepository,
	postpartum PostpartumRepository,
) *Service {
	return &Service{
		pregnancies:    pregnancies,
		prenatalVisits: prenatalVisits,
		labors:         labors,
		deliveries:     deliveries,
		newborns:       newborns,
		postpartum:     postpartum,
	}
}

// -- Pregnancy --

var validPregnancyStatuses = map[string]bool{
	"active": true, "completed": true, "ectopic": true, "molar": true,
	"miscarriage": true, "stillbirth": true, "terminated": true, "unknown": true,
}

func (s *Service) CreatePregnancy(ctx context.Context, p *Pregnancy) error {
	if p.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if p.Status == "" {
		p.Status = "active"
	}
	if !validPregnancyStatuses[p.Status] {
		return fmt.Errorf("invalid pregnancy status: %s", p.Status)
	}
	return s.pregnancies.Create(ctx, p)
}

func (s *Service) GetPregnancy(ctx context.Context, id uuid.UUID) (*Pregnancy, error) {
	return s.pregnancies.GetByID(ctx, id)
}

func (s *Service) UpdatePregnancy(ctx context.Context, p *Pregnancy) error {
	if p.Status != "" && !validPregnancyStatuses[p.Status] {
		return fmt.Errorf("invalid pregnancy status: %s", p.Status)
	}
	return s.pregnancies.Update(ctx, p)
}

func (s *Service) DeletePregnancy(ctx context.Context, id uuid.UUID) error {
	return s.pregnancies.Delete(ctx, id)
}

func (s *Service) ListPregnancies(ctx context.Context, limit, offset int) ([]*Pregnancy, int, error) {
	return s.pregnancies.List(ctx, limit, offset)
}

func (s *Service) ListPregnanciesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Pregnancy, int, error) {
	return s.pregnancies.ListByPatient(ctx, patientID, limit, offset)
}

// -- Prenatal Visit --

func (s *Service) CreatePrenatalVisit(ctx context.Context, v *PrenatalVisit) error {
	if v.PregnancyID == uuid.Nil {
		return fmt.Errorf("pregnancy_id is required")
	}
	if v.VisitDate.IsZero() {
		v.VisitDate = time.Now()
	}
	return s.prenatalVisits.Create(ctx, v)
}

func (s *Service) GetPrenatalVisit(ctx context.Context, id uuid.UUID) (*PrenatalVisit, error) {
	return s.prenatalVisits.GetByID(ctx, id)
}

func (s *Service) UpdatePrenatalVisit(ctx context.Context, v *PrenatalVisit) error {
	return s.prenatalVisits.Update(ctx, v)
}

func (s *Service) DeletePrenatalVisit(ctx context.Context, id uuid.UUID) error {
	return s.prenatalVisits.Delete(ctx, id)
}

func (s *Service) ListPrenatalVisitsByPregnancy(ctx context.Context, pregnancyID uuid.UUID, limit, offset int) ([]*PrenatalVisit, int, error) {
	return s.prenatalVisits.ListByPregnancy(ctx, pregnancyID, limit, offset)
}

// -- Labor Record --

func (s *Service) CreateLaborRecord(ctx context.Context, l *LaborRecord) error {
	if l.PregnancyID == uuid.Nil {
		return fmt.Errorf("pregnancy_id is required")
	}
	if l.Status == "" {
		l.Status = "active"
	}
	return s.labors.Create(ctx, l)
}

func (s *Service) GetLaborRecord(ctx context.Context, id uuid.UUID) (*LaborRecord, error) {
	return s.labors.GetByID(ctx, id)
}

func (s *Service) UpdateLaborRecord(ctx context.Context, l *LaborRecord) error {
	return s.labors.Update(ctx, l)
}

func (s *Service) DeleteLaborRecord(ctx context.Context, id uuid.UUID) error {
	return s.labors.Delete(ctx, id)
}

func (s *Service) ListLaborRecords(ctx context.Context, limit, offset int) ([]*LaborRecord, int, error) {
	return s.labors.List(ctx, limit, offset)
}

func (s *Service) AddCervicalExam(ctx context.Context, e *LaborCervicalExam) error {
	if e.LaborRecordID == uuid.Nil {
		return fmt.Errorf("labor_record_id is required")
	}
	if e.ExamDatetime.IsZero() {
		e.ExamDatetime = time.Now()
	}
	return s.labors.AddCervicalExam(ctx, e)
}

func (s *Service) GetCervicalExams(ctx context.Context, laborRecordID uuid.UUID) ([]*LaborCervicalExam, error) {
	return s.labors.GetCervicalExams(ctx, laborRecordID)
}

func (s *Service) AddFetalMonitoring(ctx context.Context, f *FetalMonitoring) error {
	if f.LaborRecordID == uuid.Nil {
		return fmt.Errorf("labor_record_id is required")
	}
	if f.MonitoringDatetime.IsZero() {
		f.MonitoringDatetime = time.Now()
	}
	return s.labors.AddFetalMonitoring(ctx, f)
}

func (s *Service) GetFetalMonitoring(ctx context.Context, laborRecordID uuid.UUID) ([]*FetalMonitoring, error) {
	return s.labors.GetFetalMonitoring(ctx, laborRecordID)
}

// -- Delivery Record --

func (s *Service) CreateDelivery(ctx context.Context, d *DeliveryRecord) error {
	if d.PregnancyID == uuid.Nil {
		return fmt.Errorf("pregnancy_id is required")
	}
	if d.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if d.DeliveryDatetime.IsZero() {
		return fmt.Errorf("delivery_datetime is required")
	}
	if d.DeliveryMethod == "" {
		return fmt.Errorf("delivery_method is required")
	}
	if d.DeliveringProviderID == uuid.Nil {
		return fmt.Errorf("delivering_provider_id is required")
	}
	return s.deliveries.Create(ctx, d)
}

func (s *Service) GetDelivery(ctx context.Context, id uuid.UUID) (*DeliveryRecord, error) {
	return s.deliveries.GetByID(ctx, id)
}

func (s *Service) UpdateDelivery(ctx context.Context, d *DeliveryRecord) error {
	return s.deliveries.Update(ctx, d)
}

func (s *Service) DeleteDelivery(ctx context.Context, id uuid.UUID) error {
	return s.deliveries.Delete(ctx, id)
}

func (s *Service) ListDeliveriesByPregnancy(ctx context.Context, pregnancyID uuid.UUID, limit, offset int) ([]*DeliveryRecord, int, error) {
	return s.deliveries.ListByPregnancy(ctx, pregnancyID, limit, offset)
}

// -- Newborn Record --

func (s *Service) CreateNewborn(ctx context.Context, n *NewbornRecord) error {
	if n.DeliveryID == uuid.Nil {
		return fmt.Errorf("delivery_id is required")
	}
	if n.BirthDatetime.IsZero() {
		return fmt.Errorf("birth_datetime is required")
	}
	return s.newborns.Create(ctx, n)
}

func (s *Service) GetNewborn(ctx context.Context, id uuid.UUID) (*NewbornRecord, error) {
	return s.newborns.GetByID(ctx, id)
}

func (s *Service) UpdateNewborn(ctx context.Context, n *NewbornRecord) error {
	return s.newborns.Update(ctx, n)
}

func (s *Service) DeleteNewborn(ctx context.Context, id uuid.UUID) error {
	return s.newborns.Delete(ctx, id)
}

func (s *Service) ListNewborns(ctx context.Context, limit, offset int) ([]*NewbornRecord, int, error) {
	return s.newborns.List(ctx, limit, offset)
}

// -- Postpartum Record --

func (s *Service) CreatePostpartum(ctx context.Context, p *PostpartumRecord) error {
	if p.PregnancyID == uuid.Nil {
		return fmt.Errorf("pregnancy_id is required")
	}
	if p.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if p.VisitDate.IsZero() {
		p.VisitDate = time.Now()
	}
	return s.postpartum.Create(ctx, p)
}

func (s *Service) GetPostpartum(ctx context.Context, id uuid.UUID) (*PostpartumRecord, error) {
	return s.postpartum.GetByID(ctx, id)
}

func (s *Service) UpdatePostpartum(ctx context.Context, p *PostpartumRecord) error {
	return s.postpartum.Update(ctx, p)
}

func (s *Service) DeletePostpartum(ctx context.Context, id uuid.UUID) error {
	return s.postpartum.Delete(ctx, id)
}

func (s *Service) ListPostpartumRecords(ctx context.Context, limit, offset int) ([]*PostpartumRecord, int, error) {
	return s.postpartum.List(ctx, limit, offset)
}
