package rest

import (
	"context"
	"fmt"
	"github.com/google/go-github/v34/github"
	"golang.org/x/oauth2"
)

type Client struct {
	Token string
	client *github.Client
	ctx context.Context
}

func NewClient(token string) *Client {
	client := &Client{
		Token: token,
	}
	client.client = client.getClient()
	client.ctx = context.Background()
	return client
}

func (c *Client) getClient() *github.Client {
	oauthToken := &oauth2.Token{AccessToken: c.Token}
	staticSource := oauth2.StaticTokenSource(oauthToken)
	oauthClient := oauth2.NewClient(context.Background(), staticSource)
	return github.NewClient(oauthClient)
}

func (c *Client) UpdateIssue(user string, repo string, issueID int, milestoneID int) error {
	update := &github.IssueRequest{
		Milestone: &milestoneID,
	}
	_, _, err := c.client.Issues.Edit(c.ctx, user, repo, issueID, update)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateMilestone(user string, repo string, title string) (*int, error) {
	mileStone := &github.Milestone{
		Title: &title,
	}
	milestone, _, err := c.client.Issues.CreateMilestone(c.ctx, user, repo, mileStone)
	if err != nil {
		return nil, err
	}
	return milestone.Number, nil
}

func (c *Client) GetMilestones(user string, repo string) ([]*github.Milestone, error) {
	opt := &github.MilestoneListOptions{
		State: "all",
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 10,
		},
	}
	var allMilestones []*github.Milestone
	for {
		milestones, resp, err := c.client.Issues.ListMilestones(c.ctx, user, repo, opt)
		if err != nil {
			return nil, err
		}
		allMilestones = append(allMilestones, milestones...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}
	return allMilestones, nil
}

func (c *Client) GetIssues(user string, repo string) ([]*github.Issue, error) {
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			Page:    0,
			PerPage: 10,
		},
	}
	var allIssues []*github.Issue
	for {
		issues, resp, err := c.client.Search.Issues(c.ctx, fmt.Sprintf("is:issue repo:%s/%s state:closed no:milestone", user, repo), opt)
		if err != nil {
			return nil, err
		}
		allIssues = append(allIssues, issues.Issues...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}
	return allIssues, nil
}


