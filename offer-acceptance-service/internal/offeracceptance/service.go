package offeracceptance

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrMissingRequiredDeclarations = errors.New("not all required declarations were accepted")

type Service struct {
	repo        Repository
	application ApplicationStatusUpdater
	now         func() time.Time
}

func NewService(repo Repository, application ApplicationStatusUpdater) *Service {
	return &Service{repo: repo, application: application, now: time.Now}
}

func (s *Service) Declarations(ctx context.Context) []Declaration {
	return StandardDeclarations
}

// ESign records the applicant's e-signature and advances the application
// to DOCUMENTS_REQUIRED. All required declarations must be present in
// acceptedDeclarationIDs — partial acceptance is not a valid e-signature.
func (s *Service) ESign(ctx context.Context, applicationID string, acceptedDeclarationIDs []string) (*EsignRecord, error) {
	if !containsAll(acceptedDeclarationIDs, RequiredDeclarationIDs()) {
		return nil, ErrMissingRequiredDeclarations
	}

	record := &EsignRecord{
		ApplicationID:          applicationID,
		AcceptedDeclarationIDs: acceptedDeclarationIDs,
		SignedAt:               s.now(),
	}
	if err := s.repo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("recording e-signature: %w", err)
	}

	if err := s.application.UpdateStatus(ctx, applicationID, "DOCUMENTS_REQUIRED"); err != nil {
		return nil, fmt.Errorf("advancing application status: %w", err)
	}

	return record, nil
}

func (s *Service) Get(ctx context.Context, applicationID string) (*EsignRecord, error) {
	record, err := s.repo.Get(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("getting esign record %s: %w", applicationID, err)
	}
	return record, nil
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
