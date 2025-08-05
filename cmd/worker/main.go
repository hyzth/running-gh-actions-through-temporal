package main

import (
	"fmt"
	"os"

	"github.com/temporalio/temporal_gh_actions/github"
	"github.com/temporalio/temporal_gh_actions/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// GitHub App configuration
	app := github.GitHubApp{
		Org:        "my-org",
		ID:         1234567,
		PrivateKey: "my-private-key",
	}

	// Create a Temporal client
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		fmt.Println("Failed to create Temporal client:", err)
		os.Exit(1)
	}

	// Create a Temporal worker
	taskQueue := "my-task-queue"
	temporalWorker := worker.New(temporalClient, taskQueue, worker.Options{})

	// Register workflows
	temporalWorker.RegisterWorkflow(workflows.RunGitHubAction)

	// Register activities
	activities := workflows.Activities{App: app}
	temporalWorker.RegisterActivity(&activities)

	// Run the worker
	if err = temporalWorker.Run(worker.InterruptCh()); err != nil {
		fmt.Println("Worker failure:", err)
		os.Exit(1)
	}
}
