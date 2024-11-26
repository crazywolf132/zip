package git

import (
	"emperror.dev/errors"
	"fmt"
	"strings"
)

type RepoStatus struct {
	Commit          string
	Branch          string
	StagedFiles     []string
	UnstagedFiles   []string
	ConflictedFiles []string
	NewFiles        []string
}

func (s RepoStatus) IsClean(includeUntracked bool) bool {
	if len(s.StagedFiles) > 0 || len(s.UnstagedFiles) > 0 || len(s.ConflictedFiles) > 0 {
		return false
	}

	if includeUntracked && len(s.NewFiles) > 0 {
		return false
	}

	return true
}

func (r *Repo) GetStatus() (*RepoStatus, error) {
	output, err := r.Git("status", "--porcelain=v2", "--branch", "--untracked-files")
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	return parseGitStatus(output)
}

func parseGitStatus(output string) (*RepoStatus, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	status := &RepoStatus{}

	for _, line := range lines {
		if err := parseLine(line, status); err != nil {
			return nil, err
		}
	}

	return status, nil
}

func parseLine(line string, status *RepoStatus) error {
	switch {
	case strings.HasPrefix(line, "# branch.oid "):
		status.Commit = strings.TrimPrefix(line, "# branch.oid ")
		if status.Commit == "(initial)" {
			status.Commit = ""
		}
	case strings.HasPrefix(line, "# branch.head "):
		status.Branch = strings.TrimPrefix(line, "# branch.head ")
		if status.Branch == "(detached)" {
			status.Branch = ""
		}
	case strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 "):
		return parseFileStatus(line, status)
	case strings.HasPrefix(line, "u "):
		return parseUnmergedFile(line, status)
	case strings.HasPrefix(line, "? "):
		status.NewFiles = append(status.NewFiles, strings.TrimPrefix(line, "? "))
	}
	return nil
}

func parseFileStatus(line string, status *RepoStatus) error {
	parts := strings.Fields(line)
	if len(parts) < 9 {
		return errors.New("invalid file status line")
	}

	statusCode := parts[1]
	filename := parts[8]

	switch statusCode[0] {
	case 'M', 'A', 'D', 'R', 'C':
		status.StagedFiles = append(status.StagedFiles, filename)
	}

	switch statusCode[1] {
	case 'M', 'D':
		status.UnstagedFiles = append(status.UnstagedFiles, filename)
	}

	return nil
}

func parseUnmergedFile(line string, status *RepoStatus) error {
	parts := strings.Fields(line)
	if len(parts) < 10 {
		return errors.New("invalid unmerged file line")
	}

	status.ConflictedFiles = append(status.ConflictedFiles, parts[9])
	return nil
}
