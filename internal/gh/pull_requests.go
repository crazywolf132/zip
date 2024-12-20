package gh

import (
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
	"strings"
)

// PullRequest represents a GitHub pull request with relevant information.
type PullRequest struct {
	ID          string
	Number      int
	HeadRefName string
	BaseRefName string
	IsDraft     bool
	Permalink   string
	State       string
	Title       string
	Body        string
	MergeCommit string
}

// HeadBranchName returns the name of the head branch, trimming any "refs/heads/" prefix.
func (p *PullRequest) HeadBranchName() string {
	return strings.TrimPrefix(p.HeadRefName, "refs/heads/")
}

// BaseBranchName returns the name of the base branch, trimming any "refs/heads/" prefix.
func (p *PullRequest) BaseBranchName() string {
	return strings.TrimPrefix(p.BaseRefName, "refs/heads/")
}

// GetMergeCommit returns the merge commit SHA if the pull request is merged, otherwise an empty string.
func (p *PullRequest) GetMergeCommit() string {
	if p.State == "merged" {
		return p.MergeCommit
	}
	return ""
}

// GetPullRequest retrieves a specific pull request by its number.
func (c *Client) GetPullRequest(ctx context.Context, number int) (*PullRequest, error) {
	pr, _, err := c.api.PullRequests.Get(ctx, c.owner, c.repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	return convertToPullRequest(pr), nil
}

// GetPullRequestsInput represents the input parameters for listing pull requests.
type GetPullRequestsInput struct {
	State string
	Head  string
	Base  string
	Sort  string
	Dir   string
}

// GetPullRequests retrieves a list of pull requests based on the given input.
func (c *Client) GetPullRequests(ctx context.Context, input GetPullRequestsInput) ([]*PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State:     input.State,
		Head:      input.Head,
		Base:      input.Base,
		Sort:      input.Sort,
		Direction: input.Dir,
	}

	prs, _, err := c.api.PullRequests.List(ctx, c.owner, c.repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	var result []*PullRequest
	for _, pr := range prs {
		result = append(result, convertToPullRequest(pr))
	}

	return result, nil
}

// CreatePullRequest creates a new pull request.
func (c *Client) CreatePullRequest(ctx context.Context, title, body, head, base string, draft bool) (*PullRequest, error) {
	newPr := &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Head:  &head,
		Base:  &base,
		Draft: &draft,
	}

	pr, _, err := c.api.PullRequests.Create(ctx, c.owner, c.repo, newPr)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return convertToPullRequest(pr), nil
}

// UpdatePullRequest updates an existing pull request.
func (c *Client) UpdatePullRequest(ctx context.Context, number int, title, body *string, state *string) (*PullRequest, error) {
	updatePR := &github.PullRequest{
		Title: title,
		Body:  body,
		State: state,
	}

	pr, _, err := c.api.PullRequests.Edit(ctx, c.owner, c.repo, number, updatePR)
	if err != nil {
		return nil, fmt.Errorf("failed to update pull request: %w", err)
	}

	return convertToPullRequest(pr), nil
}

// RequestReviewers requests reviewers for a pull request.
func (c *Client) RequestReviewers(ctx context.Context, number int, reviewers []string) (*PullRequest, error) {
	pr, _, err := c.api.PullRequests.RequestReviewers(ctx, c.owner, c.repo, number, github.ReviewersRequest{
		Reviewers: reviewers,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request reviews: %w", err)
	}

	return convertToPullRequest(pr), nil
}

// ConvertPullRequestToDraft converts a pull request to a draft.
func (c *Client) ConvertPullRequestToDraft(ctx context.Context, number int) (*PullRequest, error) {
	pr, _, err := c.api.PullRequests.Edit(ctx, c.owner, c.repo, number, &github.PullRequest{
		Draft: github.Bool(true),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to convert pull request to a draft: %w", err)
	}

	return convertToPullRequest(pr), nil
}

// MarkPullRequestReadyForReview marks a pull request as ready for review.
func (c *Client) MarkPullRequestReadyForReview(ctx context.Context, number int) (*PullRequest, error) {
	pr, _, err := c.api.PullRequests.Edit(ctx, c.owner, c.repo, number, &github.PullRequest{
		Draft: github.Bool(false),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to mark pull request as ready for review: %w", err)
	}

	return convertToPullRequest(pr), nil
}

// IsBranchMerged checks if a branch is merged by looking for closed pull requests
func (c *Client) IsBranchMerged(ctx context.Context, branchName string) (bool, error) {
	opts := &github.PullRequestListOptions{
		State: "closed",
		Head:  c.owner + ":" + branchName,
	}

	prs, _, err := c.api.PullRequests.List(ctx, c.owner, c.repo, opts)
	if err != nil {
		return false, fmt.Errorf("failed to list pull requests: %w", err)
	}

	for _, pr := range prs {
		if pr.GetMerged() {
			return true, nil
		}
	}

	return false, nil
}

// convertToPullRequest converts a GitHub pull request to a Stacked pull request.
func convertToPullRequest(pr *github.PullRequest) *PullRequest {
	return &PullRequest{
		ID:          pr.GetNodeID(),
		Number:      pr.GetNumber(),
		HeadRefName: pr.GetHead().GetRef(),
		BaseRefName: pr.GetBase().GetRef(),
		IsDraft:     pr.GetDraft(),
		Permalink:   pr.GetHTMLURL(),
		State:       pr.GetState(),
		Title:       pr.GetTitle(),
		Body:        pr.GetBody(),
		MergeCommit: pr.GetMergeCommitSHA(),
	}
}
