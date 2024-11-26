package git

import (
	"emperror.dev/errors"
	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	"path/filepath"
)

var ErrRemoteNotFound = errors.Sentinel("this repository doesn't have a remote origin")

const DEFAULT_REMOTE_NAME = "origin"

type Repo struct {
	repoDir string
	gitDir  string
	gitRepo *git.Repository
	log     logrus.FieldLogger
}

func OpenRepo(repoDir, gitDir string) (*Repo, error) {
	repo, err := git.PlainOpenWithOptions(repoDir, &git.PlainOpenOptions{
		DetectDotGit:          true,
		EnableDotGitCommonDir: true,
	})
	if err != nil {
		return nil, errors.Errorf("failed to open git repo: %v", err)
	}
	r := &Repo{
		repoDir,
		gitDir,
		repo,
		logrus.WithField("repo", filepath.Base(repoDir)),
	}
	return r, nil
}

func (r *Repo) GetRemoteName() string {
	return DEFAULT_REMOTE_NAME
}

type RevParse struct {
	Rev              string
	SymbolicFullName bool
}

func (r *Repo) RevParse(rp *RevParse) (string, error) {
	args := []string{"rev-parse"}
	if rp.SymbolicFullName {
		args = append(args, "--symbolic-full-name")
	}
	args = append(args, rp.Rev)
	return r.Git(args...)
}
