package main

import (
	"context"
	"log"
	"os"

	"github.com/temporalio/temporal_gh_actions/workflows"
	"go.temporal.io/sdk/client"
)

func main() {
	// Temporal client
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalf("failed to create Temporal client: %v", err)
	}

	// Start a workflow
	if err := runWaitAndEcho(context.Background(), temporalClient); err != nil {
		log.Fatalf("failed to run workflow: %v", err)
	}
}

func runWaitAndEcho(ctx context.Context, temporalClient client.Client) error {
	taskQueue := "my-task-queue"
	opts := client.StartWorkflowOptions{TaskQueue: taskQueue}

	wf := workflows.RunGitHubAction
	args := workflows.GitHubActionRequest{
		Org:          "my-org",
		Repo:         "my-repo",
		Ref:          "main",
		WorkflowFile: "wait-and-echo.yaml",
		Inputs: []struct {
			Key   string
			Value string
		}{
			{Key: "wait_time", Value: "100"},
			{Key: "message", Value: "my custom message"},
		},
	}

	future, err := temporalClient.ExecuteWorkflow(ctx, opts, wf, args)
	if err != nil {
		log.Fatalf("failed to execute workflow: %v", err)
	}

	var result workflows.GitHubActionResponse

	if err = future.Get(ctx, &result); err != nil {
		log.Fatalf("failed to get workflow result: %v", err)
	}
	return nil
}
