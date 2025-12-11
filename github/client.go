package github

import (
	"context"
	"errors"
	"iter"
	"net/http"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v80/github"
)

type GitHubApp struct {
	Org            string
	ID             int64
	PrivateKeyFile string
}

type Client struct {
	*github.Client
}

func NewClient(ctx context.Context, app GitHubApp, repo string) (*Client, error) {
	appTransport, err := ghinstallation.NewAppsTransportKeyFromFile(
		http.DefaultTransport,
		app.ID,
		app.PrivateKeyFile,
	)
	if err != nil {
		return nil, err
	}

	// Create a GitHub client for the app
	appClient := github.NewClient(&http.Client{Transport: appTransport})

	// Find the installation for the app in the repository
	installation, _, err := appClient.Apps.FindRepositoryInstallation(ctx, app.Org, repo)
	if err != nil {
		return nil, err
	}

	installationTransport, err := ghinstallation.NewKeyFromFile(
		http.DefaultTransport,
		app.ID,
		installation.GetID(),
		app.PrivateKeyFile,
	)
	if err != nil {
		return nil, err
	}

	// Create a GitHub client for the installation
	installationClient := github.NewClient(&http.Client{Transport: installationTransport})

	return &Client{Client: installationClient}, nil
}

func (client *Client) TriggerAction(
	ctx context.Context,
	org, repo, workflowFile, ref string,
	inputs map[string]any,
) error {
	// Dispatch the GitHub action
	_, err := client.Actions.CreateWorkflowDispatchEventByFileName(
		ctx,
		org,
		repo,
		workflowFile,
		github.CreateWorkflowDispatchEventRequest{
			Ref:    ref,
			Inputs: inputs,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) GetActionID(
	ctx context.Context,
	org, repo, workflowFile, ref, dispatchID string,
) (int64, error) {
	// Iterate all runs of the action
	for runItem := range client.iterateActionRuns(ctx, org, repo, workflowFile) {
		if runItem.err != nil {
			return -1, runItem.err
		}

		run := runItem.value

		// Iterate all jobs for the run
		for jobItem := range client.iterateRunJobs(ctx, org, repo, run.GetID()) {
			if jobItem.err != nil {
				return -1, jobItem.err
			}

			job := jobItem.value

			// Iterate all steps for the job
			for _, step := range job.Steps {
				// If the step's name contains our dispatch ID,
				// we have found the action we are looking for
				if strings.Contains(step.GetName(), dispatchID) {
					return run.GetID(), nil
				}
			}
		}
	}

	return -1, errors.New("Action ID not found")
}

func (client *Client) GetActionStatus(
	ctx context.Context,
	org, repo string,
	actionID int64,
) (*Action, error) {
	run, _, err := client.Actions.GetWorkflowRunByID(ctx, org, repo, actionID)
	if err != nil {
		return nil, err
	}

	return &Action{
		Status:     run.GetStatus(),
		Conclusion: run.GetConclusion(),
		URL:        run.GetURL(),
	}, nil
}

func (client *Client) iterateActionRuns(
	ctx context.Context,
	org, repo, workflowFile string,
) iter.Seq[iterItem[*github.WorkflowRun]] {
	// Define a func to fetch runs for an action
	fetch := func(page int) ([]*github.WorkflowRun, *github.Response, error) {
		actions, resp, err := client.Actions.ListWorkflowRunsByFileName(
			ctx,
			org,
			repo,
			workflowFile,
			&github.ListWorkflowRunsOptions{
				ListOptions: github.ListOptions{Page: page},
				Event:       "workflow_dispatch",
			},
		)
		if err != nil {
			return nil, resp, err
		}
		return actions.WorkflowRuns, resp, nil
	}

	return newPaginatedIter(newPageFetcher(fetch))
}

func (client *Client) iterateRunJobs(
	ctx context.Context,
	org, repo string,
	runID int64,
) iter.Seq[iterItem[*github.WorkflowJob]] {
	// Define a func to fetch jobs for a run
	fetch := func(page int) ([]*github.WorkflowJob, *github.Response, error) {
		jobs, resp, err := client.Actions.ListWorkflowJobs(
			ctx,
			org,
			repo,
			runID,
			&github.ListWorkflowJobsOptions{
				ListOptions: github.ListOptions{Page: page},
				Filter:      "latest",
			},
		)
		if err != nil {
			return nil, resp, err
		}
		return jobs.Jobs, resp, nil
	}

	return newPaginatedIter(newPageFetcher(fetch))
}
