package document

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrUnknownDocument = errors.New("unrecognized document id")

type Service struct {
	repo        Repository
	application ApplicationStatusUpdater
	now         func() time.Time
}

func NewService(repo Repository, application ApplicationStatusUpdater) *Service {
	return &Service{repo: repo, application: application, now: time.Now}
}

func (s *Service) RequiredDocuments(ctx context.Context) []RequiredDocument {
	return StandardRequiredDocuments
}

func (s *Service) Get(ctx context.Context, applicationID string) (*Record, error) {
	record, err := s.repo.Get(ctx, applicationID)
	if errors.Is(err, ErrNotFound) {
		return &Record{ApplicationID: applicationID, Status: StatusPending}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting document record %s: %w", applicationID, err)
	}
	return record, nil
}

// Submit records one document as submitted for the application. Once
// every required document id has been submitted, the record moves to
// COMPLETE and the application is advanced to OFFER_ACCEPTED — the next
// stage's applicant-facing message ("completing final verification and
// collecting your disbursement details before your loan can be funded")
// describes exactly this point: documents in, awaiting funding.
func (s *Service) Submit(ctx context.Context, applicationID, documentID string) (*Record, error) {
	if !isKnownDocument(documentID) {
		return nil, ErrUnknownDocument
	}

	record, err := s.repo.Get(ctx, applicationID)
	if errors.Is(err, ErrNotFound) {
		record = &Record{ApplicationID: applicationID, Status: StatusPending}
	} else if err != nil {
		return nil, fmt.Errorf("getting document record %s: %w", applicationID, err)
	}

	if !contains(record.SubmittedDocumentIDs, documentID) {
		record.SubmittedDocumentIDs = append(record.SubmittedDocumentIDs, documentID)
	}

	wasComplete := record.Status == StatusComplete
	if !wasComplete && containsAll(record.SubmittedDocumentIDs, RequiredDocumentIDs()) {
		record.Status = StatusComplete
		now := s.now()
		record.CompletedAt = &now
	}

	if err := s.repo.Save(ctx, record); err != nil {
		return nil, fmt.Errorf("saving document record %s: %w", applicationID, err)
	}

	if !wasComplete && record.Status == StatusComplete {
		if err := s.application.UpdateStatus(ctx, applicationID, "OFFER_ACCEPTED"); err != nil {
			return nil, fmt.Errorf("advancing application status: %w", err)
		}
	}

	return record, nil
}

func isKnownDocument(id string) bool {
	for _, d := range StandardRequiredDocuments {
		if d.ID == id {
			return true
		}
	}
	return false
}

func contains(have []string, id string) bool {
	for _, v := range have {
		if v == id {
			return true
		}
	}
	return false
}

func containsAll(have, want []string) bool {
	set := make(map[string]bool, len(have))
	for _, id := range have {
		set[id] = true
	}
	for _, id := range want {
		if !set[id] {
			return false
		}
	}
	return true
}
