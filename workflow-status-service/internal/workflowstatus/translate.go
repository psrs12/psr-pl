package workflowstatus

// stateStatus maps a currently-executing Step Functions state name (from
// definition.asl.json in pricing-orchestration-service — the two must be
// kept in sync) to what an applicant should see while that state runs.
var stateStatus = map[string]Status{
	"SoftPullRequest":         StatusSoftPullPending,
	"PricingCalculation":      StatusPricingPending,
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

// forOutcome translates the terminal decision outcome (from the
// execution's final output — see pricing.Decision) into a WorkflowStatus.
func forOutcome(applicationID, outcome string) WorkflowStatus {
	switch outcome {
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
