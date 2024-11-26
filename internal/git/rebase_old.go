package git

//
//type RebaseOperation int
//
//const (
//	RebaseNormal RebaseOperation = iota
//	RebaseContinue
//	RebaseAbort
//	RebaseSkip
//)
//
//type RebaseConfig struct {
//	Operation RebaseOperation
//	Upstream  string
//	Onto      string
//	Branch    string
//}
//
//type RebaseOutcome int
//
//const (
//	RebaseOutcomeUpToDate RebaseOutcome = iota
//	RebaseOutcomeUpdated
//	RebaseOutcomeConflict
//	RebaseOutcomeNotInProgress
//	RebaseOutcomeAborted
//)
//
//type RebaseResult struct {
//	Outcome      RebaseOutcome
//	Message      string
//	ErrorSummary string
//}
//
//func (r *Repo) RebaseOld(config RebaseConfig) (*RebaseResult, error) {
//	args := buildRebaseArgs(config)
//	env := getRebaseEnv(config)
//
//	fmt.Println("args:", args)
//	fmt.Println("env:", env)
//
//	output, err := r.Run(&RunOpts{
//		Args: args,
//		Env:  env,
//	})
//	if err != nil {
//		return nil, fmt.Errorf("git rebase execution failed: %w", err)
//	}
//
//	fmt.Println("Stdout: ", strings.TrimSpace(string(output.Stdout)))
//	fmt.Println("Stderr: ", strings.TrimSpace(string(output.Stderr)))
//
//	return parseRebaseOutput(config, string(output.Stdout))
//}
//
//func buildRebaseArgs(config RebaseConfig) []string {
//	args := []string{"rebase"}
//
//	switch config.Operation {
//	case RebaseContinue:
//		return append(args, "--continue")
//	case RebaseAbort:
//		return append(args, "--abort")
//	case RebaseSkip:
//		return append(args, "--skip")
//	case RebaseNormal:
//		fallthrough
//	default:
//		if config.Onto != "" {
//			args = append(args, "--onto", config.Onto)
//		}
//		args = append(args, config.Upstream)
//		if config.Branch != "" {
//			args = append(args, config.Branch)
//		}
//	}
//
//	return args
//}
//
//func getRebaseEnv(config RebaseConfig) []string {
//	if config.Operation == RebaseContinue {
//		return []string{"GIT_EDITOR=true"}
//	}
//	return nil
//}
//
//func parseRebaseOutput(config RebaseConfig, output string) (*RebaseResult, error) {
//	lowerOutput := strings.ToLower(output)
//
//	fmt.Println("Contains success", strings.Contains(output, "Successfully rebased"), "output: ", output)
//
//	switch {
//	case strings.Contains(output, "Successfully rebased"):
//		return &RebaseResult{Outcome: RebaseOutcomeUpdated}, nil
//	case strings.Contains(output, "is up to date"):
//		return &RebaseResult{Outcome: RebaseOutcomeUpToDate}, nil
//	case config.Operation == RebaseAbort:
//		return &RebaseResult{Outcome: RebaseOutcomeAborted}, nil
//	case strings.Contains(lowerOutput, "no rebase in progress"):
//		return &RebaseResult{Outcome: RebaseOutcomeNotInProgress}, nil
//	case strings.Contains(lowerOutput, "could not apply"):
//		return &RebaseResult{
//			Outcome:      RebaseOutcomeConflict,
//			Message:      normalizeRebaseMessage(output),
//			ErrorSummary: extractErrorSummary(output),
//		}, nil
//
//	default:
//		return &RebaseResult{
//			Outcome: RebaseOutcomeConflict,
//			Message: output,
//		}, nil
//	}
//}
//
//func normalizeRebaseMessage(message string) string {
//	message = removeCarriageReturns(message)
//	message = removeHintLines(message)
//	return strings.ReplaceAll(message, "git rebase", "zip stack sync")
//}
//
//func removeCarriageReturns(s string) string {
//	re := regexp.MustCompile(`^.+\r`)
//	return re.ReplaceAllString(s, "")
//}
//
//func removeHintLines(s string) string {
//	re := regexp.MustCompile(`(?m)^hint:.+$\n?`)
//	return re.ReplaceAllString(s, "")
//}
//
//func extractErrorSummary(message string) string {
//	re := regexp.MustCompile(`(?m)^error: (.+)$`)
//	matches := re.FindStringSubmatch(message)
//	if len(matches) > 1 {
//		return matches[1]
//	}
//	return ""
//}
