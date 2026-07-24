package offerselection

import (
	"context"
	"errors"
	"testing"

	"pricing-orchestration-service/internal/pricing"
)

var testOffers = []pricing.OfferOption{
	{ID: "standard", Label: "Requested Terms", PricedOffer: pricing.PricedOffer{AmountCents: 1000000, TermMonths: 60, APRPercentage: 11.9}},
}

type fakeRepository struct {
	offers map[string]*SelectedOffer
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{offers: map[string]*SelectedOffer{}}
}

func (f *fakeRepository) Create(ctx context.Context, offer *SelectedOffer) error {
	f.offers[offer.ApplicationID] = offer
	return nil
}

func (f *fakeRepository) Get(ctx context.Context, applicationID string) (*SelectedOffer, error) {
	offer, ok := f.offers[applicationID]
	if !ok {
		return nil, ErrNotFound
	}
	return offer, nil
}

func (f *fakeRepository) Update(ctx context.Context, offer *SelectedOffer) error {
	f.offers[offer.ApplicationID] = offer
	return nil
}

type fakeWorkflowResumer struct {
	resumed      bool
	consentGiven bool
	taskToken    string
	failWith     error
}

func (f *fakeWorkflowResumer) Resume(ctx context.Context, taskToken string, consentGiven bool, selected pricing.OfferOption) error {
	if f.failWith != nil {
		return f.failWith
	}
	f.resumed = true
	f.taskToken = taskToken
	f.consentGiven = consentGiven
	return nil
}

func TestConfirmResumesWorkflowWithConsent(t *testing.T) {
	repo := newFakeRepository()
	resumer := &fakeWorkflowResumer{}
	svc := NewService(repo, resumer)

	_ = svc.Present(context.Background(), SelectedOffer{
		ApplicationID: "app-1", Offers: testOffers, TaskToken: "token-abc",
	})

	offer, err := svc.Confirm(context.Background(), "app-1", "standard", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if offer.Status != StatusConfirmed {
		t.Errorf("Status = %s, want %s", offer.Status, StatusConfirmed)
	}
	if !resumer.resumed || !resumer.consentGiven || resumer.taskToken != "token-abc" {
		t.Errorf("workflow not resumed correctly: resumed=%v consent=%v token=%s", resumer.resumed, resumer.consentGiven, resumer.taskToken)
	}
}

func TestConfirmPropagatesDeclinedConsent(t *testing.T) {
	repo := newFakeRepository()
	resumer := &fakeWorkflowResumer{}
	svc := NewService(repo, resumer)

	_ = svc.Present(context.Background(), SelectedOffer{ApplicationID: "app-2", Offers: testOffers, TaskToken: "token-xyz"})

	_, err := svc.Confirm(context.Background(), "app-2", "standard", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resumer.consentGiven {
		t.Error("expected consentGiven=false to be passed through to the workflow resume, not overridden")
	}
}

func TestConfirmRejectsAlreadyConfirmed(t *testing.T) {
	repo := newFakeRepository()
	resumer := &fakeWorkflowResumer{}
	svc := NewService(repo, resumer)

	_ = svc.Present(context.Background(), SelectedOffer{ApplicationID: "app-3", Offers: testOffers, TaskToken: "token-1"})
	_, _ = svc.Confirm(context.Background(), "app-3", "standard", true)

	_, err := svc.Confirm(context.Background(), "app-3", "standard", true)
	if !errors.Is(err, ErrAlreadyConfirmed) {
		t.Errorf("expected ErrAlreadyConfirmed, got %v", err)
	}
}

func TestConfirmLeavesRecordRetryableWhenResumeFails(t *testing.T) {
	repo := newFakeRepository()
	resumer := &fakeWorkflowResumer{failWith: errors.New("step functions unavailable")}
	svc := NewService(repo, resumer)

	_ = svc.Present(context.Background(), SelectedOffer{ApplicationID: "app-4", Offers: testOffers, TaskToken: "token-4"})

	_, err := svc.Confirm(context.Background(), "app-4", "standard", true)
	if err == nil {
		t.Fatal("expected an error when the workflow resume fails")
	}

	offer, getErr := repo.Get(context.Background(), "app-4")
	if getErr != nil {
		t.Fatalf("unexpected error re-reading offer: %v", getErr)
	}
	if offer.Status != StatusPendingSelection {
		t.Errorf("Status = %s, want %s (must stay retryable, not be marked confirmed, when the resume failed)", offer.Status, StatusPendingSelection)
	}

	// A retry with a working resumer should now succeed.
	repo2 := repo
	svc2 := NewService(repo2, &fakeWorkflowResumer{})
	confirmed, err := svc2.Confirm(context.Background(), "app-4", "standard", true)
	if err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
	if confirmed.Status != StatusConfirmed {
		t.Errorf("Status = %s, want %s after successful retry", confirmed.Status, StatusConfirmed)
	}
}

func TestConfirmUnknownApplicationReturnsNotFound(t *testing.T) {
	svc := NewService(newFakeRepository(), &fakeWorkflowResumer{})

	_, err := svc.Confirm(context.Background(), "does-not-exist", "standard", true)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestConfirmRejectsUnknownOfferID(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo, &fakeWorkflowResumer{})

	_ = svc.Present(context.Background(), SelectedOffer{ApplicationID: "app-5", Offers: testOffers, TaskToken: "token-5"})

	_, err := svc.Confirm(context.Background(), "app-5", "not-a-real-offer-id", true)
	if !errors.Is(err, ErrOfferNotFound) {
		t.Errorf("expected ErrOfferNotFound, got %v", err)
	}
}
