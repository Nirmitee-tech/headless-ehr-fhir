package inbox

import (
	"time"

	"github.com/google/uuid"
)

// MessagePool represents a routing pool for InBasket messages.
type MessagePool struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	PoolName       string     `db:"pool_name" json:"pool_name"`
	PoolType       string     `db:"pool_type" json:"pool_type"`
	OrganizationID *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
	DepartmentID   *uuid.UUID `db:"department_id" json:"department_id,omitempty"`
	Description    *string    `db:"description" json:"description,omitempty"`
	IsActive       bool       `db:"is_active" json:"is_active"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

// MessagePoolMember represents membership in a message pool.
type MessagePoolMember struct {
	ID       uuid.UUID  `db:"id" json:"id"`
	PoolID   uuid.UUID  `db:"pool_id" json:"pool_id"`
	UserID   uuid.UUID  `db:"user_id" json:"user_id"`
	Role     *string    `db:"role" json:"role,omitempty"`
	IsActive bool       `db:"is_active" json:"is_active"`
	JoinedAt time.Time  `db:"joined_at" json:"joined_at"`
}

// InboxMessage represents an InBasket message.
type InboxMessage struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	MessageType string     `db:"message_type" json:"message_type"`
	Priority    string     `db:"priority" json:"priority"`
	Subject     string     `db:"subject" json:"subject"`
	Body        *string    `db:"body" json:"body,omitempty"`
	PatientID   *uuid.UUID `db:"patient_id" json:"patient_id,omitempty"`
	EncounterID *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	SenderID    *uuid.UUID `db:"sender_id" json:"sender_id,omitempty"`
	RecipientID *uuid.UUID `db:"recipient_id" json:"recipient_id,omitempty"`
	PoolID      *uuid.UUID `db:"pool_id" json:"pool_id,omitempty"`
	Status      string     `db:"status" json:"status"`
	ParentID    *uuid.UUID `db:"parent_id" json:"parent_id,omitempty"`
	ThreadID    *uuid.UUID `db:"thread_id" json:"thread_id,omitempty"`
	IsUrgent    bool       `db:"is_urgent" json:"is_urgent"`
	DueDate     *time.Time `db:"due_date" json:"due_date,omitempty"`
	ReadAt      *time.Time `db:"read_at" json:"read_at,omitempty"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

// CosignRequest represents a cosigning workflow request.
type CosignRequest struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	DocumentType string     `db:"document_type" json:"document_type"`
	DocumentID   *uuid.UUID `db:"document_id" json:"document_id,omitempty"`
	RequesterID  uuid.UUID  `db:"requester_id" json:"requester_id"`
	CosignerID   uuid.UUID  `db:"cosigner_id" json:"cosigner_id"`
	PatientID    *uuid.UUID `db:"patient_id" json:"patient_id,omitempty"`
	EncounterID  *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	Status       string     `db:"status" json:"status"`
	Note         *string    `db:"note" json:"note,omitempty"`
	RequestedAt  time.Time  `db:"requested_at" json:"requested_at"`
	RespondedAt  *time.Time `db:"responded_at" json:"responded_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// PatientList represents a worklist.
type PatientList struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	ListName     string     `db:"list_name" json:"list_name"`
	ListType     string     `db:"list_type" json:"list_type"`
	OwnerID      uuid.UUID  `db:"owner_id" json:"owner_id"`
	DepartmentID *uuid.UUID `db:"department_id" json:"department_id,omitempty"`
	Description  *string    `db:"description" json:"description,omitempty"`
	AutoCriteria *string    `db:"auto_criteria" json:"auto_criteria,omitempty"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// PatientListMember represents a patient on a worklist.
type PatientListMember struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	ListID    uuid.UUID  `db:"list_id" json:"list_id"`
	PatientID uuid.UUID  `db:"patient_id" json:"patient_id"`
	Priority  int        `db:"priority" json:"priority"`
	Flags     *string    `db:"flags" json:"flags,omitempty"`
	OneLiner  *string    `db:"one_liner" json:"one_liner,omitempty"`
	AddedBy   *uuid.UUID `db:"added_by" json:"added_by,omitempty"`
	AddedAt   time.Time  `db:"added_at" json:"added_at"`
	RemovedAt *time.Time `db:"removed_at" json:"removed_at,omitempty"`
}

// HandoffRecord represents an I-PASS/SBAR handoff between providers.
type HandoffRecord struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	PatientID          uuid.UUID  `db:"patient_id" json:"patient_id"`
	EncounterID        *uuid.UUID `db:"encounter_id" json:"encounter_id,omitempty"`
	FromProviderID     uuid.UUID  `db:"from_provider_id" json:"from_provider_id"`
	ToProviderID       uuid.UUID  `db:"to_provider_id" json:"to_provider_id"`
	HandoffType        string     `db:"handoff_type" json:"handoff_type"`
	IllnessSeverity    *string    `db:"illness_severity" json:"illness_severity,omitempty"`
	PatientSummary     *string    `db:"patient_summary" json:"patient_summary,omitempty"`
	ActionList         *string    `db:"action_list" json:"action_list,omitempty"`
	SituationAwareness *string    `db:"situation_awareness" json:"situation_awareness,omitempty"`
	Synthesis          *string    `db:"synthesis" json:"synthesis,omitempty"`
	ContingencyPlan    *string    `db:"contingency_plan" json:"contingency_plan,omitempty"`
	Status             string     `db:"status" json:"status"`
	AcknowledgedAt     *time.Time `db:"acknowledged_at" json:"acknowledged_at,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}
