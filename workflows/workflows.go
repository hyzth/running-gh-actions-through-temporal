package workflows

import (
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type GitHubActionRequest struct {
	// Organization and repository details for the action
	Org  string
	Repo string
	Ref  string

	// File where the action is defined
	WorkflowFile string

	// Inputs for the workflow
	Inputs []struct {
		Key   string
		Value string
	}
}

type GitHubActionResponse struct {
	Status     string
	Conclusion string
	URL        string
}

func RunGitHubAction(
	ctx workflow.Context,
	request GitHubActionRequest,
) (*GitHubActionResponse, error) {
	// Create a dispatch ID to track the GH action we are running
	dispatchID, err := uuidSideEffect(ctx)
	if err != nil {
		return nil, err
	}

	// Start the GH action
	if err = triggerGitHubAction(ctx, request, dispatchID); err != nil {
		return nil, err
	}

	// Use the dispatch ID we added to get the action's ID
	actionID, err := getActionID(ctx, request, dispatchID)
	if err != nil {
		return nil, err
	}

	// Wait for the action to finish and return its final status
	status, err := awaitActionCompletion(ctx, request, actionID)
	if err != nil {
		return nil, err
	}

	return status, nil
}

func uuidSideEffect(ctx workflow.Context) (string, error) {
	sideEffect := workflow.SideEffect(
		ctx,
		func(ctx workflow.Context) any { return uuid.New().String() },
	)

	var uuid string

	if err := sideEffect.Get(&uuid); err != nil {
		return "", err
	}

	return uuid, nil
}

func triggerGitHubAction(
	ctx workflow.Context,
	request GitHubActionRequest,
	dispatchID string,
) error {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// Timeout of the entire activity execution, including retries.
		ScheduleToCloseTimeout: 5 * time.Minute,
		// Timeout of a single activity attempt.
		StartToCloseTimeout: 5 * time.Second,
		// Define the retry behavior of the activity.
		RetryPolicy: &temporal.RetryPolicy{
			// Maximum time between retries.
			MaximumInterval: 5 * time.Second,
			// Errors that will prevent the activity from being retried.
			NonRetryableErrorTypes: []string{
				ReservedInputKeyError{}.Name(),
			},
		},
	})

	activity := (*Activities).TriggerGitHubActionActivity
	future := workflow.ExecuteActivity(ctx, activity, request, dispatchID)

	if err := future.Get(ctx, nil); err != nil {
		return err
	}

	return nil
}

func getActionID(
	ctx workflow.Context,
	request GitHubActionRequest,
	dispatchID string,
) (int64, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// Timeout of the entire activity execution, including retries.
		ScheduleToCloseTimeout: 10 * time.Minute,
		// Timeout of a single activity attempt.
		StartToCloseTimeout: 5 * time.Second,
		// Define the retry behavior of the activity.
		RetryPolicy: &temporal.RetryPolicy{
			// Maximum time between retries.
			MaximumInterval: 30 * time.Second,
		},
	})

	activity := (*Activities).GetActionIDActivity
	future := workflow.ExecuteActivity(ctx, activity, request, dispatchID)

	var actionID int64

	if err := future.Get(ctx, &actionID); err != nil {
		return -1, err
	}

	return actionID, nil
}

func awaitActionCompletion(
	ctx workflow.Context,
	request GitHubActionRequest,
	actionID int64,
) (*GitHubActionResponse, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		// Timeout of the entire activity execution, including retries.
		ScheduleToCloseTimeout: 1 * time.Hour,
		// Maximum time between activity heartbeats.
		HeartbeatTimeout: 1 * time.Minute,
		// Define the retry behavior of the activity.
		RetryPolicy: &temporal.RetryPolicy{
			// Errors that will prevent the activity from being retried.
			NonRetryableErrorTypes: []string{
				GithubActionConclusionError{}.Name(),
			},
		},
	})

	// Time between polling the action's status
	pollRate := 10 * time.Second

	activity := (*Activities).AwaitActionCompletionActivity
	future := workflow.ExecuteActivity(ctx, activity, request, actionID, pollRate)

	var resp GitHubActionResponse

	if err := future.Get(ctx, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
