package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/inbox"
	"github.com/google/uuid"
)

func TestMessagePoolCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("msgpool")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *inbox.MessagePool
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewMessagePoolRepoPG(globalDB.Pool)
			p := &inbox.MessagePool{
				PoolName:    "Cardiology Pool",
				PoolType:    "department",
				Description: ptrStr("Messages for cardiology department"),
				IsActive:    true,
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			created = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create message pool: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		pool := createTestMessagePool(t, ctx, tenantID)

		var fetched *inbox.MessagePool
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewMessagePoolRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pool.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.PoolName != "Test Pool" {
			t.Errorf("expected pool_name=Test Pool, got %s", fetched.PoolName)
		}
		if !fetched.IsActive {
			t.Error("expected pool to be active")
		}
	})

	t.Run("Update", func(t *testing.T) {
		pool := createTestMessagePool(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewMessagePoolRepoPG(globalDB.Pool)
			pool.PoolName = "Updated Pool"
			pool.Description = ptrStr("Updated description")
			pool.IsActive = false
			return repo.Update(ctx, pool)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *inbox.MessagePool
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewMessagePoolRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pool.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.PoolName != "Updated Pool" {
			t.Errorf("expected pool_name=Updated Pool, got %s", fetched.PoolName)
		}
		if fetched.IsActive {
			t.Error("expected pool to be inactive")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		pool := createTestMessagePool(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewMessagePoolRepoPG(globalDB.Pool)
			return repo.Delete(ctx, pool.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewMessagePoolRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, pool.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted message pool")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*inbox.MessagePool
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewMessagePoolRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 message pool")
		}
		_ = results
	})
}

func TestInboxMessageCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("inboxmsg")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "InboxPatient", "Test", "MRN-INBOX-001")
	senderID := createTestSystemUser(t, ctx, globalDB.Pool, tenantID, "sender_doc")
	recipientID := createTestSystemUser(t, ctx, globalDB.Pool, tenantID, "recip_doc")
	pool := createTestMessagePool(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *inbox.InboxMessage
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			m := &inbox.InboxMessage{
				MessageType: "result",
				Priority:    "normal",
				Subject:     "Lab results available",
				Body:        ptrStr("CBC results are now available for review"),
				PatientID:   &patient.ID,
				SenderID:    &senderID,
				RecipientID: &recipientID,
				Status:      "unread",
				IsUrgent:    false,
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			created = m
			return nil
		})
		if err != nil {
			t.Fatalf("Create inbox message: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.ThreadID == nil {
			t.Fatal("expected non-nil ThreadID (auto-set to ID)")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		msg := createTestInboxMessage(t, ctx, tenantID, senderID, recipientID, &patient.ID)

		var fetched *inbox.InboxMessage
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, msg.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Subject != "Test Message" {
			t.Errorf("expected subject=Test Message, got %s", fetched.Subject)
		}
		if fetched.Status != "unread" {
			t.Errorf("expected status=unread, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		msg := createTestInboxMessage(t, ctx, tenantID, senderID, recipientID, &patient.ID)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			msg.Status = "read"
			msg.ReadAt = &now
			msg.IsUrgent = true
			msg.Priority = "high"
			return repo.Update(ctx, msg)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *inbox.InboxMessage
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, msg.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "read" {
			t.Errorf("expected status=read, got %s", fetched.Status)
		}
		if fetched.ReadAt == nil {
			t.Error("expected non-nil ReadAt")
		}
		if !fetched.IsUrgent {
			t.Error("expected is_urgent=true")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		msg := createTestInboxMessage(t, ctx, tenantID, senderID, recipientID, &patient.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			return repo.Delete(ctx, msg.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, msg.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted inbox message")
		}
	})

	t.Run("ListByRecipient", func(t *testing.T) {
		createTestInboxMessage(t, ctx, tenantID, senderID, recipientID, &patient.ID)

		var results []*inbox.InboxMessage
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByRecipient(ctx, recipientID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByRecipient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 message for recipient")
		}
		for _, r := range results {
			if r.RecipientID == nil || *r.RecipientID != recipientID {
				t.Errorf("expected recipient_id=%s, got %v", recipientID, r.RecipientID)
			}
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*inbox.InboxMessage
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 message for patient")
		}
		for _, r := range results {
			if r.PatientID == nil || *r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %v", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*inbox.InboxMessage
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status":       "unread",
				"recipient_id": recipientID.String(),
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		for _, r := range results {
			if r.Status != "unread" {
				t.Errorf("expected status=unread, got %s", r.Status)
			}
		}
		_ = total
	})

	t.Run("AddPoolMember_GetPoolMembers_RemovePoolMember", func(t *testing.T) {
		var memberID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			member := &inbox.MessagePoolMember{
				PoolID:   pool.ID,
				UserID:   senderID,
				Role:     ptrStr("member"),
				IsActive: true,
			}
			if err := repo.AddPoolMember(ctx, member); err != nil {
				return err
			}
			memberID = member.ID
			return nil
		})
		if err != nil {
			t.Fatalf("AddPoolMember: %v", err)
		}

		var members []*inbox.MessagePoolMember
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			var err error
			members, err = repo.GetPoolMembers(ctx, pool.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetPoolMembers: %v", err)
		}
		if len(members) != 1 {
			t.Fatalf("expected 1 pool member, got %d", len(members))
		}
		if members[0].UserID != senderID {
			t.Errorf("expected user_id=%s, got %s", senderID, members[0].UserID)
		}

		// Remove
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			return repo.RemovePoolMember(ctx, memberID)
		})
		if err != nil {
			t.Fatalf("RemovePoolMember: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
			var err error
			members, err = repo.GetPoolMembers(ctx, pool.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetPoolMembers after remove: %v", err)
		}
		if len(members) != 0 {
			t.Errorf("expected 0 pool members after remove, got %d", len(members))
		}
	})
}

func TestCosignRequestCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("cosign")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	requesterID := createTestSystemUser(t, ctx, globalDB.Pool, tenantID, "requester_user")
	cosignerID := createTestSystemUser(t, ctx, globalDB.Pool, tenantID, "cosigner_user")

	t.Run("Create", func(t *testing.T) {
		var created *inbox.CosignRequest
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewCosignRequestRepoPG(globalDB.Pool)
			cr := &inbox.CosignRequest{
				DocumentType: "progress-note",
				RequesterID:  requesterID,
				CosignerID:   cosignerID,
				Status:       "pending",
				Note:         ptrStr("Please review and cosign"),
				RequestedAt:  time.Now(),
			}
			if err := repo.Create(ctx, cr); err != nil {
				return err
			}
			created = cr
			return nil
		})
		if err != nil {
			t.Fatalf("Create cosign request: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		cr := createTestCosignRequest(t, ctx, tenantID, requesterID, cosignerID)

		var fetched *inbox.CosignRequest
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewCosignRequestRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, cr.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.RequesterID != requesterID {
			t.Errorf("expected requester_id=%s, got %s", requesterID, fetched.RequesterID)
		}
		if fetched.Status != "pending" {
			t.Errorf("expected status=pending, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		cr := createTestCosignRequest(t, ctx, tenantID, requesterID, cosignerID)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewCosignRequestRepoPG(globalDB.Pool)
			cr.Status = "signed"
			cr.Note = ptrStr("Reviewed and approved")
			cr.RespondedAt = &now
			return repo.Update(ctx, cr)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *inbox.CosignRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewCosignRequestRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, cr.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "signed" {
			t.Errorf("expected status=signed, got %s", fetched.Status)
		}
		if fetched.RespondedAt == nil {
			t.Error("expected non-nil RespondedAt")
		}
	})

	t.Run("ListByCosigner", func(t *testing.T) {
		createTestCosignRequest(t, ctx, tenantID, requesterID, cosignerID)

		var results []*inbox.CosignRequest
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewCosignRequestRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByCosigner(ctx, cosignerID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByCosigner: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 cosign request for cosigner")
		}
		for _, r := range results {
			if r.CosignerID != cosignerID {
				t.Errorf("expected cosigner_id=%s, got %s", cosignerID, r.CosignerID)
			}
		}
	})

	t.Run("ListByRequester", func(t *testing.T) {
		var results []*inbox.CosignRequest
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewCosignRequestRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByRequester(ctx, requesterID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByRequester: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 cosign request from requester")
		}
		for _, r := range results {
			if r.RequesterID != requesterID {
				t.Errorf("expected requester_id=%s, got %s", requesterID, r.RequesterID)
			}
		}
	})
}

func TestPatientListCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("patlist")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	ownerID := createTestSystemUser(t, ctx, globalDB.Pool, tenantID, "list_owner")
	patient1 := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ListPatient1", "Test", "MRN-LIST-001")
	patient2 := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ListPatient2", "Test", "MRN-LIST-002")

	t.Run("Create", func(t *testing.T) {
		var created *inbox.PatientList
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			l := &inbox.PatientList{
				ListName:    "ICU Patients",
				ListType:    "manual",
				OwnerID:     ownerID,
				Description: ptrStr("Current ICU patients"),
				IsActive:    true,
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			created = l
			return nil
		})
		if err != nil {
			t.Fatalf("Create patient list: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		list := createTestPatientList(t, ctx, tenantID, ownerID)

		var fetched *inbox.PatientList
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, list.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.ListName != "My Patients" {
			t.Errorf("expected list_name=My Patients, got %s", fetched.ListName)
		}
		if fetched.OwnerID != ownerID {
			t.Errorf("expected owner_id=%s, got %s", ownerID, fetched.OwnerID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		list := createTestPatientList(t, ctx, tenantID, ownerID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			list.ListName = "Updated Patient List"
			list.Description = ptrStr("Updated description")
			list.IsActive = false
			return repo.Update(ctx, list)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *inbox.PatientList
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, list.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.ListName != "Updated Patient List" {
			t.Errorf("expected list_name=Updated Patient List, got %s", fetched.ListName)
		}
		if fetched.IsActive {
			t.Error("expected list to be inactive")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		list := createTestPatientList(t, ctx, tenantID, ownerID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			return repo.Delete(ctx, list.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, list.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted patient list")
		}
	})

	t.Run("ListByOwner", func(t *testing.T) {
		createTestPatientList(t, ctx, tenantID, ownerID)

		var results []*inbox.PatientList
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByOwner(ctx, ownerID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByOwner: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 list for owner")
		}
		for _, r := range results {
			if r.OwnerID != ownerID {
				t.Errorf("expected owner_id=%s, got %s", ownerID, r.OwnerID)
			}
		}
	})

	t.Run("AddMember_GetMembers_UpdateMember_RemoveMember", func(t *testing.T) {
		list := createTestPatientList(t, ctx, tenantID, ownerID)

		// Add member 1
		var member1ID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			m := &inbox.PatientListMember{
				ListID:    list.ID,
				PatientID: patient1.ID,
				Priority:  1,
				Flags:     ptrStr("critical"),
				OneLiner:  ptrStr("CHF exacerbation, awaiting echo"),
				AddedBy:   &ownerID,
			}
			if err := repo.AddMember(ctx, m); err != nil {
				return err
			}
			member1ID = m.ID
			return nil
		})
		if err != nil {
			t.Fatalf("AddMember: %v", err)
		}

		// Add member 2
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			m := &inbox.PatientListMember{
				ListID:    list.ID,
				PatientID: patient2.ID,
				Priority:  2,
				OneLiner:  ptrStr("Pneumonia, improving"),
				AddedBy:   &ownerID,
			}
			return repo.AddMember(ctx, m)
		})
		if err != nil {
			t.Fatalf("AddMember (second): %v", err)
		}

		// Get members
		var members []*inbox.PatientListMember
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			var err error
			members, total, err = repo.GetMembers(ctx, list.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("GetMembers: %v", err)
		}
		if total != 2 {
			t.Fatalf("expected 2 members, got %d", total)
		}

		// Update member
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			m := &inbox.PatientListMember{
				ID:       member1ID,
				Priority: 3,
				Flags:    ptrStr("stable"),
				OneLiner: ptrStr("CHF improving, discharge planning"),
			}
			return repo.UpdateMember(ctx, m)
		})
		if err != nil {
			t.Fatalf("UpdateMember: %v", err)
		}

		// Verify update
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			var err error
			members, total, err = repo.GetMembers(ctx, list.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("GetMembers after update: %v", err)
		}
		found := false
		for _, m := range members {
			if m.ID == member1ID {
				found = true
				if m.Priority != 3 {
					t.Errorf("expected priority=3, got %d", m.Priority)
				}
				if m.Flags == nil || *m.Flags != "stable" {
					t.Errorf("expected flags=stable, got %v", m.Flags)
				}
			}
		}
		if !found {
			t.Error("member1 not found in list")
		}

		// Remove member
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			return repo.RemoveMember(ctx, member1ID)
		})
		if err != nil {
			t.Fatalf("RemoveMember: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewPatientListRepoPG(globalDB.Pool)
			var err error
			members, total, err = repo.GetMembers(ctx, list.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("GetMembers after remove: %v", err)
		}
		if total != 1 {
			t.Errorf("expected 1 member after remove, got %d", total)
		}
	})
}

func TestHandoffCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("handoff")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "HandoffPatient", "Test", "MRN-HANDOFF-001")
	fromProviderID := createTestSystemUser(t, ctx, globalDB.Pool, tenantID, "from_provider")
	toProviderID := createTestSystemUser(t, ctx, globalDB.Pool, tenantID, "to_provider")

	t.Run("Create", func(t *testing.T) {
		var created *inbox.HandoffRecord
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewHandoffRepoPG(globalDB.Pool)
			h := &inbox.HandoffRecord{
				PatientID:          patient.ID,
				FromProviderID:     fromProviderID,
				ToProviderID:       toProviderID,
				HandoffType:        "shift-change",
				IllnessSeverity:    ptrStr("moderate"),
				PatientSummary:     ptrStr("65yo male with pneumonia, improving on day 3 of antibiotics"),
				ActionList:         ptrStr("1. Continue IV abx\n2. Repeat CXR in AM\n3. If afebrile 24hrs, switch to PO"),
				SituationAwareness: ptrStr("Watch for respiratory decompensation"),
				Synthesis:          ptrStr("Likely ready for step-down if continues to improve"),
				ContingencyPlan:    ptrStr("If O2 req increases, obtain ABG and consider ICU transfer"),
				Status:             "pending",
			}
			if err := repo.Create(ctx, h); err != nil {
				return err
			}
			created = h
			return nil
		})
		if err != nil {
			t.Fatalf("Create handoff: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		handoff := createTestHandoff(t, ctx, tenantID, patient.ID, fromProviderID, toProviderID)

		var fetched *inbox.HandoffRecord
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewHandoffRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, handoff.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
		if fetched.FromProviderID != fromProviderID {
			t.Errorf("expected from_provider_id=%s, got %s", fromProviderID, fetched.FromProviderID)
		}
		if fetched.Status != "pending" {
			t.Errorf("expected status=pending, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		handoff := createTestHandoff(t, ctx, tenantID, patient.ID, fromProviderID, toProviderID)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewHandoffRepoPG(globalDB.Pool)
			handoff.Status = "acknowledged"
			handoff.AcknowledgedAt = &now
			handoff.PatientSummary = ptrStr("Updated summary")
			return repo.Update(ctx, handoff)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *inbox.HandoffRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewHandoffRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, handoff.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "acknowledged" {
			t.Errorf("expected status=acknowledged, got %s", fetched.Status)
		}
		if fetched.AcknowledgedAt == nil {
			t.Error("expected non-nil AcknowledgedAt")
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		createTestHandoff(t, ctx, tenantID, patient.ID, fromProviderID, toProviderID)

		var results []*inbox.HandoffRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewHandoffRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 handoff for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("ListByProvider", func(t *testing.T) {
		var results []*inbox.HandoffRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := inbox.NewHandoffRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByProvider(ctx, fromProviderID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByProvider: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 handoff for provider")
		}
		for _, r := range results {
			if r.FromProviderID != fromProviderID && r.ToProviderID != fromProviderID {
				t.Errorf("expected provider_id=%s in from or to, got from=%s to=%s", fromProviderID, r.FromProviderID, r.ToProviderID)
			}
		}
	})
}

// =========== Test Helpers ===========

func createTestMessagePool(t *testing.T, ctx context.Context, tenantID string) *inbox.MessagePool {
	t.Helper()
	var result *inbox.MessagePool
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := inbox.NewMessagePoolRepoPG(globalDB.Pool)
		p := &inbox.MessagePool{
			PoolName:    "Test Pool",
			PoolType:    "general",
			Description: ptrStr("Test pool for integration tests"),
			IsActive:    true,
		}
		if err := repo.Create(ctx, p); err != nil {
			return err
		}
		result = p
		return nil
	})
	if err != nil {
		t.Fatalf("create test message pool: %v", err)
	}
	return result
}

func createTestInboxMessage(t *testing.T, ctx context.Context, tenantID string, senderID, recipientID uuid.UUID, patientID *uuid.UUID) *inbox.InboxMessage {
	t.Helper()
	var result *inbox.InboxMessage
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := inbox.NewInboxMessageRepoPG(globalDB.Pool)
		m := &inbox.InboxMessage{
			MessageType: "general",
			Priority:    "normal",
			Subject:     "Test Message",
			Body:        ptrStr("This is a test message body"),
			PatientID:   patientID,
			SenderID:    &senderID,
			RecipientID: &recipientID,
			Status:      "unread",
			IsUrgent:    false,
		}
		if err := repo.Create(ctx, m); err != nil {
			return err
		}
		result = m
		return nil
	})
	if err != nil {
		t.Fatalf("create test inbox message: %v", err)
	}
	return result
}

func createTestCosignRequest(t *testing.T, ctx context.Context, tenantID string, requesterID, cosignerID uuid.UUID) *inbox.CosignRequest {
	t.Helper()
	var result *inbox.CosignRequest
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := inbox.NewCosignRequestRepoPG(globalDB.Pool)
		cr := &inbox.CosignRequest{
			DocumentType: "note",
			RequesterID:  requesterID,
			CosignerID:   cosignerID,
			Status:       "pending",
			Note:         ptrStr("Please cosign"),
			RequestedAt:  time.Now(),
		}
		if err := repo.Create(ctx, cr); err != nil {
			return err
		}
		result = cr
		return nil
	})
	if err != nil {
		t.Fatalf("create test cosign request: %v", err)
	}
	return result
}

func createTestPatientList(t *testing.T, ctx context.Context, tenantID string, ownerID uuid.UUID) *inbox.PatientList {
	t.Helper()
	var result *inbox.PatientList
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := inbox.NewPatientListRepoPG(globalDB.Pool)
		l := &inbox.PatientList{
			ListName:    "My Patients",
			ListType:    "manual",
			OwnerID:     ownerID,
			Description: ptrStr("My patient worklist"),
			IsActive:    true,
		}
		if err := repo.Create(ctx, l); err != nil {
			return err
		}
		result = l
		return nil
	})
	if err != nil {
		t.Fatalf("create test patient list: %v", err)
	}
	return result
}

func createTestHandoff(t *testing.T, ctx context.Context, tenantID string, patientID, fromID, toID uuid.UUID) *inbox.HandoffRecord {
	t.Helper()
	var result *inbox.HandoffRecord
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := inbox.NewHandoffRepoPG(globalDB.Pool)
		h := &inbox.HandoffRecord{
			PatientID:       patientID,
			FromProviderID:  fromID,
			ToProviderID:    toID,
			HandoffType:     "shift-change",
			IllnessSeverity: ptrStr("stable"),
			PatientSummary:  ptrStr("Test patient summary"),
			ActionList:      ptrStr("Continue current plan"),
			Status:          "pending",
		}
		if err := repo.Create(ctx, h); err != nil {
			return err
		}
		result = h
		return nil
	})
	if err != nil {
		t.Fatalf("create test handoff: %v", err)
	}
	return result
}
