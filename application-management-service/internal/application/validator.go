package application

import (
	"fmt"
	"strings"
)

// Validate enforces the required-field set from the application-management
// capability spec: identical requirements on every entry path, since only
// prefill/editability differs by path, not which fields are needed.
func Validate(a Applicant) error {
	var missing []string

	require := func(ok bool, field string) {
		if !ok {
			missing = append(missing, field)
		}
	}

	require(a.FirstName != "", "firstName")
	require(a.LastName != "", "lastName")
	require(a.Address.Line1 != "", "address.line1")
	require(a.Address.City != "", "address.city")
	require(a.Address.State != "", "address.state")
	require(a.Address.PostalCode != "", "address.postalCode")
	require(a.Phone != "", "phone")
	require(a.Email != "", "email")
	require(a.Employment.EmploymentStatus != "", "employment.employmentStatus")
	require(a.AnnualIncomeCents > 0, "annualIncomeCents")
	require(a.SSN != "", "ssn")
	require(a.DateOfBirth != "", "dateOfBirth")
	require(a.CitizenshipStatus != "", "citizenshipStatus")
	require(a.RequestedAmountCents > 0, "requestedAmountCents")
	require(a.RequestedTermMonths > 0, "requestedTermMonths")
	require(a.LoanPurpose != "", "loanPurpose")

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}
