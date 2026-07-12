package offerselection

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

// WorkflowResumer is a small, consumer-owned interface: this service only
// ever resumes a paused execution (SendTaskSuccess), it never starts or
// otherwise controls one — that stays with application-management-service
// and the state machine itself.
type WorkflowResumer interface {
	Resume(ctx context.Context, taskToken string, consentGiven bool) error
}

type sfnWorkflowResumer struct {
	client *sfn.Client
}

func NewSFNWorkflowResumer(client *sfn.Client) *sfnWorkflowResumer {
	return &sfnWorkflowResumer{client: client}
}

func (r *sfnWorkflowResumer) Resume(ctx context.Context, taskToken string, consentGiven bool) error {
	output, err := json.Marshal(map[string]bool{"consentGiven": consentGiven})
	if err != nil {
		return fmt.Errorf("encoding resume output: %w", err)
	}

	_, err = r.client.SendTaskSuccess(ctx, &sfn.SendTaskSuccessInput{
		TaskToken: &taskToken,
		Output:    stringPtr(string(output)),
	})
	if err != nil {
		return fmt.Errorf("sending task success: %w", err)
	}
	return nil
}

func stringPtr(s string) *string { return &s }
