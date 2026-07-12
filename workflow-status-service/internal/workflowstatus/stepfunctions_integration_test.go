package workflowstatus

import (
	"context"
	"os"
	"testing"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

// TestSFNExecutionReaderAgainstRealAPI exercises the real Step Functions
// SDK calls against a running endpoint (e.g. Step Functions Local) —
// skipped unless STEPFUNCTIONS_ENDPOINT_URL and TEST_EXECUTION_ARN are
// set, since it needs a real execution to inspect. See
// pricing-orchestration-service/Makefile's stepfunctions-up /
// run-happy-path targets to produce one locally.
func TestSFNExecutionReaderAgainstRealAPI(t *testing.T) {
	endpoint := os.Getenv("STEPFUNCTIONS_ENDPOINT_URL")
	executionArn := os.Getenv("TEST_EXECUTION_ARN")
	if endpoint == "" || executionArn == "" {
		t.Skip("set STEPFUNCTIONS_ENDPOINT_URL and TEST_EXECUTION_ARN to run this against a real Step Functions endpoint")
	}

	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("loading AWS config: %v", err)
	}
	client := sfn.NewFromConfig(cfg, func(o *sfn.Options) { o.BaseEndpoint = &endpoint })
	reader := NewSFNExecutionReader(client)

	state, err := reader.CurrentState(ctx, executionArn)
	if err != nil {
		t.Fatalf("CurrentState: %v", err)
	}
	if state == "" {
		t.Error("CurrentState returned empty — expected the most recently entered state name")
	}
	t.Logf("CurrentState() = %q", state)

	status, outcome, err := reader.Outcome(ctx, executionArn)
	if err != nil {
		t.Fatalf("Outcome: %v", err)
	}
	t.Logf("Outcome() = status=%s outcome=%q", status, outcome)
}
