package document

// StandardRequiredDocuments is the fixed set of documents every applicant
// must submit. Placeholder content — a real implementation would source
// this from a configurable, product/jurisdiction-specific policy, not a
// Go literal.
var StandardRequiredDocuments = []RequiredDocument{
	{
		ID:          "proof-of-identity",
		Title:       "Proof of Identity",
		Description: "A government-issued photo ID (driver's license, passport, or state ID).",
		Required:    true,
	},
	{
		ID:          "proof-of-income",
		Title:       "Proof of Income",
		Description: "Your most recent pay stub, or two years of tax returns if self-employed.",
		Required:    true,
	},
	{
		ID:          "proof-of-address",
		Title:       "Proof of Address",
		Description: "A recent utility bill or bank statement showing your current address.",
		Required:    true,
	},
}

func RequiredDocumentIDs() []string {
	var ids []string
	for _, d := range StandardRequiredDocuments {
		if d.Required {
			ids = append(ids, d.ID)
		}
	}
	return ids
}
