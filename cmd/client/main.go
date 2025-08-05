package main

import (
	"context"
	"fmt"
	"os"

	"github.com/temporalio/temporal_gh_actions/workflows"
	"go.temporal.io/sdk/client"
)

func main() {
	// Temporal client
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		fmt.Println("Failed to create Temporal client:", err)
		os.Exit(1)
	}

	// Start a workflow
	if err := runWaitAndEcho(context.Background(), temporalClient); err != nil {
		fmt.Println("Failed to run workflow:", err)
		os.Exit(1)
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
		fmt.Println("Failed to execute workflow:", err)
		os.Exit(1)
	}

	var result workflows.GitHubActionResponse

	if err = future.Get(ctx, &result); err != nil {
		fmt.Println("Failed to get workflow result:", err)
		os.Exit(1)
	}
	return nil
}
