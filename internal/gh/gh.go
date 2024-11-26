package gh

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"os/exec"
	"strings"
	"sync"

	"github.com/google/go-github/v62/github"
)

type Client struct {
	api   *github.Client
	token string
	ctx   context.Context
	owner string
	repo  string
}

var (
	instance *Client
	once     sync.Once
)

// NewClient creates and returns a new GitHub client.
func NewClient(owner, repo string) (*Client, error) {
	var err error
	once.Do(func() {
		instance, err = initClient(owner, repo)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Github client: %w", err)
	}
	return instance, nil
}

func initClient(owner, repo string) (*Client, error) {
	token, err := getGitHubToken()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		api:   github.NewClient(tc),
		token: token,
		ctx:   ctx,
		owner: owner,
		repo:  repo,
	}, nil
}

func getGitHubToken() (string, error) {
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return "", fmt.Errorf("GitHub CLI not found: %w", err)
	}

	out, err := exec.Command(ghPath, "auth", "token").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub token: %w", err)
	}

	token := strings.TrimSpace(string(out))
	if token == "" {
		return "", fmt.Errorf("empty GitHub token received")
	}

	return token, nil
}

// GetOwner returns the owner of the repository.
func (c *Client) GetOwner() string {
	return c.owner
}

// GetRepo returns the name of the repository.
func (c *Client) GetRepo() string {
	return c.repo
}

// GetContext returns the context used by the client.
func (c *Client) GetContext() context.Context {
	return c.ctx
}

// GetAPIClient returns the underlying GitHub API client.
func (c *Client) GetAPIClient() *github.Client {
	return c.api
}
