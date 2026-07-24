package pricing

// minTermMonths/maxTermMonths bound the term variants PriceOptions
// generates around the requested term -- placeholders, not underwriting
// policy.
const (
	minTermMonths  = 24
	maxTermMonths  = 84
	termStepMonths = 12
)

// PriceOptions is a simplified underwriting/pricing stand-in: it caps
// each variant's amount by a debt-to-income check at that variant's term,
// and tiers APR off the credit score. Not a real pricing engine —
// structurally correct (reads credit + financials, returns priced terms)
// but the actual formula is a placeholder for later replacement.
//
// It returns multiple offers, not one: the requested term, plus a
// longer/lower-payment variant and a shorter/faster-payoff variant where
// those stay within [minTermMonths, maxTermMonths]. A real pricing engine
// would likely vary APR by term/risk too; this keeps APR constant across
// variants (by score only) as a deliberate simplification, same caveat as
// the rest of this placeholder formula.
func PriceOptions(input WorkflowInput, credit CreditSummary) []OfferOption {
	apr := aprForScore(credit.Score)

	options := []OfferOption{
		buildOption("standard", "Requested Terms", input, input.RequestedTermMonths, apr),
	}
	if longerTerm := input.RequestedTermMonths + termStepMonths; longerTerm <= maxTermMonths {
		options = append(options, buildOption("lower-payment", "Lower Monthly Payment", input, longerTerm, apr))
	}
	if shorterTerm := input.RequestedTermMonths - termStepMonths; shorterTerm >= minTermMonths {
		options = append(options, buildOption("faster-payoff", "Faster Payoff", input, shorterTerm, apr))
	}
	return options
}

func buildOption(id, label string, input WorkflowInput, termMonths int, apr float64) OfferOption {
	amount := input.RequestedAmountCents
	if maxAffordable := maxAffordableAmount(input, termMonths); amount > maxAffordable {
		amount = maxAffordable
	}
	return OfferOption{
		ID:    id,
		Label: label,
		PricedOffer: PricedOffer{
			AmountCents:   amount,
			TermMonths:    termMonths,
			APRPercentage: apr,
		},
	}
}

func maxAffordableAmount(input WorkflowInput, termMonths int) int64 {
	monthlyIncomeCents := input.AnnualIncomeCents / 12
	availableCents := monthlyIncomeCents - input.MonthlyObligationsCents
	if availableCents <= 0 {
		return 0
	}
	return availableCents * int64(termMonths)
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
