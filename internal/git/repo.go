package git

import (
	"emperror.dev/errors"
	giturls "github.com/whilp/git-urls"
	"net/url"
	"path/filepath"
	"strings"
)

func (r *Repo) Dir() string    { return r.repoDir }
func (r *Repo) GitDir() string { return r.gitDir }
func (r *Repo) ZipDir() string { return filepath.Join(r.gitDir, "zip") }

type Origin struct {
	URL *url.URL

	RepoSlug string
}

func (r *Repo) GetOrigin() (*Origin, error) {
	result, err := r.Git("remote", "get-url", "origin")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get origin URL")
	}
	origin := strings.TrimSpace(string(result))
	if origin == "" {
		return nil, errors.New("origin URL is empty")
	}

	u, err := giturls.Parse(origin)
	if err != nil {
		return nil, errors.WrapIff(err, "failed to parse origin url %q", origin)
	}

	repoSlug := strings.TrimSuffix(u.Path, ".git")
	repoSlug = strings.TrimPrefix(repoSlug, "/")
	return &Origin{
		URL:      u,
		RepoSlug: repoSlug,
	}, nil
}

func (r *Repo) RemoteOwnerAndName() (string, string, error) {
	details, err := r.GetOrigin()
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(details.RepoSlug, "/")
	if len(parts) != 2 {
		return "", "", errors.New("unexpected format")
	}

	return parts[0], parts[1], nil
}

func (r *Repo) Fetch() error {
	_, err := r.Git("fetch", "--all", "--prune")
	return err
}
