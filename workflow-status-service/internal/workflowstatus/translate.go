package workflowstatus

// stateStatus maps a currently-executing Step Functions state name (from
// definition.asl.json in pricing-orchestration-service — the two must be
// kept in sync) to what an applicant should see while that state runs.
var stateStatus = map[string]Status{
	"SoftPullRequest":         StatusSoftPullPending,
	"PricingCalculation":      StatusPricingPending,
	"AwaitOfferSelection":     StatusOfferPending,
	"ConsentGivenCheck":       StatusOfferPending,
	"ConsentDeclined":         StatusDecisionPending,
	"HardPullRequest":         StatusHardPullPending,
	"DecisionRouting":         StatusDecisionPending,
	"UpdateApplicationStatus": StatusDecisionPending,
}

var waitingNextSteps = []NextStep{
	{Action: "WAIT", Description: "We're processing your application. This page will update automatically."},
}

// forRunningState translates the currently-executing state name into a
// WorkflowStatus. An unrecognized state name still returns something
// reasonable (PROCESSING) rather than failing the request — a new state
// added to the state machine without updating this map shouldn't break
// the status endpoint, just report generically until this map catches up.
func forRunningState(applicationID, currentState string) WorkflowStatus {
	status, ok := stateStatus[currentState]
	if !ok {
		status = StatusProcessing
	}
	return WorkflowStatus{
		ApplicationID: applicationID,
		Status:        status,
		CurrentStep:   currentState,
		NextSteps:     waitingNextSteps,
	}
}

// forApplicationStatus translates application-management-service's live
// status into a WorkflowStatus. Called once the linked Step Functions
// execution has gone terminal: the execution's own decision outcome never
// changes after that point, but the application keeps advancing past it
// (e-sign, documents, funding) in application-management-service's own
// record, which is the single source of truth from here on -- trusting
// only the frozen execution outcome would make this endpoint report
// "APPROVED" forever, even after the applicant e-signs and finishes
// documents. An unrecognized status still returns something reasonable
// (forError) rather than failing the request.
func forApplicationStatus(applicationID, status string) WorkflowStatus {
	switch status {
	case "APPROVED":
		return WorkflowStatus{
			ApplicationID: applicationID,
			Status:        StatusApproved,
			NextSteps: []NextStep{
				{Action: "SELECT_OFFER", Description: "Review and accept your loan offer."},
			},
		}
	case "DECLINED":
		return WorkflowStatus{
			ApplicationID: applicationID,
			Status:        StatusDeclined,
			NextSteps: []NextStep{
				{Action: "NONE", Description: "This application was not approved. See your adverse action notice for details."},
			},
		}
	case "REFERRED":
		return WorkflowStatus{
			ApplicationID: applicationID,
			Status:        StatusReferred,
			NextSteps: []NextStep{
				{Action: "WAIT", Description: "Your application requires manual review by our underwriting team."},
			},
		}
	case "DOCUMENTS_REQUIRED":
		return WorkflowStatus{
			ApplicationID: applicationID,
			Status:        StatusDocumentsRequired,
			NextSteps: []NextStep{
				{Action: "UPLOAD_DOCUMENTS", Description: "Upload the required documents to continue your application."},
			},
		}
	case "OFFER_ACCEPTED":
		return WorkflowStatus{
			ApplicationID: applicationID,
			Status:        StatusOfferAccepted,
			NextSteps: []NextStep{
				{Action: "WAIT", Description: "We are completing final verification and collecting your disbursement details before your loan can be funded."},
			},
		}
	case "FUNDING_PENDING":
		return WorkflowStatus{
			ApplicationID: applicationID,
			Status:        StatusFundingPending,
			NextSteps: []NextStep{
				{Action: "WAIT", Description: "Your loan funds are being prepared and will be disbursed shortly."},
			},
		}
	case "FUNDED":
		return WorkflowStatus{
			ApplicationID: applicationID,
			Status:        StatusFunded,
			NextSteps: []NextStep{
				{Action: "NONE", Description: "Your loan has been funded. The funds have been sent to your nominated account."},
			},
		}
	case "COMPLETED":
		return WorkflowStatus{
			ApplicationID: applicationID,
			Status:        StatusCompleted,
			NextSteps: []NextStep{
				{Action: "NONE", Description: "Your loan application is complete. Thank you for choosing us."},
			},
		}
	default:
		return forError(applicationID)
	}
}

func forError(applicationID string) WorkflowStatus {
	return WorkflowStatus{
		ApplicationID: applicationID,
		Status:        StatusError,
		NextSteps: []NextStep{
			{Action: "CONTACT_SUPPORT", Description: "Something went wrong processing your application. Please contact support."},
		},
	}
}
