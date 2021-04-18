package graphql

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Client struct {
	Token  string
	client *githubv4.Client
}

func NewClient(token string) *Client {
	client := &Client{
		Token: token,
	}
	client.client = client.getClient()
	return client
}

func (c *Client) getClient() *githubv4.Client {
	oauthToken := &oauth2.Token{AccessToken: c.Token}
	staticSource := oauth2.StaticTokenSource(oauthToken)
	oauthClient := oauth2.NewClient(context.Background(), staticSource)
	return githubv4.NewClient(oauthClient)
}

func (c *Client) updateIssue(id githubv4.ID, milestone githubv4.ID) error {
	parsedID := &id
	if id == "" {
		parsedID = nil
	}
	parsedMilestone := &milestone
	if milestone == "" {
		parsedMilestone = nil
	}
	input := githubv4.UpdateIssueInput{
		ID:          githubv4.NewID(parsedID),
		MilestoneID: githubv4.NewID(parsedMilestone),
	}
	err := c.client.Mutate(context.Background(), &updateIssueMilestone, input, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) getIssues(user string, repo string) ([]githubv4.ID, error) {
	var allIssues []githubv4.ID
	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil),
		"query":  githubv4.String(fmt.Sprintf("is:issue user:%s repo:%s state:closed no:milestone", user, repo)),
	}
	for {
		err := c.client.Query(context.Background(), &getIssuesQuery, variables)
		if err != nil {
			return nil, err
		}
		for _, edge := range getIssuesQuery.Search.Edges {
			allIssues = append(allIssues, edge.Node.Issue.ID)
		}
		if !getIssuesQuery.Search.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(getIssuesQuery.Search.PageInfo.EndCursor)
	}
	return allIssues, nil
}

func (c *Client) getMilestones(user string, repo string) ([]githubv4.ID, error) {
	var allMilestones []githubv4.ID
	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil),
		"owner":  githubv4.String(user),
		"name":   githubv4.String(repo),
	}
	for {
		err := c.client.Query(context.Background(), &getMilestoneQuery, variables)
		if err != nil {
			return nil, err
		}
		for _, edge := range getMilestoneQuery.Repository.Milestones.Edges {
			allMilestones = append(allMilestones, edge.Node.ID)
		}
		if !getMilestoneQuery.Repository.Milestones.PageInfo.HasNextPage {
			break
		}
		variables["cursor"] = githubv4.NewString(getMilestoneQuery.Repository.Milestones.PageInfo.EndCursor)
	}
	return allMilestones, nil
}

func (c *Client) checkMilestoneExists(user string, repo string, title string) (bool, error) {
	milestones, err := c.getMilestones(user, repo)
	if err != nil {
		return false, err
	}
	for index := range milestones {
		if milestones[index] == title {
			return true, nil
		}
	}
	return false, nil
}

type pageInfo struct {
	StartCursor githubv4.String
	EndCursor   githubv4.String
	HasNextPage githubv4.Boolean
}

type Milestone struct {
	ID    githubv4.ID
	Title string
}

type Issue struct {
	ID        githubv4.ID
	Milestone struct {
		ID githubv4.ID
	}
}

var getMilestoneQuery struct {
	Repository struct {
		Milestones struct {
			PageInfo pageInfo
			Edges    []struct {
				Node Milestone
			}
		} `graphql:"milestones(first: 10, after: $cursor)"`
	} `graphql:"repository(name: $name, owner: $owner)"`
}

var getIssuesQuery struct {
	Search struct {
		PageInfo pageInfo
		Edges    []struct {
			Node struct {
				Issue `graphql:"... on Issue"`
			}
		}
	} `graphql:"search(first: 50, type: ISSUE, query: $query, after: $cursor)"`
}

var updateIssueMilestone struct {
	UpdateIssue struct {
		ClientMutationID string
	} `graphql:"updateIssue(input: $input)"`
}
