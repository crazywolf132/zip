package git

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"strconv"
	"strings"
	"time"
)

type CommitInfo struct {
	Hash      string
	ShortHash string
	Subject   string
	Body      string
	Timestamp time.Time
}

type LogOptions struct {
	RevisionRange    []string
	SpecificToBranch bool
}

func (r *Repo) FetchGitLog(opts LogOptions) ([]*CommitInfo, error) {
	args := append([]string{"log", "--format=%H%x00%h%x00%s%x00%b%x00%ct%x00"})

	if opts.SpecificToBranch {
		args = append(args, "--no-merges", "--first-parent")
	} else {
		args = append(args, opts.RevisionRange...)
	}

	args = append(args, "--")

	result, err := r.Run(&RunOpts{
		Args: args,
	})
	if err != nil {
		return nil, err
	}
	logrus.WithField("range", opts.RevisionRange).Debug("Fetched git log")

	return parseGitLogOutput(result.Stdout)
}

func parseGitLogOutput(output []byte) ([]*CommitInfo, error) {
	reader := bufio.NewReader(bytes.NewBuffer(output))
	var commits []*CommitInfo

	for {
		commit, err := parseCommit(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

func parseCommit(reader *bufio.Reader) (*CommitInfo, error) {
	fields := make([]string, 5)
	for i := range fields {
		value, err := reader.ReadString('\x00')
		if err != nil {
			return nil, err
		}
		fields[i] = strings.TrimSpace(strings.Trim(value, "\x00"))
	}

	timestamp, err := strconv.ParseInt(fields[4], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse commit timestamp: %w", err)
	}

	return &CommitInfo{
		Hash:      fields[0],
		ShortHash: fields[1],
		Subject:   fields[2],
		Body:      fields[3],
		Timestamp: time.Unix(timestamp, 0),
	}, nil
}

type BranchLog struct {
	Name       string
	IsCurrent  bool
	LastCommit time.Time
	Commits    []*CommitInfo
}

func (r *Repo) GetBranchLogs(branches []string, limit int) ([]BranchLog, error) {
	var logs []BranchLog

	currentBranch, err := r.CurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	for _, branch := range branches {
		commits, err := r.FetchGitLog(LogOptions{
			RevisionRange: []string{branch, fmt.Sprintf("-%d", limit)},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get commits for branch %s: %w", branch, err)
		}

		var lastCommitTime time.Time
		if len(commits) > 0 {
			lastCommitTime, err = r.GetCommitTime(commits[0].Hash)
			if err != nil {
				return nil, fmt.Errorf("failed to get last commit time for branch %s: %w", branch, err)
			}
		}

		logs = append(logs, BranchLog{
			Name:       branch,
			IsCurrent:  branch == currentBranch,
			LastCommit: lastCommitTime,
			Commits:    commits,
		})
	}

	return logs, nil
}

func (r *Repo) GetCommitTime(commitHash string) (time.Time, error) {
	output, err := r.Run(&RunOpts{
		Args: []string{"show", "-s", "--format=%ct", commitHash},
	})
	if err != nil {
		return time.Time{}, err
	}

	println()
	fmt.Println(string(output.Stdout))
	println()

	timestamp, err := strconv.ParseInt(string(output.Stdout), 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse commit timestamp: %w", err)
	}

	return time.Unix(timestamp, 0), nil
}
