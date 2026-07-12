package offeracceptance

// StandardDeclarations is the fixed set of declarations every applicant
// must acknowledge. Placeholder content — a real implementation would
// source this from a legal/compliance-managed template, not a Go literal.
var StandardDeclarations = []Declaration{
	{
		ID:       "tila-disclosure",
		Title:    "Truth in Lending Act Disclosure",
		Text:     "I acknowledge I have reviewed the annual percentage rate, finance charge, amount financed, and total of payments for this loan.",
		Required: true,
	},
	{
		ID:       "esign-consent",
		Title:    "Electronic Signature Consent",
		Text:     "I consent to sign this agreement electronically and understand it has the same legal effect as a handwritten signature.",
		Required: true,
	},
	{
		ID:       "arbitration-agreement",
		Title:    "Arbitration Agreement",
		Text:     "I agree that disputes arising from this agreement will be resolved through binding arbitration rather than in court.",
		Required: true,
	},
}

func RequiredDeclarationIDs() []string {
	var ids []string
	for _, d := range StandardDeclarations {
		if d.Required {
			ids = append(ids, d.ID)
		}
	}
	return ids
}
