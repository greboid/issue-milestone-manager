package main

import (
	"errors"
	"fmt"
	"github.com/google/go-github/v34/github"
	"issue-manager/rest"
	"os"
	"strings"

	"github.com/blang/semver/v4"
)

func main() {
	client, user, repo, version, err := parseInput(
		os.Getenv("INPUT_TOKEN"),
		os.Getenv("GITHUB_REPOSITORY"),
		os.Getenv("GITHUB_REF"),
		os.Getenv("GITHUB_SHA"),
		)
	if err != nil {
		fmt.Printf("Error with input: %s", err.Error())
	}
	err = tagIssuesWithVersionAsMilestone(client, user, repo, version)
	if err != nil {
		fmt.Printf("Error tagging issues: %s", err.Error())
	} else {
		fmt.Printf("Issues tagged for %s", version)
	}
}

func parseInput(token string, repository string, ref string, sha string) (IssueManager, string, string, string, error) {
	if token == "" {
		return nil, "", "", "", fmt.Errorf("no token specified")
	}
	repo, user := getRepo(repository)
	if repo == "" || user == "" {
		return nil, "", "", "", fmt.Errorf("no user or repo")
	}
	version := refToVersion(ref, sha)
	if version == "unknown" {
		return nil, "", "", "", fmt.Errorf("unable to get version")
	}
	client := rest.NewClient(token)
	return client, user, repo, version, nil
}

func tagIssuesWithVersionAsMilestone(client IssueManager, user string, repo string, version string) error {
	issues, err := client.GetIssues(user, repo)
	if err != nil {
		return fmt.Errorf("unable to get issues: %s", err)
	}
	milestones, err := client.GetMilestones(user, repo)
	if err != nil {
		return fmt.Errorf("unable to get milestones: %s", err.Error())
	}
	currentMilestone := checkMilestoneExists(milestones, &version)
	if currentMilestone == nil {
		currentMilestone, err = client.CreateMilestone(user, repo, version)
		if err != nil {
			return fmt.Errorf("unable to create milestone: %s", err.Error())
		}
	}
	updateErrors := UpdateIssues(client, user, repo, issues, *currentMilestone)
	if updateErrors {
		return errors.New("error updating some issues")
	}
	return nil
}

func getRepo(fullRepo string) (string, string) {
	split := strings.Split(fullRepo, "/")
	if len(split) == 2 {
		return split[0], split[1]
	}
	return "", ""
}

func UpdateIssues(client IssueManager, user string, repo string, issues []*github.Issue, milestoneID int) bool {
	updateErrors := false
	for index := range issues {
		err := client.UpdateIssue(user, repo, *issues[index].Number, milestoneID)
		if err != nil {
			updateErrors = true
			fmt.Printf("Unable to update issue: %s", err.Error())
		}
	}
	return updateErrors
}

func checkMilestoneExists(milestones []*github.Milestone, title *string) *int {
	for index := range milestones {
		if *milestones[index].Title == *title {
			return milestones[index].Number
		}
	}
	return nil
}

func refToVersion(ref string, sha string) string {
	if ref == "refs/heads/master" || ref == "refs/heads/main" {
		return sha
	}
	ref = strings.TrimPrefix(ref, "refs/tags/")
	ref = strings.TrimPrefix(ref, "v")
	version, err := semver.New(ref)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return "unknown"
	}
	return fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
}

type IssueManager interface {
	UpdateIssue(user string, repo string, i int, id int) error
	GetIssues(user string, repo string) ([]*github.Issue, error)
	GetMilestones(user string, repo string) ([]*github.Milestone, error)
	CreateMilestone(user string, repo string, version string) (*int, error)
}