// offerlog-mock is a local stand-in for the external OfferLog system, for
// exercising application-management-service's invitation path in dev
// without a real OfferLog to point at. It is not part of the production
// service and is never deployed.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
)

type offer struct {
	IntakeID        string  `json:"intakeId"`
	OfferID         string  `json:"offerId"`
	CampaignOfferID string  `json:"campaignOfferId"`
	TermMonths      int     `json:"requestedTermMonths"`
	AmountCents     int64   `json:"requestedAmountCents"`
	APRPercentage   float64 `json:"aprPercentage"`
	FirstName       string  `json:"firstName"`
	LastName        string  `json:"lastName"`
	Address         address `json:"address"`
	Status          string  `json:"status"`
}

type address struct {
	Line1    string `json:"line1"`
	City     string `json:"city"`
	State    string `json:"state"`
	Postcode string `json:"postcode"`
}

// store keys every offer by both its invitation token and its intakeId, so
// ValidateOffer can be called with either (application-management-service
// calls it once with the raw token, and again with the intakeId at
// submission time to re-validate live).
type store struct {
	mu    sync.Mutex
	byKey map[string]*offer
}

func newStore() *store {
	s := &store{byKey: map[string]*offer{}}

	seed := func(token string, o *offer) {
		s.byKey[token] = o
		s.byKey[o.IntakeID] = o
	}

	seed("VALID-INVITE-123", &offer{
		IntakeID: "intake-fixture-active", OfferID: "offer-1", CampaignOfferID: "campaign-1",
		TermMonths: 60, AmountCents: 1500000, APRPercentage: 11.9,
		FirstName: "Jane", LastName: "Smith",
		Address: address{Line1: "47 W 5th St", City: "Austin", State: "TX", Postcode: "78701"},
		Status:  "ACTIVE",
	})
	seed("USED-INVITE-456", &offer{
		IntakeID: "intake-fixture-used", OfferID: "offer-2", CampaignOfferID: "campaign-1",
		TermMonths: 48, AmountCents: 1000000, APRPercentage: 13.5,
		FirstName: "Sam", LastName: "Lee",
		Address: address{Line1: "10 Congress Ave", City: "Austin", State: "TX", Postcode: "78701"},
		Status:  "USED",
	})
	return s
}

// any token not already seeded is treated as unknown → 404, except tokens
// starting with "VALID-" which mint a fresh active offer on first use, so
// ad-hoc tokens work without pre-registering them.
func (s *store) resolve(token string) (*offer, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if o, ok := s.byKey[token]; ok {
		return o, true
	}
	if len(token) > len("VALID-") && token[:len("VALID-")] == "VALID-" {
		o := &offer{
			IntakeID: "intake-" + randomHex(), OfferID: "offer-" + randomHex(), CampaignOfferID: "campaign-adhoc",
			TermMonths: 60, AmountCents: 1500000, APRPercentage: 11.9,
			FirstName: "Jane", LastName: "Smith",
			Address: address{Line1: "47 W 5th St", City: "Austin", State: "TX", Postcode: "78701"},
			Status:  "ACTIVE",
		}
		s.byKey[token] = o
		s.byKey[o.IntakeID] = o
		return o, true
	}
	return nil, false
}

func (s *store) setStatus(intakeID, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	o, ok := s.byKey[intakeID]
	if !ok {
		return false
	}
	o.Status = status
	return true
}

func randomHex() string {
	buf := make([]byte, 4)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func main() {
	port := os.Getenv("OFFERLOG_MOCK_PORT")
	if port == "" {
		port = "9090"
	}

	s := newStore()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /invitations/validate", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		o, ok := s.resolve(body.Token)
		if !ok {
			http.Error(w, "no such invitation", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, o)
	})

	mux.HandleFunc("PUT /invitations/{intakeId}/status", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !s.setStatus(r.PathValue("intakeId"), body.Status) {
			http.Error(w, "no such invitation", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": body.Status})
	})

	log.Printf("offerlog-mock listening on :%s (fixtures: VALID-INVITE-123 active, USED-INVITE-456 used, any VALID-* token mints a fresh active offer)", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
