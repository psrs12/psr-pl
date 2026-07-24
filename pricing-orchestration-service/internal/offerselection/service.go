package offerselection

import (
	"context"
	"errors"
	"fmt"

	"pricing-orchestration-service/internal/pricing"
)

var ErrAlreadyConfirmed = errors.New("offer already confirmed")
var ErrOfferNotFound = errors.New("selected offer id not found among presented offers")

type Service struct {
	repo     Repository
	workflow WorkflowResumer
}

func NewService(repo Repository, workflow WorkflowResumer) *Service {
	return &Service{repo: repo, workflow: workflow}
}

// Present is called by present-offer-lambda when the workflow pauses at
// AwaitOfferSelection — it persists the offer and the task token needed
// to resume the execution later. This is the only writer of TaskToken.
func (s *Service) Present(ctx context.Context, offer SelectedOffer) error {
	offer.Status = StatusPendingSelection
	if err := s.repo.Create(ctx, &offer); err != nil {
		return fmt.Errorf("presenting offer: %w", err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, applicationID string) (*SelectedOffer, error) {
	offer, err := s.repo.Get(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("getting selected offer %s: %w", applicationID, err)
	}
	return offer, nil
}

// Confirm records the applicant's offer selection and hard-pull consent,
// then resumes the paused Step Functions execution. Consent must be
// explicit and true — this is the FCRA hard-pull consent gate the
// workflow was missing before this capability existed. selectedOfferID
// must match one of the offers presented in Present; the same field is
// required even when declining, since it identifies which offer the
// decline applies to.
func (s *Service) Confirm(ctx context.Context, applicationID, selectedOfferID string, consentGiven bool) (*SelectedOffer, error) {
	offer, err := s.repo.Get(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("getting selected offer %s: %w", applicationID, err)
	}
	if offer.Status == StatusConfirmed {
		return nil, ErrAlreadyConfirmed
	}

	selected, err := findOffer(offer.Offers, selectedOfferID)
	if err != nil {
		return nil, err
	}

	// Resume the workflow before persisting CONFIRMED: if the resume call
	// fails, the record must stay PENDING_SELECTION so the applicant can
	// retry — marking it confirmed first would strand it, unconfirmable
	// and unretryable, on a resume failure.
	if err := s.workflow.Resume(ctx, offer.TaskToken, consentGiven, selected); err != nil {
		return nil, fmt.Errorf("resuming workflow for %s: %w", applicationID, err)
	}

	offer.Status = StatusConfirmed
	offer.ConsentGiven = consentGiven
	offer.SelectedOfferID = selectedOfferID
	if err := s.repo.Update(ctx, offer); err != nil {
		return nil, fmt.Errorf("updating selected offer %s: %w", applicationID, err)
	}
	return offer, nil
}

func findOffer(offers []pricing.OfferOption, id string) (pricing.OfferOption, error) {
	for _, o := range offers {
		if o.ID == id {
			return o, nil
		}
	}
	return pricing.OfferOption{}, ErrOfferNotFound
}
