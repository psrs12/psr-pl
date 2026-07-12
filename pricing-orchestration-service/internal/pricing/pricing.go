package pricing

// Price is a simplified underwriting/pricing stand-in: it caps the
// requested amount by a debt-to-income check and tiers APR off the
// credit score. Not a real pricing engine — structurally correct
// (reads credit + financials, returns priced terms) but the actual
// formula is a placeholder for later replacement.
func Price(input WorkflowInput, credit CreditSummary) PricedOffer {
	amount := input.RequestedAmountCents
	if maxAffordable := maxAffordableAmount(input); amount > maxAffordable {
		amount = maxAffordable
	}

	return PricedOffer{
		AmountCents:   amount,
		TermMonths:    input.RequestedTermMonths,
		APRPercentage: aprForScore(credit.Score),
	}
}

func maxAffordableAmount(input WorkflowInput) int64 {
	monthlyIncomeCents := input.AnnualIncomeCents / 12
	availableCents := monthlyIncomeCents - input.MonthlyObligationsCents
	if availableCents <= 0 {
		return 0
	}
	return availableCents * int64(input.RequestedTermMonths)
}

func aprForScore(score int) float64 {
	switch {
	case score >= 740:
		return 8.9
	case score >= 670:
		return 11.9
	case score >= 600:
		return 17.9
	default:
		return 24.9
	}
}
