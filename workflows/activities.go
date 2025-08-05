package workflows

import (
	"context"
	"time"

	"github.com/temporalio/temporal_gh_actions/github"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

type Activities struct {
	App github.GitHubApp
}

func (a *Activities) TriggerGitHubActionActivity(
	ctx context.Context,
	request GitHubActionRequest,
	dispatchID string,
) error {
	// Convert our input slice to a map
	inputs := mapFromSlice(request.Inputs)

	// Check if the user is using the dispatch_id input key
	if _, ok := inputs["dispatch_id"]; ok {
		err := ReservedInputKeyError{}
		return temporal.NewNonRetryableApplicationError(err.Error(), err.Name(), nil)
	}

	// Add the dispatch ID to the inputs
	inputs["dispatch_id"] = dispatchID

	client, err := github.NewClient(ctx, a.App, request.Repo)
	if err != nil {
		return err
	}

	return client.TriggerAction(
		ctx,
		request.Org,
		request.Repo,
		request.WorkflowFile,
		request.Ref,
		inputs,
	)
}

func (a *Activities) GetActionIDActivity(
	ctx context.Context,
	request GitHubActionRequest,
	dispatchID string,
) (int64, error) {
	client, err := github.NewClient(ctx, a.App, request.Repo)
	if err != nil {
		return -1, err
	}

	return client.GetActionID(
		ctx,
		request.Org,
		request.Repo,
		request.WorkflowFile,
		request.Ref,
		dispatchID,
	)
}

func (a *Activities) AwaitActionCompletionActivity(
	ctx context.Context,
	request GitHubActionRequest,
	actionID int64,
	pollRate time.Duration,
) (*GitHubActionResponse, error) {
	client, err := github.NewClient(ctx, a.App, request.Repo)
	if err != nil {
		return nil, err
	}

	for {
		status, err := client.GetActionStatus(ctx, request.Org, request.Repo, actionID)
		if err != nil {
			return nil, err
		}

		if status.IsRunning() {
			select {
			// The ctx timed out or was cancelled before the action completed
			case <-ctx.Done():
				return nil, ctx.Err()
			// Wait for the specified poll rate before checking again
			case <-time.After(pollRate):
				activity.RecordHeartbeat(ctx, status)
				continue
			}
		}

		resp := &GitHubActionResponse{
			Status:     status.Status,
			Conclusion: status.Conclusion,
			URL:        status.URL,
		}

		// Return a non-retryable error if the action failed since this activity cannot retry the action
		if status.IsFailure() {
			err := GithubActionConclusionError{}
			return resp, temporal.NewNonRetryableApplicationError(
				err.Error(),
				err.Name(),
				nil,
				status.Conclusion,
			)
		}

		// The action completed successfully
		return resp, nil
	}
}

func mapFromSlice(slice []struct {
	Key   string
	Value string
},
) map[string]any {
	result := map[string]any{}
	for _, item := range slice {
		result[item.Key] = item.Value
	}
	return result
}

type ReservedInputKeyError struct{}

func (e ReservedInputKeyError) Error() string {
	return "dispatch_id input is a reserved input key"
}

func (e ReservedInputKeyError) Name() string {
	return "ReservedInputKeyError"
}

type GithubActionConclusionError struct{}

func (e GithubActionConclusionError) Error() string {
	return "GitHub action concluded with a failure"
}

func (e GithubActionConclusionError) Name() string {
	return "GitHubActionConclusionError"
}
