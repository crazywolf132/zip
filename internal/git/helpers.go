package git

import (
	"bytes"
	"emperror.dev/errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Git runs git with the given arguments and returns the output as a string.
func (r *Repo) Git(args ...string) (string, error) {
	startTime := time.Now()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.repoDir
	out, err := cmd.Output()
	log := r.log.WithField("duration", time.Since(startTime))
	if err != nil {
		stderr := "<no output>"
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			stderr = string(exitError.Stderr)
		}
		log.Debugf("git %s failed: %s: %s", args, err, stderr)
		return strings.TrimSpace(string(out)), errors.Wrapf(err, "git %s", args[0])
	}

	// trim trailing newline
	log.Debugf("git %s", args)
	return strings.TrimSpace(string(out)), nil
}

func (r *Repo) DoesRefExist(ref string) (bool, error) {
	out, err := r.Run(&RunOpts{
		Args: []string{"show-ref", ref},
	})
	if err != nil {
		return false, errors.Errorf("ref %s does not exist: %v", ref, err)
	}
	if len(out.Stdout) > 0 {
		return true, nil
	}
	return false, nil
}

type RunOpts struct {
	Args []string
	Env  []string

	// Return the error if the command exits with a non-zero exit code.
	ExitError bool

	// If true, the command will be run interactively.
	Interactive bool

	Stdin io.Reader
}

type Output struct {
	ExitCode  int
	ExitError *exec.ExitError
	Stdout    []byte
	Stderr    []byte
}

func (o Output) Lines() []string {
	s := strings.TrimSpace(string(o.Stdout))
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func (r *Repo) Run(opts *RunOpts) (*Output, error) {
	cmd := exec.Command("git", opts.Args...)
	cmd.Dir = r.repoDir
	r.log.Debugf("git %s", opts.Args)
	var stdout, stderr bytes.Buffer
	if opts.Interactive {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}
	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}

	cmd.Env = append(os.Environ(), opts.Env...)
	err := cmd.Run()
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		return nil, errors.Wrapf(err, "git %s", opts.Args)
	}
	if err != nil && opts.ExitError && exitError.ExitCode() != 0 {
		exitError.Stderr = stderr.Bytes()
		return nil, errors.Wrapf(err, "git %s (%s)", opts.Args, stderr.String())
	}

	return &Output{
		ExitCode:  cmd.ProcessState.ExitCode(),
		ExitError: exitError,
		Stdout:    stdout.Bytes(),
		Stderr:    stderr.Bytes(),
	}, nil
}

// GetPRTemplate returns the PR template for the repo
func (r *Repo) GetPRTemplate() (string, error) {
	// Check if the repo has a .github directory
	if _, err := os.Stat(filepath.Join(r.repoDir, ".github")); err != nil {
		return "", err
	}

	// Check if the repo has a .github/pull_request_template.md file
	if _, err := os.Stat(filepath.Join(r.repoDir, ".github", "pull_request_template.md")); err != nil {
		return "", err
	}

	// Read the file contents
	file, err := os.Open(filepath.Join(r.repoDir, ".github", "pull_request_template.md"))
	if err != nil {
		return "", err
	}

	defer file.Close()

	// Read the contents of the file
	contents, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}
