package github

import "slices"

type Action struct {
	Status     string
	Conclusion string
	URL        string
}

// https://docs.github.com/en/rest/actions/workflow-runs?apiVersion=2022-11-28#list-workflow-runs-for-a-repository

// States that indicate the action has finished
var terminalStates = []string{
	"completed",
	"cancelled",
	"failure",
	"neutral",
	"skipped",
	"success",
	"timed_out",
}

// States that indicate the action has failed
var failureStates = []string{"failure", "timed_out"}

func (action Action) IsRunning() bool {
	return !slices.Contains(terminalStates, action.Status)
}

func (action Action) IsFailure() bool {
	return slices.Contains(failureStates, action.Conclusion)
}
