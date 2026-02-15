package conformance

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	namingSystems        NamingSystemRepository
	operationDefinitions OperationDefinitionRepository
	messageDefinitions   MessageDefinitionRepository
	messageHeaders       MessageHeaderRepository
	vt                   *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(ns NamingSystemRepository, od OperationDefinitionRepository, md MessageDefinitionRepository, mh MessageHeaderRepository) *Service {
	return &Service{namingSystems: ns, operationDefinitions: od, messageDefinitions: md, messageHeaders: mh}
}

// -- NamingSystem --

var validConformanceStatuses = map[string]bool{
	"draft": true, "active": true, "retired": true, "unknown": true,
}

var validNamingSystemKinds = map[string]bool{
	"codesystem": true, "identifier": true, "root": true,
}

func (s *Service) CreateNamingSystem(ctx context.Context, ns *NamingSystem) error {
	if ns.Name == "" {
		return fmt.Errorf("name is required")
	}
	if ns.Kind == "" {
		ns.Kind = "identifier"
	}
	if !validNamingSystemKinds[ns.Kind] {
		return fmt.Errorf("invalid kind: %s", ns.Kind)
	}
	if ns.Status == "" {
		ns.Status = "draft"
	}
	if !validConformanceStatuses[ns.Status] {
		return fmt.Errorf("invalid status: %s", ns.Status)
	}
	if err := s.namingSystems.Create(ctx, ns); err != nil {
		return err
	}
	ns.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "NamingSystem", ns.FHIRID, ns.ToFHIR())
	}
	return nil
}

func (s *Service) GetNamingSystem(ctx context.Context, id uuid.UUID) (*NamingSystem, error) {
	return s.namingSystems.GetByID(ctx, id)
}

func (s *Service) GetNamingSystemByFHIRID(ctx context.Context, fhirID string) (*NamingSystem, error) {
	return s.namingSystems.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateNamingSystem(ctx context.Context, ns *NamingSystem) error {
	if ns.Status != "" && !validConformanceStatuses[ns.Status] {
		return fmt.Errorf("invalid status: %s", ns.Status)
	}
	if ns.Kind != "" && !validNamingSystemKinds[ns.Kind] {
		return fmt.Errorf("invalid kind: %s", ns.Kind)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "NamingSystem", ns.FHIRID, ns.VersionID, ns.ToFHIR())
		if err == nil {
			ns.VersionID = newVer
		}
	}
	return s.namingSystems.Update(ctx, ns)
}

func (s *Service) DeleteNamingSystem(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		ns, err := s.namingSystems.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "NamingSystem", ns.FHIRID, ns.VersionID)
		}
	}
	return s.namingSystems.Delete(ctx, id)
}

func (s *Service) ListNamingSystems(ctx context.Context, limit, offset int) ([]*NamingSystem, int, error) {
	return s.namingSystems.List(ctx, limit, offset)
}

func (s *Service) SearchNamingSystems(ctx context.Context, params map[string]string, limit, offset int) ([]*NamingSystem, int, error) {
	return s.namingSystems.Search(ctx, params, limit, offset)
}

func (s *Service) AddNamingSystemUniqueID(ctx context.Context, uid *NamingSystemUniqueID) error {
	if uid.NamingSystemID == uuid.Nil {
		return fmt.Errorf("naming_system_id is required")
	}
	if uid.Type == "" {
		return fmt.Errorf("type is required")
	}
	if uid.Value == "" {
		return fmt.Errorf("value is required")
	}
	return s.namingSystems.AddUniqueID(ctx, uid)
}

func (s *Service) GetNamingSystemUniqueIDs(ctx context.Context, namingSystemID uuid.UUID) ([]*NamingSystemUniqueID, error) {
	return s.namingSystems.GetUniqueIDs(ctx, namingSystemID)
}

// -- OperationDefinition --

var validOpDefKinds = map[string]bool{
	"operation": true, "query": true,
}

func (s *Service) CreateOperationDefinition(ctx context.Context, od *OperationDefinition) error {
	if od.Name == "" {
		return fmt.Errorf("name is required")
	}
	if od.Code == "" {
		return fmt.Errorf("code is required")
	}
	if od.Kind == "" {
		od.Kind = "operation"
	}
	if !validOpDefKinds[od.Kind] {
		return fmt.Errorf("invalid kind: %s", od.Kind)
	}
	if od.Status == "" {
		od.Status = "draft"
	}
	if !validConformanceStatuses[od.Status] {
		return fmt.Errorf("invalid status: %s", od.Status)
	}
	if err := s.operationDefinitions.Create(ctx, od); err != nil {
		return err
	}
	od.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "OperationDefinition", od.FHIRID, od.ToFHIR())
	}
	return nil
}

func (s *Service) GetOperationDefinition(ctx context.Context, id uuid.UUID) (*OperationDefinition, error) {
	return s.operationDefinitions.GetByID(ctx, id)
}

func (s *Service) GetOperationDefinitionByFHIRID(ctx context.Context, fhirID string) (*OperationDefinition, error) {
	return s.operationDefinitions.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateOperationDefinition(ctx context.Context, od *OperationDefinition) error {
	if od.Status != "" && !validConformanceStatuses[od.Status] {
		return fmt.Errorf("invalid status: %s", od.Status)
	}
	if od.Kind != "" && !validOpDefKinds[od.Kind] {
		return fmt.Errorf("invalid kind: %s", od.Kind)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "OperationDefinition", od.FHIRID, od.VersionID, od.ToFHIR())
		if err == nil {
			od.VersionID = newVer
		}
	}
	return s.operationDefinitions.Update(ctx, od)
}

func (s *Service) DeleteOperationDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		od, err := s.operationDefinitions.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "OperationDefinition", od.FHIRID, od.VersionID)
		}
	}
	return s.operationDefinitions.Delete(ctx, id)
}

func (s *Service) ListOperationDefinitions(ctx context.Context, limit, offset int) ([]*OperationDefinition, int, error) {
	return s.operationDefinitions.List(ctx, limit, offset)
}

func (s *Service) SearchOperationDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*OperationDefinition, int, error) {
	return s.operationDefinitions.Search(ctx, params, limit, offset)
}

func (s *Service) AddOperationDefinitionParameter(ctx context.Context, p *OperationDefinitionParameter) error {
	if p.OperationDefinitionID == uuid.Nil {
		return fmt.Errorf("operation_definition_id is required")
	}
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.Use == "" {
		return fmt.Errorf("use is required")
	}
	return s.operationDefinitions.AddParameter(ctx, p)
}

func (s *Service) GetOperationDefinitionParameters(ctx context.Context, opDefID uuid.UUID) ([]*OperationDefinitionParameter, error) {
	return s.operationDefinitions.GetParameters(ctx, opDefID)
}

// -- MessageDefinition --

func (s *Service) CreateMessageDefinition(ctx context.Context, md *MessageDefinition) error {
	if md.EventCodingCode == "" {
		return fmt.Errorf("event_coding_code is required")
	}
	if md.Status == "" {
		md.Status = "draft"
	}
	if !validConformanceStatuses[md.Status] {
		return fmt.Errorf("invalid status: %s", md.Status)
	}
	if err := s.messageDefinitions.Create(ctx, md); err != nil {
		return err
	}
	md.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "MessageDefinition", md.FHIRID, md.ToFHIR())
	}
	return nil
}

func (s *Service) GetMessageDefinition(ctx context.Context, id uuid.UUID) (*MessageDefinition, error) {
	return s.messageDefinitions.GetByID(ctx, id)
}

func (s *Service) GetMessageDefinitionByFHIRID(ctx context.Context, fhirID string) (*MessageDefinition, error) {
	return s.messageDefinitions.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMessageDefinition(ctx context.Context, md *MessageDefinition) error {
	if md.Status != "" && !validConformanceStatuses[md.Status] {
		return fmt.Errorf("invalid status: %s", md.Status)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "MessageDefinition", md.FHIRID, md.VersionID, md.ToFHIR())
		if err == nil {
			md.VersionID = newVer
		}
	}
	return s.messageDefinitions.Update(ctx, md)
}

func (s *Service) DeleteMessageDefinition(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		md, err := s.messageDefinitions.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "MessageDefinition", md.FHIRID, md.VersionID)
		}
	}
	return s.messageDefinitions.Delete(ctx, id)
}

func (s *Service) ListMessageDefinitions(ctx context.Context, limit, offset int) ([]*MessageDefinition, int, error) {
	return s.messageDefinitions.List(ctx, limit, offset)
}

func (s *Service) SearchMessageDefinitions(ctx context.Context, params map[string]string, limit, offset int) ([]*MessageDefinition, int, error) {
	return s.messageDefinitions.Search(ctx, params, limit, offset)
}

// -- MessageHeader --

func (s *Service) CreateMessageHeader(ctx context.Context, mh *MessageHeader) error {
	if mh.EventCodingCode == "" {
		return fmt.Errorf("event_coding_code is required")
	}
	if mh.SourceEndpoint == "" {
		return fmt.Errorf("source_endpoint is required")
	}
	if err := s.messageHeaders.Create(ctx, mh); err != nil {
		return err
	}
	mh.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "MessageHeader", mh.FHIRID, mh.ToFHIR())
	}
	return nil
}

func (s *Service) GetMessageHeader(ctx context.Context, id uuid.UUID) (*MessageHeader, error) {
	return s.messageHeaders.GetByID(ctx, id)
}

func (s *Service) GetMessageHeaderByFHIRID(ctx context.Context, fhirID string) (*MessageHeader, error) {
	return s.messageHeaders.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateMessageHeader(ctx context.Context, mh *MessageHeader) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "MessageHeader", mh.FHIRID, mh.VersionID, mh.ToFHIR())
		if err == nil {
			mh.VersionID = newVer
		}
	}
	return s.messageHeaders.Update(ctx, mh)
}

func (s *Service) DeleteMessageHeader(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		mh, err := s.messageHeaders.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "MessageHeader", mh.FHIRID, mh.VersionID)
		}
	}
	return s.messageHeaders.Delete(ctx, id)
}

func (s *Service) ListMessageHeaders(ctx context.Context, limit, offset int) ([]*MessageHeader, int, error) {
	return s.messageHeaders.List(ctx, limit, offset)
}

func (s *Service) SearchMessageHeaders(ctx context.Context, params map[string]string, limit, offset int) ([]*MessageHeader, int, error) {
	return s.messageHeaders.Search(ctx, params, limit, offset)
}
