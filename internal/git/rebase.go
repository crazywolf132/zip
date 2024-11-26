package git

import (
	"regexp"
	"strings"
)

type RebaseOperation int

const (
	RebaseNormal RebaseOperation = iota
	RebaseContinue
	RebaseAbort
	RebaseSkip
)

type RebaseConfig struct {
	// Required (unless Continue is true)
	// The upstream branch to rebase onto.
	Upstream string
	// Operation to perform
	Operation RebaseOperation
	Onto      string
	// Optional
	// If set, this is the branch that will be rebased; otherwise, the current
	// branch is rebased
	Branch string
}

func (r *Repo) Rebase(opts RebaseConfig) (*RebaseResult, error) {
	args := []string{"rebase"}
	var env []string

	switch opts.Operation {
	case RebaseContinue:
		args = append(args, "--continue")
		env = append(env, "GIT_EDITOR=true")
	case RebaseAbort:
		args = append(args, "--abort")
	case RebaseSkip:
		args = append(args, "--skip")
	case RebaseNormal:
		fallthrough
	default:
		if opts.Onto != "" {
			args = append(args, "--onto", opts.Onto)
		}
		args = append(args, opts.Upstream)
		if opts.Branch != "" {
			args = append(args, opts.Branch)
		}
	}

	out, err := r.Run(&RunOpts{Args: args})
	if err != nil {
		return nil, err
	}
	return parseRebaseResult(opts, out)
}

type RebaseStatus int

const (
	RebaseAlreadyUpToDate RebaseStatus = iota
	RebaseUpdated
	RebaseConflict
	RebaseNotInProgress
	RebaseAborted
)

type RebaseResult struct {
	Status        RebaseStatus
	Hint          string
	ErrorHeadline string
}

var carriageReturnRegex = regexp.MustCompile(`^.+\r`)
var hintRegex = regexp.MustCompile(`(?m)^hint:.+$\n?`)
var errorMatchRegex = regexp.MustCompile(`(?m)^error: (.+)$`)

func normalizeRebaseHint(stderr []byte) string {
	res := string(stderr)
	res = carriageReturnRegex.ReplaceAllString(res, "")
	res = hintRegex.ReplaceAllString(res, "")
	res = strings.ReplaceAll(res, "git rebase", "zip stack sync")
	return res
}

func parseRebaseResult(opts RebaseConfig, out *Output) (*RebaseResult, error) {
	stdout := string(out.Stdout)
	stderr := string(out.Stderr)

	if out.ExitCode == 0 {
		if strings.Contains(stderr, "Successfully rebased") {
			return &RebaseResult{Status: RebaseUpdated}, nil
		}
		if strings.Contains(stdout, "is up to date") {
			return &RebaseResult{Status: RebaseAlreadyUpToDate}, nil
		}

		if opts.Operation == RebaseAbort {
			return &RebaseResult{Status: RebaseAborted}, nil
		}

		return &RebaseResult{Status: RebaseUpdated}, nil
	}

	var status RebaseStatus
	lowerStderr := strings.ToLower(stderr)
	switch {
	case strings.Contains(lowerStderr, "no rebase in progress"):
		status = RebaseNotInProgress
	case strings.Contains(lowerStderr, "could not apply"):
		status = RebaseConflict
	default:
		return &RebaseResult{
			Status: RebaseConflict,
			Hint:   stderr,
		}, nil
	}

	hint := normalizeRebaseHint(out.Stderr)
	headline := ""
	errorMatches := errorMatchRegex.FindStringSubmatch(hint)
	if len(errorMatches) > 1 {
		headline = errorMatches[1]
	}

	return &RebaseResult{
		Status:        status,
		Hint:          hint,
		ErrorHeadline: headline,
	}, nil
}
