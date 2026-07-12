package pricing

// Decide is a simplified decision-routing stand-in, tiering the hard-pull
// score against the priced offer. A real implementation would also weigh
// fraud/compliance signals from earlier gates; those aren't modeled here.
func Decide(offer PricedOffer, hardPull CreditSummary) Decision {
	switch {
	case offer.AmountCents <= 0:
		return Decision{Outcome: "DECLINED", Reason: "insufficient debt-to-income capacity"}
	case hardPull.Score >= 670:
		return Decision{Outcome: "APPROVED", Reason: "meets standard approval threshold"}
	case hardPull.Score >= 600:
		return Decision{Outcome: "REFERRED", Reason: "borderline score requires manual underwriting"}
	default:
		return Decision{Outcome: "DECLINED", Reason: "below minimum credit threshold"}
	}
}
