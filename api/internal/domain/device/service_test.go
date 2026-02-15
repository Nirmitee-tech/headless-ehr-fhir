package device

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// -- Mock Repository --

type mockDeviceRepo struct {
	store map[uuid.UUID]*Device
}

func newMockDeviceRepo() *mockDeviceRepo {
	return &mockDeviceRepo{store: make(map[uuid.UUID]*Device)}
}

func (m *mockDeviceRepo) Create(_ context.Context, d *Device) error {
	d.ID = uuid.New()
	if d.FHIRID == "" {
		d.FHIRID = d.ID.String()
	}
	m.store[d.ID] = d
	return nil
}

func (m *mockDeviceRepo) GetByID(_ context.Context, id uuid.UUID) (*Device, error) {
	d, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return d, nil
}

func (m *mockDeviceRepo) GetByFHIRID(_ context.Context, fhirID string) (*Device, error) {
	for _, d := range m.store {
		if d.FHIRID == fhirID {
			return d, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockDeviceRepo) Update(_ context.Context, d *Device) error {
	if _, ok := m.store[d.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[d.ID] = d
	return nil
}

func (m *mockDeviceRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockDeviceRepo) List(_ context.Context, limit, offset int) ([]*Device, int, error) {
	var r []*Device
	for _, d := range m.store {
		r = append(r, d)
	}
	return r, len(r), nil
}

func (m *mockDeviceRepo) ListByPatient(_ context.Context, pid uuid.UUID, limit, offset int) ([]*Device, int, error) {
	var r []*Device
	for _, d := range m.store {
		if d.PatientID != nil && *d.PatientID == pid {
			r = append(r, d)
		}
	}
	return r, len(r), nil
}

func (m *mockDeviceRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Device, int, error) {
	var r []*Device
	for _, d := range m.store {
		r = append(r, d)
	}
	return r, len(r), nil
}

func newTestService() *Service {
	return NewService(newMockDeviceRepo())
}

// -- Service Tests --

func TestCreateDevice_Success(t *testing.T) {
	svc := newTestService()
	d := &Device{DeviceName: "Pulse Oximeter", Status: "active"}
	if err := svc.CreateDevice(context.Background(), d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if d.FHIRID == "" {
		t.Error("expected FHIRID to be set")
	}
	if d.Status != "active" {
		t.Errorf("expected status 'active', got %q", d.Status)
	}
}

func TestCreateDevice_DefaultStatus(t *testing.T) {
	svc := newTestService()
	d := &Device{DeviceName: "Thermometer"}
	if err := svc.CreateDevice(context.Background(), d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != "active" {
		t.Errorf("expected default status 'active', got %q", d.Status)
	}
}

func TestCreateDevice_MissingStatus(t *testing.T) {
	// When status is empty, it should default to "active" (not error)
	svc := newTestService()
	d := &Device{DeviceName: "Thermometer"}
	if err := svc.CreateDevice(context.Background(), d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != "active" {
		t.Errorf("expected default status 'active', got %q", d.Status)
	}
}

func TestCreateDevice_InvalidStatus(t *testing.T) {
	svc := newTestService()
	d := &Device{DeviceName: "Thermometer", Status: "bogus"}
	if err := svc.CreateDevice(context.Background(), d); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestCreateDevice_ValidStatuses(t *testing.T) {
	for _, s := range []string{"active", "inactive", "entered-in-error", "unknown"} {
		svc := newTestService()
		d := &Device{DeviceName: "Test Device", Status: s}
		if err := svc.CreateDevice(context.Background(), d); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
}

func TestCreateDevice_MissingDeviceName(t *testing.T) {
	svc := newTestService()
	d := &Device{Status: "active"}
	if err := svc.CreateDevice(context.Background(), d); err == nil {
		t.Fatal("expected error for missing device_name")
	}
}

func TestGetDevice_Success(t *testing.T) {
	svc := newTestService()
	d := &Device{DeviceName: "Pulse Oximeter", Status: "active"}
	svc.CreateDevice(context.Background(), d)
	got, err := svc.GetDevice(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != d.ID {
		t.Error("ID mismatch")
	}
	if got.DeviceName != "Pulse Oximeter" {
		t.Errorf("DeviceName = %v, want Pulse Oximeter", got.DeviceName)
	}
}

func TestGetDevice_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.GetDevice(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetDeviceByFHIRID(t *testing.T) {
	svc := newTestService()
	d := &Device{DeviceName: "Pulse Oximeter", Status: "active"}
	svc.CreateDevice(context.Background(), d)
	got, err := svc.GetDeviceByFHIRID(context.Background(), d.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.FHIRID != d.FHIRID {
		t.Errorf("FHIRID mismatch: got %v, want %v", got.FHIRID, d.FHIRID)
	}
}

func TestGetDeviceByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.GetDeviceByFHIRID(context.Background(), "nonexistent"); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateDevice_Success(t *testing.T) {
	svc := newTestService()
	d := &Device{DeviceName: "Pulse Oximeter", Status: "active"}
	svc.CreateDevice(context.Background(), d)
	d.Status = "inactive"
	d.DeviceName = "Updated Pulse Oximeter"
	if err := svc.UpdateDevice(context.Background(), d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetDevice(context.Background(), d.ID)
	if got.Status != "inactive" {
		t.Errorf("status = %v, want inactive", got.Status)
	}
	if got.DeviceName != "Updated Pulse Oximeter" {
		t.Errorf("DeviceName = %v, want Updated Pulse Oximeter", got.DeviceName)
	}
}

func TestUpdateDevice_InvalidStatus(t *testing.T) {
	svc := newTestService()
	d := &Device{DeviceName: "Pulse Oximeter", Status: "active"}
	svc.CreateDevice(context.Background(), d)
	d.Status = "invalid"
	if err := svc.UpdateDevice(context.Background(), d); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestDeleteDevice_Success(t *testing.T) {
	svc := newTestService()
	d := &Device{DeviceName: "Pulse Oximeter", Status: "active"}
	svc.CreateDevice(context.Background(), d)
	if err := svc.DeleteDevice(context.Background(), d.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.GetDevice(context.Background(), d.ID); err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestListDevices(t *testing.T) {
	svc := newTestService()
	svc.CreateDevice(context.Background(), &Device{DeviceName: "Device A", Status: "active"})
	svc.CreateDevice(context.Background(), &Device{DeviceName: "Device B", Status: "active"})
	svc.CreateDevice(context.Background(), &Device{DeviceName: "Device C", Status: "inactive"})
	items, total, err := svc.SearchDevices(context.Background(), nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 3 || len(items) != 3 {
		t.Errorf("expected 3 devices, got total=%d len=%d", total, len(items))
	}
}

func TestListDevicesByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateDevice(context.Background(), &Device{DeviceName: "Device A", Status: "active", PatientID: ptrUUID(pid)})
	svc.CreateDevice(context.Background(), &Device{DeviceName: "Device B", Status: "active", PatientID: ptrUUID(pid)})
	svc.CreateDevice(context.Background(), &Device{DeviceName: "Device C", Status: "active", PatientID: ptrUUID(uuid.New())})
	items, total, err := svc.ListDevicesByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 devices, got total=%d len=%d", total, len(items))
	}
}

func TestSearchDevices(t *testing.T) {
	svc := newTestService()
	svc.CreateDevice(context.Background(), &Device{DeviceName: "Device A", Status: "active"})
	svc.CreateDevice(context.Background(), &Device{DeviceName: "Device B", Status: "inactive"})
	params := map[string]string{"status": "active"}
	items, total, err := svc.SearchDevices(context.Background(), params, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Mock returns all items regardless of params, just verify it runs
	if total < 1 || len(items) < 1 {
		t.Errorf("expected at least 1 device, got total=%d len=%d", total, len(items))
	}
}
