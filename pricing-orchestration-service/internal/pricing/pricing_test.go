package pricing

import "testing"

func TestPriceOptionsCapsToAffordability(t *testing.T) {
	input := WorkflowInput{
		ApplicationID:           "app-1",
		RequestedAmountCents:    7_000_000, // $70,000 — exceeds the $60,000 affordability ceiling below
		RequestedTermMonths:     60,
		AnnualIncomeCents:       3_600_000, // $36,000/yr -> $3,000/mo
		MonthlyObligationsCents: 200_000,   // $2,000/mo
	}
	credit := CreditSummary{Score: 700}

	offers := PriceOptions(input, credit)
	standard := findOption(t, offers, "standard")

	wantMax := int64(100_000 * 60) // $1,000/mo available * 60 months
	if standard.AmountCents != wantMax {
		t.Errorf("AmountCents = %d, want %d (capped by affordability)", standard.AmountCents, wantMax)
	}
	if standard.APRPercentage != 11.9 {
		t.Errorf("APRPercentage = %v, want 11.9 for a 700 score", standard.APRPercentage)
	}
}

func TestPriceOptionsWithinRequestedAmount(t *testing.T) {
	input := WorkflowInput{
		ApplicationID:           "app-2",
		RequestedAmountCents:    1_000_000,
		RequestedTermMonths:     36,
		AnnualIncomeCents:       12_000_000,
		MonthlyObligationsCents: 0,
	}
	credit := CreditSummary{Score: 760}

	offers := PriceOptions(input, credit)
	standard := findOption(t, offers, "standard")

	if standard.AmountCents != input.RequestedAmountCents {
		t.Errorf("AmountCents = %d, want the full requested %d (well within affordability)", standard.AmountCents, input.RequestedAmountCents)
	}
	if standard.APRPercentage != 8.9 {
		t.Errorf("APRPercentage = %v, want 8.9 for a 760 score", standard.APRPercentage)
	}
}

func TestPriceOptionsIncludesLowerAndFasterVariants(t *testing.T) {
	input := WorkflowInput{
		ApplicationID:           "app-3",
		RequestedAmountCents:    1_000_000,
		RequestedTermMonths:     48,
		AnnualIncomeCents:       12_000_000,
		MonthlyObligationsCents: 0,
	}
	offers := PriceOptions(input, CreditSummary{Score: 760})

	if len(offers) != 3 {
		t.Fatalf("len(offers) = %d, want 3 (standard, lower-payment, faster-payoff)", len(offers))
	}
	lower := findOption(t, offers, "lower-payment")
	if lower.TermMonths != 60 {
		t.Errorf("lower-payment TermMonths = %d, want 60", lower.TermMonths)
	}
	faster := findOption(t, offers, "faster-payoff")
	if faster.TermMonths != 36 {
		t.Errorf("faster-payoff TermMonths = %d, want 36", faster.TermMonths)
	}
}

func findOption(t *testing.T, offers []OfferOption, id string) OfferOption {
	t.Helper()
	for _, o := range offers {
		if o.ID == id {
			return o
		}
	}
	t.Fatalf("no offer option with id %q in %+v", id, offers)
	return OfferOption{}
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
