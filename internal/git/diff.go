package git

import "fmt"

type DiffConfig struct {
	Revisions []string
	IsQuiet   bool
	UseColor  bool
	FilePaths []string
}

type DiffResult struct {
	HasDifferences bool
	Content        string
}

func (r *Repo) CalculateDiff(config DiffConfig) (*DiffResult, error) {
	args := buildDiffArgs(config)
	output, err := r.Run(&RunOpts{
		Args: args,
	})
	if err != nil {
		return nil, fmt.Errorf("git diff execution failed: %w", err)
	}

	return parseDiffOutput(output, config.IsQuiet)
}

func buildDiffArgs(config DiffConfig) []string {
	args := []string{"diff", "--exit-code"}

	if config.IsQuiet {
		args = append(args, "--quiet")
	}
	if config.UseColor {
		args = append(args, "--color=always")
	}

	args = append(args, config.Revisions...)
	args = append(args, "--")
	args = append(args, config.FilePaths...)

	return args
}

func parseDiffOutput(output *Output, isQuiet bool) (*DiffResult, error) {
	switch output.ExitCode {
	case 0:
		return &DiffResult{HasDifferences: false, Content: string(output.Stdout)}, nil
	case 1:
		return &DiffResult{HasDifferences: true, Content: string(output.Stdout)}, nil
	default:
		return nil, fmt.Errorf("git diff failed: %s", string(output.Stderr))
	}
}
