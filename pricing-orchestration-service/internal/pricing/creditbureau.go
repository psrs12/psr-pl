package pricing

import "hash/fnv"

// SimulatedCreditPull stands in for a real credit bureau integration.
// It derives a deterministic score from the application id (so the same
// application always gets the same simulated score across soft and hard
// pull) rather than a random one, so runs are reproducible for testing.
// source distinguishes a soft pull from a hard pull in the returned
// summary, matching what a real bureau response would carry.
func SimulatedCreditPull(applicationID, source string) CreditSummary {
	h := fnv.New32a()
	_, _ = h.Write([]byte(applicationID))
	// Map the hash into a plausible FICO-like range: 580-820.
	score := 580 + int(h.Sum32()%240)
	return CreditSummary{Score: score, Source: source}
}
