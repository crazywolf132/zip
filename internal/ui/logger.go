package ui

import (
	"fmt"
	"strings"
	"time"
	"zip/internal/git"
)

type Logger struct {
	logs []git.BranchLog
}

func NewLogger(logs []git.BranchLog) *Logger {
	return &Logger{
		logs,
	}
}

func (l *Logger) FormatBranchLogs() string {
	var output strings.Builder

	for i, log := range l.logs {
		// Branch name and current indicator
		if log.IsCurrent {
			output.WriteString(fmt.Sprintf("◉ %s (current)\n", log.Name))
		} else {
			output.WriteString(fmt.Sprintf("◯ %s\n", log.Name))
		}

		// Last commit time
		output.WriteString(fmt.Sprintf("│ %s\n│\n", l.formatTimeSince(log.LastCommit)))

		// Commits
		for _, commit := range log.Commits {
			output.WriteString(fmt.Sprintf("│ %s - %s\n", commit.ShortHash, commit.Subject))
		}

		// Add a blank line between branches, except for the last one
		if i < len(l.logs)-1 {
			output.WriteString("│\n")
		}
	}

	return output.String()
}

func (l *Logger) formatTimeSince(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	}
}
