package pricing

import "testing"

func TestPriceCapsToAffordability(t *testing.T) {
	input := WorkflowInput{
		ApplicationID:           "app-1",
		RequestedAmountCents:    7_000_000, // $70,000 — exceeds the $60,000 affordability ceiling below
		RequestedTermMonths:     60,
		AnnualIncomeCents:       3_600_000, // $36,000/yr -> $3,000/mo
		MonthlyObligationsCents: 200_000,   // $2,000/mo
	}
	credit := CreditSummary{Score: 700}

	offer := Price(input, credit)

	wantMax := int64(100_000 * 60) // $1,000/mo available * 60 months
	if offer.AmountCents != wantMax {
		t.Errorf("AmountCents = %d, want %d (capped by affordability)", offer.AmountCents, wantMax)
	}
	if offer.APRPercentage != 11.9 {
		t.Errorf("APRPercentage = %v, want 11.9 for a 700 score", offer.APRPercentage)
	}
}

func TestPriceWithinRequestedAmount(t *testing.T) {
	input := WorkflowInput{
		ApplicationID:           "app-2",
		RequestedAmountCents:    1_000_000,
		RequestedTermMonths:     36,
		AnnualIncomeCents:       12_000_000,
		MonthlyObligationsCents: 0,
	}
	credit := CreditSummary{Score: 760}

	offer := Price(input, credit)

	if offer.AmountCents != input.RequestedAmountCents {
		t.Errorf("AmountCents = %d, want the full requested %d (well within affordability)", offer.AmountCents, input.RequestedAmountCents)
	}
	if offer.APRPercentage != 8.9 {
		t.Errorf("APRPercentage = %v, want 8.9 for a 760 score", offer.APRPercentage)
	}
}

func TestDecideDeclinesZeroAffordability(t *testing.T) {
	d := Decide(PricedOffer{AmountCents: 0}, CreditSummary{Score: 800})
	if d.Outcome != "DECLINED" {
		t.Errorf("Outcome = %s, want DECLINED when priced amount is zero", d.Outcome)
	}
}

func TestDecideApprovesGoodScore(t *testing.T) {
	d := Decide(PricedOffer{AmountCents: 1000}, CreditSummary{Score: 700})
	if d.Outcome != "APPROVED" {
		t.Errorf("Outcome = %s, want APPROVED for score 700", d.Outcome)
	}
}

func TestSimulatedCreditPullDeterministic(t *testing.T) {
	a := SimulatedCreditPull("app-3", "SOFT_PULL")
	b := SimulatedCreditPull("app-3", "SOFT_PULL")
	if a.Score != b.Score {
		t.Errorf("expected deterministic score for the same applicationId, got %d and %d", a.Score, b.Score)
	}
	if a.Score < 580 || a.Score > 819 {
		t.Errorf("score %d out of expected 580-819 range", a.Score)
	}
}
