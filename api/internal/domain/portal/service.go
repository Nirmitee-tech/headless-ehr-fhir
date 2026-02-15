package portal

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	accounts      PortalAccountRepository
	messages      PortalMessageRepository
	questionnaires QuestionnaireRepository
	responses     QuestionnaireResponseRepository
	checkins      PatientCheckinRepository
	vt            *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(
	accounts PortalAccountRepository,
	messages PortalMessageRepository,
	questionnaires QuestionnaireRepository,
	responses QuestionnaireResponseRepository,
	checkins PatientCheckinRepository,
) *Service {
	return &Service{
		accounts:       accounts,
		messages:       messages,
		questionnaires: questionnaires,
		responses:      responses,
		checkins:       checkins,
	}
}

// -- Portal Account --

var validAccountStatuses = map[string]bool{
	"active": true, "inactive": true, "locked": true,
	"pending-activation": true, "suspended": true,
}

func (s *Service) CreatePortalAccount(ctx context.Context, a *PortalAccount) error {
	if a.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if a.Status == "" {
		a.Status = "pending-activation"
	}
	if !validAccountStatuses[a.Status] {
		return fmt.Errorf("invalid status: %s", a.Status)
	}
	return s.accounts.Create(ctx, a)
}

func (s *Service) GetPortalAccount(ctx context.Context, id uuid.UUID) (*PortalAccount, error) {
	return s.accounts.GetByID(ctx, id)
}

func (s *Service) UpdatePortalAccount(ctx context.Context, a *PortalAccount) error {
	if a.Status != "" && !validAccountStatuses[a.Status] {
		return fmt.Errorf("invalid status: %s", a.Status)
	}
	return s.accounts.Update(ctx, a)
}

func (s *Service) DeletePortalAccount(ctx context.Context, id uuid.UUID) error {
	return s.accounts.Delete(ctx, id)
}

func (s *Service) ListPortalAccounts(ctx context.Context, limit, offset int) ([]*PortalAccount, int, error) {
	return s.accounts.List(ctx, limit, offset)
}

func (s *Service) ListPortalAccountsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PortalAccount, int, error) {
	return s.accounts.ListByPatient(ctx, patientID, limit, offset)
}

// -- Portal Message --

var validMessageStatuses = map[string]bool{
	"sent": true, "delivered": true, "read": true,
	"replied": true, "closed": true, "archived": true,
}

func (s *Service) CreatePortalMessage(ctx context.Context, m *PortalMessage) error {
	if m.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if m.Body == "" {
		return fmt.Errorf("body is required")
	}
	if m.Status == "" {
		m.Status = "sent"
	}
	if !validMessageStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	return s.messages.Create(ctx, m)
}

func (s *Service) GetPortalMessage(ctx context.Context, id uuid.UUID) (*PortalMessage, error) {
	return s.messages.GetByID(ctx, id)
}

func (s *Service) UpdatePortalMessage(ctx context.Context, m *PortalMessage) error {
	if m.Status != "" && !validMessageStatuses[m.Status] {
		return fmt.Errorf("invalid status: %s", m.Status)
	}
	return s.messages.Update(ctx, m)
}

func (s *Service) DeletePortalMessage(ctx context.Context, id uuid.UUID) error {
	return s.messages.Delete(ctx, id)
}

func (s *Service) ListPortalMessages(ctx context.Context, limit, offset int) ([]*PortalMessage, int, error) {
	return s.messages.List(ctx, limit, offset)
}

func (s *Service) ListPortalMessagesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PortalMessage, int, error) {
	return s.messages.ListByPatient(ctx, patientID, limit, offset)
}

// -- Questionnaire --

var validQuestionnaireStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

func (s *Service) CreateQuestionnaire(ctx context.Context, q *Questionnaire) error {
	if q.Name == "" {
		return fmt.Errorf("name is required")
	}
	if q.Status == "" {
		q.Status = "draft"
	}
	if !validQuestionnaireStatuses[q.Status] {
		return fmt.Errorf("invalid status: %s", q.Status)
	}
	if err := s.questionnaires.Create(ctx, q); err != nil {
		return err
	}
	q.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Questionnaire", q.FHIRID, q.ToFHIR())
	}
	return nil
}

func (s *Service) GetQuestionnaire(ctx context.Context, id uuid.UUID) (*Questionnaire, error) {
	return s.questionnaires.GetByID(ctx, id)
}

func (s *Service) GetQuestionnaireByFHIRID(ctx context.Context, fhirID string) (*Questionnaire, error) {
	return s.questionnaires.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateQuestionnaire(ctx context.Context, q *Questionnaire) error {
	if q.Status != "" && !validQuestionnaireStatuses[q.Status] {
		return fmt.Errorf("invalid status: %s", q.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Questionnaire", q.FHIRID, q.VersionID, q.ToFHIR())
		if err == nil {
			q.VersionID = newVer
		}
	}
	return s.questionnaires.Update(ctx, q)
}

func (s *Service) DeleteQuestionnaire(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		q, err := s.questionnaires.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Questionnaire", q.FHIRID, q.VersionID)
		}
	}
	return s.questionnaires.Delete(ctx, id)
}

func (s *Service) ListQuestionnaires(ctx context.Context, limit, offset int) ([]*Questionnaire, int, error) {
	return s.questionnaires.List(ctx, limit, offset)
}

func (s *Service) SearchQuestionnaires(ctx context.Context, params map[string]string, limit, offset int) ([]*Questionnaire, int, error) {
	return s.questionnaires.Search(ctx, params, limit, offset)
}

func (s *Service) AddQuestionnaireItem(ctx context.Context, item *QuestionnaireItem) error {
	if item.QuestionnaireID == uuid.Nil {
		return fmt.Errorf("questionnaire_id is required")
	}
	if item.LinkID == "" {
		return fmt.Errorf("link_id is required")
	}
	if item.Text == "" {
		return fmt.Errorf("text is required")
	}
	return s.questionnaires.AddItem(ctx, item)
}

func (s *Service) GetQuestionnaireItems(ctx context.Context, questionnaireID uuid.UUID) ([]*QuestionnaireItem, error) {
	return s.questionnaires.GetItems(ctx, questionnaireID)
}

// -- Questionnaire Response --

var validQRStatuses = map[string]bool{
	"in-progress": true, "completed": true, "amended": true,
	"entered-in-error": true, "stopped": true,
}

func (s *Service) CreateQuestionnaireResponse(ctx context.Context, qr *QuestionnaireResponse) error {
	if qr.QuestionnaireID == uuid.Nil {
		return fmt.Errorf("questionnaire_id is required")
	}
	if qr.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if qr.Status == "" {
		qr.Status = "in-progress"
	}
	if !validQRStatuses[qr.Status] {
		return fmt.Errorf("invalid status: %s", qr.Status)
	}
	if err := s.responses.Create(ctx, qr); err != nil {
		return err
	}
	qr.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "QuestionnaireResponse", qr.FHIRID, qr.ToFHIR())
	}
	return nil
}

func (s *Service) GetQuestionnaireResponse(ctx context.Context, id uuid.UUID) (*QuestionnaireResponse, error) {
	return s.responses.GetByID(ctx, id)
}

func (s *Service) GetQuestionnaireResponseByFHIRID(ctx context.Context, fhirID string) (*QuestionnaireResponse, error) {
	return s.responses.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateQuestionnaireResponse(ctx context.Context, qr *QuestionnaireResponse) error {
	if qr.Status != "" && !validQRStatuses[qr.Status] {
		return fmt.Errorf("invalid status: %s", qr.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "QuestionnaireResponse", qr.FHIRID, qr.VersionID, qr.ToFHIR())
		if err == nil {
			qr.VersionID = newVer
		}
	}
	return s.responses.Update(ctx, qr)
}

func (s *Service) DeleteQuestionnaireResponse(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		qr, err := s.responses.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "QuestionnaireResponse", qr.FHIRID, qr.VersionID)
		}
	}
	return s.responses.Delete(ctx, id)
}

func (s *Service) ListQuestionnaireResponses(ctx context.Context, limit, offset int) ([]*QuestionnaireResponse, int, error) {
	return s.responses.List(ctx, limit, offset)
}

func (s *Service) ListQuestionnaireResponsesByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*QuestionnaireResponse, int, error) {
	return s.responses.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) SearchQuestionnaireResponses(ctx context.Context, params map[string]string, limit, offset int) ([]*QuestionnaireResponse, int, error) {
	return s.responses.Search(ctx, params, limit, offset)
}

func (s *Service) AddQuestionnaireResponseItem(ctx context.Context, item *QuestionnaireResponseItem) error {
	if item.ResponseID == uuid.Nil {
		return fmt.Errorf("response_id is required")
	}
	if item.LinkID == "" {
		return fmt.Errorf("link_id is required")
	}
	return s.responses.AddResponseItem(ctx, item)
}

func (s *Service) GetQuestionnaireResponseItems(ctx context.Context, responseID uuid.UUID) ([]*QuestionnaireResponseItem, error) {
	return s.responses.GetResponseItems(ctx, responseID)
}

// -- Patient Checkin --

func (s *Service) CreatePatientCheckin(ctx context.Context, c *PatientCheckin) error {
	if c.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if c.Status == "" {
		c.Status = "pending"
	}
	return s.checkins.Create(ctx, c)
}

func (s *Service) GetPatientCheckin(ctx context.Context, id uuid.UUID) (*PatientCheckin, error) {
	return s.checkins.GetByID(ctx, id)
}

func (s *Service) UpdatePatientCheckin(ctx context.Context, c *PatientCheckin) error {
	return s.checkins.Update(ctx, c)
}

func (s *Service) DeletePatientCheckin(ctx context.Context, id uuid.UUID) error {
	return s.checkins.Delete(ctx, id)
}

func (s *Service) ListPatientCheckins(ctx context.Context, limit, offset int) ([]*PatientCheckin, int, error) {
	return s.checkins.List(ctx, limit, offset)
}

func (s *Service) ListPatientCheckinsByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PatientCheckin, int, error) {
	return s.checkins.ListByPatient(ctx, patientID, limit, offset)
}
