# README

This repo is forked from
https://github.com/temporal-community/running-gh-actions-through-temporal.

This code covers using Temporal to orchestrate GitHub Actions.

## Setup

Before you can run GitHub Actions, you need to provide information about your own GitHub app, organization, and repository.

```go
// cmd/worker/main.go
app := github.GitHubApp{
  Org:        "my-org",
  ID:         123456,
  PrivateKey: "my-private-key",
}

// cmd/client/main.go
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
```

After you have set everything up, you can run the Temporal server, the worker, and the client.

```bash
# Start a Temporal server in the background
temporal server start-dev &

# Start the worker in the background
go run ./cmd/worker/main.go &

# Run the client to trigger an action
go run ./cmd/client/main.go
```

Go to `http://localhost:7233` to see the Temporal Web UI and monitor the workflow.

Go to your repository's Actions tab to see the GitHub Action run.
