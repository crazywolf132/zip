package storage

import (
	"fmt"
	"slices"
)

type ReadTx struct {
	db *Database
}

func (tx *ReadTx) Repository() Repository {
	return tx.db.state.Repository
}

func (tx *ReadTx) Branch(name string) (Branch, bool) {
	branch, ok := tx.db.state.Branches[name]
	return branch, ok
}

func (tx *ReadTx) AllBranches() map[string]Branch {
	branches := make(map[string]Branch)
	for name, branch := range tx.db.state.Branches {
		branches[name] = branch
	}
	return branches
}
func (tx *ReadTx) Stack(name string) (Stack, bool) {
	stack, ok := tx.db.state.Stacks[name]
	return stack, ok
}

func (tx *ReadTx) AllStacks() map[string]Stack {
	stacks := make(map[string]Stack)
	for name, stack := range tx.db.state.Stacks {
		stacks[name] = stack
	}
	return stacks
}

func (tx *ReadTx) CurrentStack() (Stack, bool) {
	if tx.db.state.Repository.CurrentStack == "" {
		return Stack{}, false
	}
	stack, ok := tx.db.state.Stacks[tx.db.state.Repository.CurrentStack]
	return stack, ok
}

func (tx *ReadTx) FindStackByBranch(branchName string) (Stack, bool) {
	for _, stack := range tx.db.state.Stacks {
		for _, branch := range stack.Branches {
			if branch == branchName {
				return stack, true
			}
		}
	}
	return Stack{}, false
}

func (tx *ReadTx) AllStackBranches() (map[string]Branch, error) {
	// Getting the stack first.
	stack, exists := tx.CurrentStack()
	if !exists {
		return nil, fmt.Errorf("no active stack. Please create or switch to a stack first")
	}
	// Getting all branches in the stack.
	branches := make(map[string]Branch)
	for _, branchName := range stack.Branches {
		branch, ok := tx.Branch(branchName)
		if !ok && branchName != stack.BaseBranch {
			return nil, fmt.Errorf("branch %s does not exist", branch)
		}
		branches[branch.Name] = branch
	}
	return branches, nil
}

func (tx *ReadTx) GetParent(branchName string) (Branch, bool) {
	branch, ok := tx.Branch(branchName)
	if !ok {
		return Branch{}, false
	}
	parent, ok := tx.Branch(branch.Parent.Name)
	if !ok {
		return Branch{}, false
	}
	return parent, true
}

func (tx *ReadTx) GetHeritage(branchName string) ([]Branch, error) {
	var heritage []Branch
	currentBranch, exists := tx.Branch(branchName)
	if !exists {
		return []Branch{}, fmt.Errorf("failed to get branch: %s", branchName)
	}

	heritage = append(heritage, currentBranch)

	lookingAt := currentBranch
	for {
		if lookingAt.Parent.Trunk {
			heritage = append(heritage, Branch{Name: lookingAt.Parent.Name})
			break
		}

		lookingAt, exists = tx.Branch(lookingAt.Parent.Name)
		if !exists {
			return []Branch{}, fmt.Errorf("failed to get branch: %s", lookingAt.Parent.Name)
		}

		heritage = append(heritage, lookingAt)
	}

	return heritage, nil
}

func (tx *ReadTx) ChildrenBranches(branchName string) []Branch {
	branches := tx.AllBranches()
	var children []Branch
	for _, branch := range branches {
		if branch.Parent.Name == branchName {
			children = append(children, branch)
		}
	}

	// Sort for determinism.
	slices.SortFunc(children, func(a, b Branch) int {
		if a.CreatedDate.Before(b.CreatedDate) {
			return -1
		} else if a.CreatedDate.After(b.CreatedDate) {
			return 1
		}
		return 0
	})
	return children
}

func (tx *ReadTx) GetOrderedStackBranches(stackName string) ([]Branch, error) {
	stack, exists := tx.Stack(stackName)
	if !exists {
		return nil, fmt.Errorf("stack %s does not exist", stackName)
	}

	branches := make([]Branch, 0, len(stack.Branches))
	branchMap := make(map[string]Branch)

	// First, get all branches and create a map
	for _, branchName := range stack.Branches {
		if branchName == stack.BaseBranch {
			continue
		}
		branch, exists := tx.Branch(branchName)
		if !exists {
			return nil, fmt.Errorf("branch %s in stack %s does not exist", branchName, stackName)
		}
		branchMap[branchName] = branch
		branches = append(branches, branch)
	}

	// Sort branches based on their parent-child relationships
	slices.SortFunc(branches, func(a, b Branch) int {
		if a.Name == stack.BaseBranch {
			return -1
		}
		if b.Name == stack.BaseBranch {
			return 1
		}
		if a.Parent.Name == b.Name {
			return 1
		}
		if b.Parent.Name == a.Name {
			return -1
		}
		return 0
	})

	// Verify that the order is correct
	for i, branch := range branches {
		if i == 0 {
			if !branch.Parent.Trunk {
				return nil, fmt.Errorf("first branch is not the base branch")
			}
		} else {
			parentBranch, exists := branchMap[branch.Parent.Name]
			if !exists {
				return nil, fmt.Errorf("parent branch %s of %s does not exist in the stack", branch.Parent.Name, branch.Name)
			}
			if slices.Index(branches[:i], parentBranch) == -1 {
				return nil, fmt.Errorf("branch %s appears before its parent %s", branch.Name, parentBranch.Name)
			}
		}
	}

	return branches, nil
}

func (tx *ReadTx) Close() {
	tx.db.mu.RUnlock()
}
