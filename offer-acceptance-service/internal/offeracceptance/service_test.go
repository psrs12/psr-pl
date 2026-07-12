package offeracceptance

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	records map[string]*EsignRecord
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{records: map[string]*EsignRecord{}}
}

func (f *fakeRepository) Create(ctx context.Context, record *EsignRecord) error {
	if _, exists := f.records[record.ApplicationID]; exists {
		return ErrAlreadySigned
	}
	f.records[record.ApplicationID] = record
	return nil
}

func (f *fakeRepository) Get(ctx context.Context, applicationID string) (*EsignRecord, error) {
	record, ok := f.records[applicationID]
	if !ok {
		return nil, ErrNotFound
	}
	return record, nil
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

func TestESignRejectsPartialAcceptance(t *testing.T) {
	svc := NewService(newFakeRepository(), &fakeApplicationClient{})

	_, err := svc.ESign(context.Background(), "app-1", []string{"tila-disclosure"})
	if !errors.Is(err, ErrMissingRequiredDeclarations) {
		t.Fatalf("expected ErrMissingRequiredDeclarations, got %v", err)
	}
}

func TestESignAcceptsAllRequired(t *testing.T) {
	appClient := &fakeApplicationClient{}
	svc := NewService(newFakeRepository(), appClient)

	record, err := svc.ESign(context.Background(), "app-2", RequiredDeclarationIDs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.ApplicationID != "app-2" {
		t.Errorf("ApplicationID = %s, want app-2", record.ApplicationID)
	}
	if !appClient.updated || appClient.status != "DOCUMENTS_REQUIRED" {
		t.Errorf("expected application status advanced to DOCUMENTS_REQUIRED, got updated=%v status=%s", appClient.updated, appClient.status)
	}
}

func TestESignRejectsDoubleSigning(t *testing.T) {
	svc := NewService(newFakeRepository(), &fakeApplicationClient{})

	_, err := svc.ESign(context.Background(), "app-3", RequiredDeclarationIDs())
	if err != nil {
		t.Fatalf("unexpected error on first sign: %v", err)
	}

	_, err = svc.ESign(context.Background(), "app-3", RequiredDeclarationIDs())
	if !errors.Is(err, ErrAlreadySigned) {
		t.Fatalf("expected ErrAlreadySigned on second sign, got %v", err)
	}
}
