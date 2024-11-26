package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"zip/internal/gh"
)

type Database struct {
	filePath string
	mu       sync.RWMutex
	state    *State
}

type State struct {
	Repository Repository        `json:"repository"`
	Branches   map[string]Branch `json:"branches"`
	Stacks     map[string]Stack  `json:"stacks"`
}

type Repository struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	Name         string `json:"name"`
	CurrentStack string `json:"current_stack"`
}

type Branch struct {
	Name        string       `json:"name"`
	CreatedDate time.Time    `json:"created_date"`
	Parent      BranchState  `json:"parent"`
	PullRequest *PullRequest `json:"pull_request,omitempty"`
	MergeCommit string       `json:"merge_commit,omitempty"`
}

type BranchState struct {
	Name  string `json:"name"`
	Trunk bool   `json:"trunk,omitempty"`
	Head  string `json:"head,omitempty"`
}

type PullRequest struct {
	ID          string `json:"id"`
	Number      int    `json:"number"`
	Permalink   string `json:"permalink"`
	State       string `json:"state"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	IsDraft     bool   `json:"is_draft"`
	MergeCommit string `json:"merge_commit"`
}

type Stack struct {
	Name        string    `json:"name"`
	Creator     string    `json:"creator"`
	CreatedDate time.Time `json:"created_date"`
	BaseBranch  string    `json:"base_branch"`
	Branches    []string  `json:"branches"`
}

func OpenDatabase(path string) (*Database, bool, error) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create directory: %w", err)
	}

	db := &Database{
		filePath: path,
		state:    &State{Branches: make(map[string]Branch), Stacks: make(map[string]Stack)},
	}

	exists, err := db.load()
	if err != nil {
		return nil, false, fmt.Errorf("failed to load database: %w", err)
	}

	return db, exists, nil
}

func (db *Database) load() (bool, error) {
	data, err := os.ReadFile(db.filePath)
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	err = json.Unmarshal(data, db.state)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return true, nil
}

func (db *Database) save() error {
	data, err := json.MarshalIndent(db.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database: %w", err)
	}

	err = os.WriteFile(db.filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write database: %w", err)
	}

	return nil
}

func (db *Database) ReadTx() *ReadTx {
	db.mu.RLock()
	return &ReadTx{db: db}
}

func (db *Database) WriteTx() *WriteTx {
	db.mu.Lock()
	return &WriteTx{db: db, ReadTx: &ReadTx{db: db}}
}

func MakePRData(pr *gh.PullRequest) *PullRequest {
	return &PullRequest{
		ID:          pr.ID,
		Number:      pr.Number,
		Permalink:   pr.Permalink,
		State:       pr.State,
		Title:       pr.Title,
		Body:        pr.Body,
		IsDraft:     pr.IsDraft,
		MergeCommit: pr.MergeCommit,
	}
}
