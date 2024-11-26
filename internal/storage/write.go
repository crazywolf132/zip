package storage

type WriteTx struct {
	db     *Database
	ReadTx *ReadTx
}

func (tx *WriteTx) SetRepository(repo Repository) {
	tx.db.state.Repository = repo
}

func (tx *WriteTx) SetBranch(branch Branch) {
	tx.db.state.Branches[branch.Name] = branch
}

func (tx *WriteTx) DeleteBranch(name string) {
	delete(tx.db.state.Branches, name)
}

func (tx *WriteTx) SetStack(stack Stack) {
	tx.db.state.Stacks[stack.Name] = stack
}

func (tx *WriteTx) SetCurrentStack(name string) {
	tx.db.state.Repository.CurrentStack = name
}

func (tx *WriteTx) DeleteStack(name string) {
	delete(tx.db.state.Stacks, name)
}

func (tx *WriteTx) Commit() error {
	err := tx.db.save()
	//tx.db.mu.Unlock()
	return err
}

func (tx *WriteTx) Abort() {
	tx.db.mu.Unlock()
}
