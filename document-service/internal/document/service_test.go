package document

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	records map[string]*Record
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{records: map[string]*Record{}}
}

func (f *fakeRepository) Get(ctx context.Context, applicationID string) (*Record, error) {
	record, ok := f.records[applicationID]
	if !ok {
		return nil, ErrNotFound
	}
	return record, nil
}

func (f *fakeRepository) Save(ctx context.Context, record *Record) error {
	f.records[record.ApplicationID] = record
	return nil
}

type fakeApplicationClient struct {
	updated bool
	status  string
}

func (f *fakeApplicationClient) UpdateStatus(ctx context.Context, applicationID, status string) error {
	f.updated = true
	f.status = status
	return nil
}

func TestSubmitRejectsUnknownDocument(t *testing.T) {
	svc := NewService(newFakeRepository(), &fakeApplicationClient{})

	_, err := svc.Submit(context.Background(), "app-1", "not-a-real-document")
	if !errors.Is(err, ErrUnknownDocument) {
		t.Errorf("expected ErrUnknownDocument, got %v", err)
	}
}

func TestSubmitStaysPendingUntilAllRequiredDocumentsIn(t *testing.T) {
	appClient := &fakeApplicationClient{}
	svc := NewService(newFakeRepository(), appClient)

	record, err := svc.Submit(context.Background(), "app-2", "proof-of-identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.Status != StatusPending {
		t.Errorf("Status = %s, want %s after only one of three required documents", record.Status, StatusPending)
	}
	if appClient.updated {
		t.Error("application status should not be updated before all required documents are in")
	}
}

func TestSubmitCompletesAndAdvancesStatusOnceAllRequiredDocumentsIn(t *testing.T) {
	appClient := &fakeApplicationClient{}
	svc := NewService(newFakeRepository(), appClient)

	var record *Record
	var err error
	for _, id := range RequiredDocumentIDs() {
		record, err = svc.Submit(context.Background(), "app-3", id)
		if err != nil {
			t.Fatalf("unexpected error submitting %s: %v", id, err)
		}
	}

	if record.Status != StatusComplete {
		t.Errorf("Status = %s, want %s once every required document is in", record.Status, StatusComplete)
	}
	if record.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
	if !appClient.updated || appClient.status != "OFFER_ACCEPTED" {
		t.Errorf("expected application status advanced to OFFER_ACCEPTED, got updated=%v status=%s", appClient.updated, appClient.status)
	}
}

func TestSubmitIsIdempotentForRepeatedDocument(t *testing.T) {
	svc := NewService(newFakeRepository(), &fakeApplicationClient{})

	_, _ = svc.Submit(context.Background(), "app-4", "proof-of-identity")
	record, err := svc.Submit(context.Background(), "app-4", "proof-of-identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(record.SubmittedDocumentIDs) != 1 {
		t.Errorf("SubmittedDocumentIDs = %v, want exactly one entry for a repeated submission", record.SubmittedDocumentIDs)
	}
}
