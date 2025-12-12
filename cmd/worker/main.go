package main

import (
	"log"
	"os"
	"strconv"

	"github.com/temporalio/temporal_gh_actions/github"
	"github.com/temporalio/temporal_gh_actions/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	id, err := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)
	if err != nil {
		log.Fatalf("unable to parse GITHUB_APP_ID: %v", err)
	}
	// GitHub App configuration
	app := github.GitHubApp{
		Org:            os.Getenv("GITHUB_ORG"),
		ID:             id,
		PrivateKeyFile: os.Getenv("GITHUB_PRIVATE_KEY_FILE"),
	}

	// Create a Temporal client
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalf("failed to create Temporal client: %v", err)
	}

	// Create a Temporal worker
	taskQueue := "my-task-queue"
	temporalWorker := worker.New(temporalClient, taskQueue, worker.Options{})

	// Register workflows
	temporalWorker.RegisterWorkflow(workflows.ExecuteSequentialGitHubActions)
	temporalWorker.RegisterWorkflow(workflows.ExecuteConcurrentGitHubActions)

	// Register activities
	activities := workflows.Activities{App: app}
	temporalWorker.RegisterActivity(&activities)

	// Run the worker
	if err = temporalWorker.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("worker failure: %v", err)
	}
}
