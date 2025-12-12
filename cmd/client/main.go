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
	if err := run(context.Background(), temporalClient); err != nil {
		log.Fatalf("failed to run workflow: %v", err)
	}
}

func run(ctx context.Context, temporalClient client.Client) error {
	taskQueue := "my-task-queue"
	opts := client.StartWorkflowOptions{TaskQueue: taskQueue}

	wf := workflows.ExecuteConcurrentGitHubActions
	args := []workflows.GitHubActionRequest{
		{
			Org:          os.Getenv("GITHUB_ORG"),
			Repo:         "running-gh-actions-through-temporal",
			Ref:          "main",
			WorkflowFile: "wait-and-echo.yaml",
			Inputs: []struct {
				Key   string
				Value string
			}{
				{Key: "wait_time", Value: "20"},
				{Key: "message", Value: "my custom message"},
			},
		},
		{
			Org:          os.Getenv("GITHUB_ORG"),
			Repo:         "test-gha-1",
			Ref:          "main",
			WorkflowFile: "test-1.yaml",
			Inputs: []struct {
				Key   string
				Value string
			}{
				{Key: "wait_time", Value: "30"},
				{Key: "message", Value: "my custom message"},
			},
		},
	}

	future, err := temporalClient.ExecuteWorkflow(ctx, opts, wf, args)
	if err != nil {
		log.Fatalf("failed to execute workflow: %v", err)
	}

	var results []workflows.GitHubActionResponse

	if err = future.Get(ctx, &results); err != nil {
		log.Fatalf("failed to get workflow result: %v", err)
	}
	return nil
}
