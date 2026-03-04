package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/antopolskiy/kanban-md/internal/board"
	"github.com/antopolskiy/kanban-md/internal/config"
)

// newCreateCmd creates a fresh cobra command with create flags for testing.
func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("title", "", "")
	cmd.Flags().String("status", "", "")
	cmd.Flags().String("priority", "", "")
	cmd.Flags().String("assignee", "", "")
	cmd.Flags().StringSlice("tags", nil, "")
	cmd.Flags().String("due", "", "")
	cmd.Flags().String("estimate", "", "")
	cmd.Flags().Int("parent", 0, "")
	cmd.Flags().IntSlice("depends-on", nil, "")
	cmd.Flags().String("body", "", "")
	cmd.Flags().String("class", "", "")
	cmd.Flags().String("claim", "", "")
	return cmd
}

func TestBuildCreateParams_Status(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("status", "in-progress")

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != "in-progress" {
		t.Errorf("status = %q, want %q", p.Status, "in-progress")
	}
}

func TestBuildCreateParams_Priority(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("priority", priorityHigh)

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Priority != priorityHigh {
		t.Errorf("priority = %q, want %q", p.Priority, priorityHigh)
	}
}

func TestBuildCreateParams_Assignee(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("assignee", "alice")

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Assignee != "alice" {
		t.Errorf("assignee = %q, want %q", p.Assignee, "alice")
	}
}

func TestBuildCreateParams_Tags(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("tags", "bug,urgent")

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Tags) != 2 || p.Tags[0] != "bug" || p.Tags[1] != "urgent" {
		t.Errorf("tags = %v, want [bug, urgent]", p.Tags)
	}
}

func TestBuildCreateParams_Due(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("due", "2025-06-15")

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Due == nil {
		t.Fatal("due should be set")
	}
	if p.Due.Year() != 2025 || p.Due.Month() != 6 || p.Due.Day() != 15 {
		t.Errorf("due = %v, want 2025-06-15", p.Due)
	}
}

func TestBuildCreateParams_InvalidDue(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("due", "not-a-date")

	_, err := buildCreateParams(cmd, "test")
	if err == nil {
		t.Fatal("expected error for invalid due date")
	}
}

func TestBuildCreateParams_Estimate(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("estimate", "4h")

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Estimate != "4h" {
		t.Errorf("estimate = %q, want %q", p.Estimate, "4h")
	}
}

func TestBuildCreateParams_Parent(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("parent", "5")

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Parent == nil || *p.Parent != 5 {
		t.Errorf("parent = %v, want 5", p.Parent)
	}
}

func TestBuildCreateParams_DependsOn(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("depends-on", "2,3")

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.DependsOn) != 2 || p.DependsOn[0] != 2 || p.DependsOn[1] != 3 {
		t.Errorf("depends_on = %v, want [2, 3]", p.DependsOn)
	}
}

func TestBuildCreateParams_Body(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("body", "task description")

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Body != "task description" {
		t.Errorf("body = %q, want %q", p.Body, "task description")
	}
}

func TestBuildCreateParams_Class(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("class", "expedite")

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Class != "expedite" {
		t.Errorf("class = %q, want %q", p.Class, "expedite")
	}
}

func TestBuildCreateParams_Claim(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("claim", "agent-test")

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Claimant != "agent-test" {
		t.Errorf("claimant = %q, want %q", p.Claimant, "agent-test")
	}
}

func TestBuildCreateParams_NoFlags(t *testing.T) {
	cmd := newCreateCmd()

	p, err := buildCreateParams(cmd, "test")
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != "" {
		t.Errorf("status should be empty, got %q", p.Status)
	}
	if p.Priority != "" {
		t.Errorf("priority should be empty, got %q", p.Priority)
	}
}

// --- Validation via board.Create ---

func TestBoardCreate_InvalidStatus(t *testing.T) {
	kanbanDir := setupBoard(t)
	cfg, err := config.Load(kanbanDir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = board.Create(cfg, board.CreateParams{
		Title:  "test",
		Status: "nonexistent",
	}, time.Now())
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestBoardCreate_InvalidPriority(t *testing.T) {
	kanbanDir := setupBoard(t)
	cfg, err := config.Load(kanbanDir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = board.Create(cfg, board.CreateParams{
		Title:    "test",
		Priority: "ultra",
	}, time.Now())
	if err == nil {
		t.Fatal("expected error for invalid priority")
	}
}

func TestBoardCreate_InvalidClass(t *testing.T) {
	kanbanDir := setupBoard(t)
	cfg, err := config.Load(kanbanDir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = board.Create(cfg, board.CreateParams{
		Title: "test",
		Class: "invalid-class",
	}, time.Now())
	if err == nil {
		t.Fatal("expected error for invalid class")
	}
}

func TestRunCreate_Integration(t *testing.T) {
	kanbanDir := setupBoard(t)

	oldFlagDir := flagDir
	flagDir = kanbanDir
	t.Cleanup(func() { flagDir = oldFlagDir })

	setFlags(t, false, true, false)
	r, w := captureStdout(t)

	cmd := newCreateCmd()
	_ = cmd.Flags().Set("priority", priorityHigh)
	_ = cmd.Flags().Set("tags", "test")

	err := runCreate(cmd, []string{"Test task title"})

	got := drainPipe(t, r, w)

	if err != nil {
		t.Fatalf("runCreate error: %v", err)
	}
	if !containsSubstring(got, "Created task #1") {
		t.Errorf("expected 'Created task #1' in output, got: %s", got)
	}
	if !containsSubstring(got, "Test task title") {
		t.Errorf("expected title in output, got: %s", got)
	}

	// Verify the file was created.
	entries, err := os.ReadDir(filepath.Join(kanbanDir, "tasks"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 task file, got %d", len(entries))
	}

	// Verify config was updated.
	cfg, err := config.Load(kanbanDir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.NextID != 2 {
		t.Errorf("NextID = %d, want 2", cfg.NextID)
	}
}

func TestRunCreate_JSONOutput(t *testing.T) {
	kanbanDir := setupBoard(t)

	oldFlagDir := flagDir
	flagDir = kanbanDir
	t.Cleanup(func() { flagDir = oldFlagDir })

	setFlags(t, true, false, false)
	r, w := captureStdout(t)

	cmd := newCreateCmd()
	err := runCreate(cmd, []string{"JSON test"})

	got := drainPipe(t, r, w)

	if err != nil {
		t.Fatalf("runCreate error: %v", err)
	}
	if !containsSubstring(got, `"title": "JSON test"`) {
		t.Errorf("expected JSON output with title, got: %s", got)
	}
}

func TestResolveCreateTitle_Positional(t *testing.T) {
	cmd := newCreateCmd()
	title, err := resolveCreateTitle(cmd, []string{"My task"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title != "My task" {
		t.Errorf("title = %q, want %q", title, "My task")
	}
}

func TestResolveCreateTitle_Flag(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("title", "Flag task")
	title, err := resolveCreateTitle(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title != "Flag task" {
		t.Errorf("title = %q, want %q", title, "Flag task")
	}
}

func TestResolveCreateTitle_BothError(t *testing.T) {
	cmd := newCreateCmd()
	_ = cmd.Flags().Set("title", "Flag task")
	_, err := resolveCreateTitle(cmd, []string{"Positional task"})
	if err == nil {
		t.Fatal("expected error when both positional and --title provided")
	}
}

func TestResolveCreateTitle_NeitherError(t *testing.T) {
	cmd := newCreateCmd()
	_, err := resolveCreateTitle(cmd, nil)
	if err == nil {
		t.Fatal("expected error when no title provided")
	}
}
