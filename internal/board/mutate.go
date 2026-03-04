package board

import (
	"fmt"
	"time"

	"github.com/antopolskiy/kanban-md/internal/config"
	"github.com/antopolskiy/kanban-md/internal/task"
)

// DeleteResult is returned after a successful soft-delete (archive).
type DeleteResult struct {
	Task     *task.Task
	Warnings []string // dependent task warnings
}

// Delete soft-deletes (archives) a task. It validates claim ownership and
// collects warnings about dependent tasks. The operation is idempotent —
// archiving an already-archived task is a no-op.
func Delete(cfg *config.Config, id int, claimant string, now time.Time) (*DeleteResult, error) {
	path, err := task.FindByID(cfg.TasksPath(), id)
	if err != nil {
		return nil, err
	}

	t, err := task.Read(path)
	if err != nil {
		return nil, err
	}

	// Validate claim ownership.
	if err := task.CheckClaim(t, claimant, cfg.ClaimTimeoutDuration()); err != nil {
		return nil, err
	}

	// Collect dependent warnings (best-effort).
	warnings := FindDependents(cfg.TasksPath(), id)

	// Idempotent: already archived → no-op.
	if t.Status == config.ArchivedStatus {
		return &DeleteResult{Task: t, Warnings: warnings}, nil
	}

	oldStatus := t.Status
	t.Status = config.ArchivedStatus
	task.UpdateTimestamps(t, oldStatus, t.Status, cfg)
	t.Updated = now

	if err := task.Write(path, t); err != nil {
		return nil, fmt.Errorf("writing task: %w", err)
	}

	LogMutation(cfg.Dir(), "delete", t.ID, t.Title)

	return &DeleteResult{Task: t, Warnings: warnings}, nil
}
