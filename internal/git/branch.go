package git

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

func (r *Repo) CurrentBranch() (string, error) {
	branch, err := r.Git("symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", errors.Wrap(err, "failed to determine current branch")
	}
	return branch, nil
}

type SwitchOpts struct {
	// If true, create the branch if it doesn't exist.
	Create bool
	// Name of the new branch.
	Name string
	// Starting point for the new branch.
	NewHeadRef string
}

func (r *Repo) Switch(opts *SwitchOpts) (string, error) {
	previousBranch, er := r.CurrentBranch()
	if er != nil {
		return "", er
	}
	args := []string{"switch"}
	if opts.Create {
		args = append(args, "-c")
	}
	args = append(args, opts.Name)

	if opts.NewHeadRef != "" {
		args = append(args, opts.NewHeadRef)
	}

	result, err := r.Run(&RunOpts{
		Args: args,
	})
	if err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		logrus.WithFields(logrus.Fields{
			"stdout": string(result.Stdout),
			"stderr": string(result.Stderr),
		}).Debug("git switch failed")
		return "", errors.Errorf("failed to switch branch: %q: %s", opts.Name, string(result.Stderr))
	}
	return previousBranch, err
}

func (r *Repo) DefaultBranch() (string, error) {
	ref, err := r.Git("symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		logrus.WithError(err).Debug("failed to determine remote HEAD")
		return "", errors.New("failed to determine remote HEAD")
	}
	return strings.TrimPrefix(ref, "refs/remotes/origin/"), nil
}

func (r *Repo) BranchExists(name string, remote bool) (bool, error) {
	if remote {
		return r.DoesRefExist(fmt.Sprintf("refs/remotes/origin/%s", name))
	}
	return r.DoesRefExist(fmt.Sprintf("refs/heads/%s", name))
}

func (r *Repo) IsTrunkBranch(name string) (bool, error) {
	defaultBranch, err := r.DefaultBranch()
	if err != nil {
		return false, err
	}
	return name == defaultBranch, nil
}

func (r *Repo) IsCurrentBranchTrunk() (bool, error) {
	currentBranch, err := r.CurrentBranch()
	if err != nil {
		return false, err
	}
	return r.IsTrunkBranch(currentBranch)
}

func (r *Repo) GetLastCommit() (string, error) {
	res, err := r.Run(&RunOpts{
		Args: []string{"rev-parse", "HEAD"},
	})
	if err != nil {
		return "", errors.WrapIff(err, "failed to get last commit: %w", err)
	}
	if res.ExitCode != 0 {
		return "", errors.Errorf("failed to get last commit: %s", res.Stderr)
	}
	return string(res.Stdout), nil
}

// Pull fetches and merges changes from the remote for the specified branch
func (r *Repo) Pull(branchName string) error {
	currentBranch, err := r.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch != branchName {
		if _, err := r.Switch(&SwitchOpts{Name: branchName}); err != nil {
			return fmt.Errorf("failed to switch to branch %s: %w", branchName, err)
		}
		defer r.Switch(&SwitchOpts{Name: currentBranch}) // Switch back to the original branch
	}

	output, err := r.Run(&RunOpts{
		Args: []string{"pull", "--ff-only", "origin", branchName},
	})
	if err != nil {
		return fmt.Errorf("failed to pull changes for branch %s: %w", branchName, err)
	}

	if strings.Contains(string(output.Stdout), "Already up to date.") {
		return nil
	}

	return nil
}

// PushNewBranch pushes a new branch to the remote for the first time
func (r *Repo) PushNewBranch(branchName string) error {
	args := []string{"push", "--set-upstream", "origin", branchName}
	_, err := r.Run(&RunOpts{Args: args})
	if err != nil {
		return fmt.Errorf("failed to push new branch %s: %w", branchName, err)
	}
	return nil
}

// FastForwardBranch fast-forwards the current branch to the specified ref
func (r *Repo) FastForwardBranch(ref string) error {
	args := []string{"merge", "--ff-only", ref}
	_, err := r.Run(&RunOpts{Args: args})
	if err != nil {
		return fmt.Errorf("failed to fast-forward branch to %s: %w", ref, err)
	}
	return nil
}

// PushWithForceWithLease pushes the current branch with --force-with-lease
func (r *Repo) PushWithForceWithLease(branchName string) error {
	args := []string{"push", "--force-with-lease", "origin", branchName}
	_, err := r.Run(&RunOpts{Args: args})
	if err != nil {
		return fmt.Errorf("failed to push branch %s with force-with-lease: %w", branchName, err)
	}
	return nil
}

// IsBranchMerged checks if a branch is merged into the default branch
func (r *Repo) IsBranchMerged(branchName string) (bool, error) {
	defaultBranch, err := r.DefaultBranch()
	if err != nil {
		return false, fmt.Errorf("failed to get default branch: %w", err)
	}

	args := []string{"branch", "--merged", defaultBranch}
	output, err := r.Run(&RunOpts{Args: args})
	if err != nil {
		return false, fmt.Errorf("failed to check if branch %s is merged: %w", branchName, err)
	}

	mergedBranches := strings.Split(strings.TrimSpace(string(output.Stdout)), "\n")
	for _, branch := range mergedBranches {
		if strings.TrimSpace(branch) == branchName {
			return true, nil
		}
	}
	return false, nil
}

// AreChangesInBranch checks if all commits from sourceBranch are present in targetBranch
func (r *Repo) AreChangesInBranch(targetBranch, sourceBranch string) (bool, error) {
	// Get the merge base of the two branches
	mergeBase, err := r.Run(&RunOpts{
		Args: []string{"merge-base", targetBranch, sourceBranch},
	})
	if err != nil {
		return false, fmt.Errorf("failed to find merge base: %w", err)
	}

	// Get the diff between merge-base and sourceBranch
	diff, err := r.Run(&RunOpts{
		Args: []string{"diff", "--name-only", strings.TrimSpace(string(mergeBase.Stdout)), sourceBranch},
	})
	if err != nil {
		return false, fmt.Errorf("failed to get diff: %w", err)
	}

	// If there's no diff, all changes from sourceBranch are in targetBranch
	return len(diff.Stdout) == 0, nil
}

// DeleteBranch deletes a branch both locally and remotely
func (r *Repo) DeleteBranch(branchName string) error {
	// Delete the local branch
	_, err := r.Run(&RunOpts{
		Args: []string{"branch", "-D", branchName},
	})
	if err != nil {
		return fmt.Errorf("failed to delete local branch %s: %w", branchName, err)
	}

	// Check if the branch exists on the remote
	remoteBranchExists, err := r.BranchExists(branchName, true)
	if err != nil {
		return fmt.Errorf("failed to check if remote branch %s exists: %w", branchName, err)
	}

	// If the branch exists on the remote, delete it
	if remoteBranchExists {
		_, err = r.Run(&RunOpts{
			Args: []string{"push", "origin", "--delete", branchName},
		})
		if err != nil {
			return fmt.Errorf("failed to delete remote branch %s: %w", branchName, err)
		}
	}

	return nil
}

// AV - push.go L#280
func (r *Repo) Push(branchName, remoteCommit string) error {
	pushArgs := []string{"push", r.GetRemoteName(), "--atomic", fmt.Sprintf("--force-with-lease=%s:%s", branchName, remoteCommit)}
	res, err := r.Run(&RunOpts{
		Args: pushArgs,
	})
	if err != nil {
		return errors.WrapIff(err, "failed to push branch to GitHub: %s\n%w", branchName, err)
	}
	if res.ExitCode != 0 {
		return errors.Errorf("failed to push branch to GitHub: %s\n%s ", branchName, res.Stderr)
	}
	return nil
}
