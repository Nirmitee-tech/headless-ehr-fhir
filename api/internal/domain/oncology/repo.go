package oncology

import (
	"context"

	"github.com/google/uuid"
)

type CancerDiagnosisRepository interface {
	Create(ctx context.Context, d *CancerDiagnosis) error
	GetByID(ctx context.Context, id uuid.UUID) (*CancerDiagnosis, error)
	Update(ctx context.Context, d *CancerDiagnosis) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*CancerDiagnosis, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CancerDiagnosis, int, error)
}

type TreatmentProtocolRepository interface {
	Create(ctx context.Context, p *TreatmentProtocol) error
	GetByID(ctx context.Context, id uuid.UUID) (*TreatmentProtocol, error)
	Update(ctx context.Context, p *TreatmentProtocol) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*TreatmentProtocol, int, error)
	// Drugs
	AddDrug(ctx context.Context, d *TreatmentProtocolDrug) error
	GetDrugs(ctx context.Context, protocolID uuid.UUID) ([]*TreatmentProtocolDrug, error)
}

type ChemoCycleRepository interface {
	Create(ctx context.Context, c *ChemoCycle) error
	GetByID(ctx context.Context, id uuid.UUID) (*ChemoCycle, error)
	Update(ctx context.Context, c *ChemoCycle) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ChemoCycle, int, error)
	// Administrations
	AddAdministration(ctx context.Context, a *ChemoAdministration) error
	GetAdministrations(ctx context.Context, cycleID uuid.UUID) ([]*ChemoAdministration, error)
}

type RadiationTherapyRepository interface {
	Create(ctx context.Context, r *RadiationTherapy) error
	GetByID(ctx context.Context, id uuid.UUID) (*RadiationTherapy, error)
	Update(ctx context.Context, r *RadiationTherapy) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*RadiationTherapy, int, error)
	// Sessions
	AddSession(ctx context.Context, s *RadiationSession) error
	GetSessions(ctx context.Context, radiationID uuid.UUID) ([]*RadiationSession, error)
}

type TumorMarkerRepository interface {
	Create(ctx context.Context, m *TumorMarker) error
	GetByID(ctx context.Context, id uuid.UUID) (*TumorMarker, error)
	Update(ctx context.Context, m *TumorMarker) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*TumorMarker, int, error)
}

type TumorBoardRepository interface {
	Create(ctx context.Context, r *TumorBoardReview) error
	GetByID(ctx context.Context, id uuid.UUID) (*TumorBoardReview, error)
	Update(ctx context.Context, r *TumorBoardReview) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*TumorBoardReview, int, error)
}
