package storage

import (
	"fmt"
	"time"
)

func (tx *WriteTx) GetCurrentStack() (Stack, bool) {
	if tx.db.state.Repository.CurrentStack == "" {
		return Stack{}, false
	}
	return tx.db.state.Stacks[tx.db.state.Repository.CurrentStack], true
}

func (tx *WriteTx) CreateStack(name, creator, baseBranch string, branches []string) (*Stack, error) {
	if _, exists := tx.db.state.Stacks[name]; exists {
		return nil, fmt.Errorf("stack %s already exists", name)
	}

	stack := Stack{
		Name:        name,
		Creator:     creator,
		BaseBranch:  baseBranch,
		CreatedDate: time.Now(),
		Branches:    branches,
	}

	tx.db.state.Stacks[name] = stack
	return &stack, nil
}

func (tx *WriteTx) AddBranchToStack(stackName, branchName string) error {
	stack, exists := tx.db.state.Stacks[stackName]
	if !exists {
		return fmt.Errorf("stack %s does not exist", stackName)
	}

	for _, branch := range stack.Branches {
		if branch == branchName {
			return fmt.Errorf("branch %s already exists in stack %s", branchName, stackName)
		}
	}

	stack.Branches = append(stack.Branches, branchName)
	tx.db.state.Stacks[stackName] = stack
	return nil
}

func (tx *WriteTx) RemoveBranchFromStack(stackName, branchName string) error {
	stack, exists := tx.db.state.Stacks[stackName]
	if !exists {
		return fmt.Errorf("stack %s does not exist", stackName)
	}

	for i, branch := range stack.Branches {
		if branch == branchName {
			stack.Branches = append(stack.Branches[:i], stack.Branches[i+1:]...)
			tx.db.state.Stacks[stackName] = stack
			return nil
		}
	}

	return fmt.Errorf("branch %s does not exist in stack %s", branchName, stackName)
}
