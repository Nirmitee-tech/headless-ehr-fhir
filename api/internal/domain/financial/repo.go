package financial

import (
	"context"

	"github.com/google/uuid"
)

type AccountRepository interface {
	Create(ctx context.Context, a *Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Account, error)
	Update(ctx context.Context, a *Account) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Account, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Account, int, error)
}

type InsurancePlanRepository interface {
	Create(ctx context.Context, ip *InsurancePlan) error
	GetByID(ctx context.Context, id uuid.UUID) (*InsurancePlan, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*InsurancePlan, error)
	Update(ctx context.Context, ip *InsurancePlan) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*InsurancePlan, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*InsurancePlan, int, error)
}

type PaymentNoticeRepository interface {
	Create(ctx context.Context, pn *PaymentNotice) error
	GetByID(ctx context.Context, id uuid.UUID) (*PaymentNotice, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*PaymentNotice, error)
	Update(ctx context.Context, pn *PaymentNotice) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*PaymentNotice, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*PaymentNotice, int, error)
}

type PaymentReconciliationRepository interface {
	Create(ctx context.Context, pr *PaymentReconciliation) error
	GetByID(ctx context.Context, id uuid.UUID) (*PaymentReconciliation, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*PaymentReconciliation, error)
	Update(ctx context.Context, pr *PaymentReconciliation) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*PaymentReconciliation, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*PaymentReconciliation, int, error)
}

type ChargeItemRepository interface {
	Create(ctx context.Context, ci *ChargeItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*ChargeItem, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ChargeItem, error)
	Update(ctx context.Context, ci *ChargeItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ChargeItem, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ChargeItem, int, error)
}

type ChargeItemDefinitionRepository interface {
	Create(ctx context.Context, cd *ChargeItemDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*ChargeItemDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ChargeItemDefinition, error)
	Update(ctx context.Context, cd *ChargeItemDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ChargeItemDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ChargeItemDefinition, int, error)
}

type ContractRepository interface {
	Create(ctx context.Context, ct *Contract) error
	GetByID(ctx context.Context, id uuid.UUID) (*Contract, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Contract, error)
	Update(ctx context.Context, ct *Contract) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Contract, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Contract, int, error)
}

type EnrollmentRequestRepository interface {
	Create(ctx context.Context, er *EnrollmentRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*EnrollmentRequest, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*EnrollmentRequest, error)
	Update(ctx context.Context, er *EnrollmentRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*EnrollmentRequest, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EnrollmentRequest, int, error)
}

type EnrollmentResponseRepository interface {
	Create(ctx context.Context, er *EnrollmentResponse) error
	GetByID(ctx context.Context, id uuid.UUID) (*EnrollmentResponse, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*EnrollmentResponse, error)
	Update(ctx context.Context, er *EnrollmentResponse) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*EnrollmentResponse, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EnrollmentResponse, int, error)
}
