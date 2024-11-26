package gh

import (
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
	"strings"
)

const stackCommentIdentifier = "This stack of pull requests is managed by zip."

// FindStackComment searches for the stack comment in a pull request
func (c *Client) FindStackComment(ctx context.Context, number int) (*github.IssueComment, error) {
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		comments, resp, err := c.api.Issues.ListComments(ctx, c.owner, c.repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list comments: %w", err)
		}

		for _, comment := range comments {
			if comment.Body != nil && strings.Contains(*comment.Body, stackCommentIdentifier) {
				return comment, nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil, nil // Comment not found
}

// RemoveComment removes a specific comment from a pull request
func (c *Client) RemoveComment(ctx context.Context, commentID int64) error {
	_, err := c.api.Issues.DeleteComment(ctx, c.owner, c.repo, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	return nil
}

// AddComment adds a new comment to a pull request
func (c *Client) AddComment(ctx context.Context, number int, body string) (*github.IssueComment, error) {
	comment, _, err := c.api.Issues.CreateComment(ctx, c.owner, c.repo, number, &github.IssueComment{Body: &body})
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}
	return comment, nil
}
