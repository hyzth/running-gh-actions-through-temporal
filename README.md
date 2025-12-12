# running-gh-actions-through-temporal

Orchestrate GitHub Actions using Temporal.

For more details, see
https://temporal.io/blog/running-github-actions-temporal-guide.

## Usage

### Prerequisites

- Authentication is handled via GitHub Apps. You will need a GitHub App and
  install it in every repo where you want to run GitHub Actions using Temporal.

- Provide your GitHub organization name, GitHub App ID, and GitHub App private
  key file's path via `GITHUB_ORG`, `GITHUB_APP_ID`, and
  `GITHUB_PRIVATE_KEY_PATH` env vars respectively.

- Set up the worker and client if needed via `cmd/worker/main.go` and
  `cmd/client/main.go` respectively.

### Running it

After you have set everything up, run the Temporal server, worker, and client

```bash
# Start a Temporal server in the background
temporal server start-dev &

# Start the worker in the background
go run ./cmd/worker/main.go &

# Run the client to trigger an action
go run ./cmd/client/main.go
```

View the workflows in the Temporal UI at `http://localhost:7233`.

Go to your repository's Actions tab to see the GitHub Action run.
