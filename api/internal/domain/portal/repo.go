package portal

import (
	"context"

	"github.com/google/uuid"
)

type PortalAccountRepository interface {
	Create(ctx context.Context, a *PortalAccount) error
	GetByID(ctx context.Context, id uuid.UUID) (*PortalAccount, error)
	Update(ctx context.Context, a *PortalAccount) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*PortalAccount, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PortalAccount, int, error)
}

type PortalMessageRepository interface {
	Create(ctx context.Context, m *PortalMessage) error
	GetByID(ctx context.Context, id uuid.UUID) (*PortalMessage, error)
	Update(ctx context.Context, m *PortalMessage) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*PortalMessage, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PortalMessage, int, error)
}

type QuestionnaireRepository interface {
	Create(ctx context.Context, q *Questionnaire) error
	GetByID(ctx context.Context, id uuid.UUID) (*Questionnaire, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Questionnaire, error)
	Update(ctx context.Context, q *Questionnaire) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Questionnaire, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Questionnaire, int, error)
	// Items
	AddItem(ctx context.Context, item *QuestionnaireItem) error
	GetItems(ctx context.Context, questionnaireID uuid.UUID) ([]*QuestionnaireItem, error)
}

type QuestionnaireResponseRepository interface {
	Create(ctx context.Context, qr *QuestionnaireResponse) error
	GetByID(ctx context.Context, id uuid.UUID) (*QuestionnaireResponse, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*QuestionnaireResponse, error)
	Update(ctx context.Context, qr *QuestionnaireResponse) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*QuestionnaireResponse, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*QuestionnaireResponse, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*QuestionnaireResponse, int, error)
	// Response Items
	AddResponseItem(ctx context.Context, item *QuestionnaireResponseItem) error
	GetResponseItems(ctx context.Context, responseID uuid.UUID) ([]*QuestionnaireResponseItem, error)
}

type PatientCheckinRepository interface {
	Create(ctx context.Context, c *PatientCheckin) error
	GetByID(ctx context.Context, id uuid.UUID) (*PatientCheckin, error)
	Update(ctx context.Context, c *PatientCheckin) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*PatientCheckin, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*PatientCheckin, int, error)
}
