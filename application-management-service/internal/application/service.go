package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrIncompleteApplication = errors.New("application is missing required fields")
	ErrLoginFailed           = errors.New("application id, last 4 of SSN, and date of birth do not match")
)

const sessionTTL = 30 * time.Minute

type Service struct {
	repo     Repository
	offers   offerLogAdapter
	sessions SessionStore
	workflow WorkflowStarter
	now      func() time.Time
}

type offerLogAdapter interface {
	OfferValidator
	OfferStatusUpdater
}

func NewService(repo Repository, offers offerLogAdapter, sessions SessionStore, workflow WorkflowStarter) *Service {
	return &Service{repo: repo, offers: offers, sessions: sessions, workflow: workflow, now: time.Now}
}

// ResolveInvitation validates an invitation token against OfferLog and
// returns the prefill data for the invitation entry path. It persists
// nothing: the intake flow is stateless until submission. If the offer
// fails validation, the caller degrades to the direct path rather than
// treating this as an error.
func (s *Service) ResolveInvitation(ctx context.Context, token string) (offer *Offer, prefill Applicant, degraded bool, err error) {
	offer, prefill, err = s.offers.ValidateOffer(ctx, token)
	if errors.Is(err, ErrOfferUnavailable) {
		return nil, Applicant{}, true, nil
	}
	if err != nil {
		return nil, Applicant{}, false, fmt.Errorf("resolving invitation: %w", err)
	}
	return offer, prefill, false, nil
}

type SubmitRequest struct {
	EntryPath EntryPath
	IntakeID  string
	PartnerID string
	Applicant Applicant
}

// Submit is the sole persistence trigger for an application: nothing is
// written before this point. For the invitation path, it re-validates the
// offer live via its intakeId (no local cache of "already in use"),
// snapshots the offer economics onto the record, persists the
// application, and only then marks the offer used in OfferLog.
func (s *Service) Submit(ctx context.Context, req SubmitRequest) (*Application, error) {
	if err := Validate(req.Applicant); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrIncompleteApplication, err)
	}

	var snapshot *Offer
	if req.EntryPath == EntryPathInvitation && req.IntakeID != "" {
		offer, prefill, err := s.offers.ValidateOffer(ctx, req.IntakeID)
		switch {
		case errors.Is(err, ErrOfferUnavailable):
			req.EntryPath = EntryPathDirect
		case err != nil:
			return nil, fmt.Errorf("re-validating offer at submission: %w", err)
		default:
			req.Applicant.FirstName = prefill.FirstName
			req.Applicant.LastName = prefill.LastName
			req.Applicant.Address = prefill.Address
			snapshot = offer
		}
	}

	now := s.now()
	app := &Application{
		ID:           uuid.NewString(),
		EntryPath:    req.EntryPath,
		PartnerID:    req.PartnerID,
		Applicant:    req.Applicant,
		Offer:        snapshot,
		Status:       StatusSubmitted,
		SubmittedAt:  now,
		LastActionAt: now,
	}

	if err := s.repo.Create(ctx, app); err != nil {
		return nil, fmt.Errorf("persisting application: %w", err)
	}

	executionArn, err := s.workflow.Start(ctx, WorkflowInput{
		ApplicationID:           app.ID,
		RequestedAmountCents:    app.Applicant.RequestedAmountCents,
		RequestedTermMonths:     app.Applicant.RequestedTermMonths,
		AnnualIncomeCents:       app.Applicant.AnnualIncomeCents,
		MonthlyObligationsCents: app.Applicant.MonthlyObligationsCents,
	})
	if err != nil {
		return nil, fmt.Errorf("starting pricing workflow: %w", err)
	}
	app.ExecutionARN = executionArn
	if err := s.repo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("recording workflow execution: %w", err)
	}

	if snapshot != nil {
		if err := s.offers.MarkUsed(ctx, snapshot.IntakeID); err != nil {
			return nil, fmt.Errorf("marking offer used: %w", err)
		}
	}

	return app, nil
}

// ExecutionARN returns the Step Functions execution linked to an
// application, for workflow-status-service's internal use only — the
// mapping lives here, not duplicated into workflow-status-service's own
// storage.
func (s *Service) ExecutionARN(ctx context.Context, id string) (string, error) {
	app, err := s.repo.Get(ctx, id)
	if err != nil {
		return "", fmt.Errorf("getting application %s: %w", id, err)
	}
	return app.ExecutionARN, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Application, error) {
	app, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting application %s: %w", id, err)
	}
	return app, nil
}

// UpdateStatus is used by the workflow (Step Functions / expiry sweep) to
// advance an application's status after submission. There is no applicant-
// facing edit path once an application is submitted.
func (s *Service) UpdateStatus(ctx context.Context, id string, status Status) (*Application, error) {
	app, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting application %s: %w", id, err)
	}

	app.Status = status
	app.LastActionAt = s.now()

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("updating application %s: %w", id, err)
	}
	return app, nil
}

// Login verifies an applicant's identity against the persisted application
// (application id, last 4 of SSN, date of birth) and issues a short-lived
// session token for the self-service status portal.
func (s *Service) Login(ctx context.Context, applicationID, last4SSN, dateOfBirth string) (token string, err error) {
	app, err := s.repo.Get(ctx, applicationID)
	if errors.Is(err, ErrNotFound) {
		return "", ErrLoginFailed
	}
	if err != nil {
		return "", fmt.Errorf("getting application %s: %w", applicationID, err)
	}

	if !strings.HasSuffix(app.Applicant.SSN, last4SSN) || app.Applicant.DateOfBirth != dateOfBirth {
		return "", ErrLoginFailed
	}

	token, err = s.sessions.Create(ctx, app.ID, sessionTTL)
	if err != nil {
		return "", fmt.Errorf("creating session: %w", err)
	}
	return token, nil
}

// Authenticate validates a session token and confirms it belongs to the
// requested application, for use as handler-level authorization.
func (s *Service) Authenticate(ctx context.Context, token, applicationID string) error {
	ownerID, err := s.sessions.Validate(ctx, token)
	if errors.Is(err, ErrSessionInvalid) {
		return ErrSessionInvalid
	}
	if err != nil {
		return fmt.Errorf("validating session: %w", err)
	}
	if ownerID != applicationID {
		return ErrSessionInvalid
	}
	return nil
}
