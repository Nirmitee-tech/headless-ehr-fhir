package obstetrics

import (
	"context"

	"github.com/google/uuid"
)

type PregnancyRepository interface {
	Create(ctx context.Context, p *Pregnancy) error
	GetByID(ctx context.Context, id uuid.UUID) (*Pregnancy, error)
	Update(ctx context.Context, p *Pregnancy) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Pregnancy, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Pregnancy, int, error)
}

type PrenatalVisitRepository interface {
	Create(ctx context.Context, v *PrenatalVisit) error
	GetByID(ctx context.Context, id uuid.UUID) (*PrenatalVisit, error)
	Update(ctx context.Context, v *PrenatalVisit) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPregnancy(ctx context.Context, pregnancyID uuid.UUID, limit, offset int) ([]*PrenatalVisit, int, error)
}

type LaborRepository interface {
	Create(ctx context.Context, l *LaborRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*LaborRecord, error)
	Update(ctx context.Context, l *LaborRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*LaborRecord, int, error)
	// Cervical exams
	AddCervicalExam(ctx context.Context, e *LaborCervicalExam) error
	GetCervicalExams(ctx context.Context, laborRecordID uuid.UUID) ([]*LaborCervicalExam, error)
	// Fetal monitoring
	AddFetalMonitoring(ctx context.Context, f *FetalMonitoring) error
	GetFetalMonitoring(ctx context.Context, laborRecordID uuid.UUID) ([]*FetalMonitoring, error)
}

type DeliveryRepository interface {
	Create(ctx context.Context, d *DeliveryRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*DeliveryRecord, error)
	Update(ctx context.Context, d *DeliveryRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPregnancy(ctx context.Context, pregnancyID uuid.UUID, limit, offset int) ([]*DeliveryRecord, int, error)
}

type NewbornRepository interface {
	Create(ctx context.Context, n *NewbornRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*NewbornRecord, error)
	Update(ctx context.Context, n *NewbornRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*NewbornRecord, int, error)
}

type PostpartumRepository interface {
	Create(ctx context.Context, p *PostpartumRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*PostpartumRecord, error)
	Update(ctx context.Context, p *PostpartumRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*PostpartumRecord, int, error)
}
