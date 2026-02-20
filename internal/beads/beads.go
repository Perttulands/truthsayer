package beads

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBRCommand = "bd"
	defaultTimeout   = 5 * time.Second

	errorPriority   = 1
	warningPriority = 2
)

type commandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// BeadCreator creates beads by shelling out to the bd CLI.
type BeadCreator struct {
	command string
	timeout time.Duration
	run     commandRunner
}

// NewBeadCreator returns a creator configured to use the default bd binary.
func NewBeadCreator() *BeadCreator {
	return NewBeadCreatorWithCommand(defaultBRCommand)
}

// NewBeadCreatorWithCommand returns a creator using a custom bd command path.
func NewBeadCreatorWithCommand(command string) *BeadCreator {
	if strings.TrimSpace(command) == "" {
		command = defaultBRCommand
	}
	return &BeadCreator{
		command: command,
		timeout: defaultTimeout,
		run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).Output()
		},
	}
}

// SetTimeout sets the command timeout.
func (b *BeadCreator) SetTimeout(d time.Duration) {
	b.timeout = d
}

// CreateProblemBead creates an error-priority bead and returns the bead ID.
func (b *BeadCreator) CreateProblemBead(rule string, file string, count int) (string, error) {
	title := fmt.Sprintf("[truthsayer] %s: %d errors in %s", rule, count, file)
	return b.create(title, errorPriority)
}

// CreateWarningBead creates a warning-priority bead and returns the bead ID.
func (b *BeadCreator) CreateWarningBead(rule string, file string, count int) (string, error) {
	title := fmt.Sprintf("[truthsayer] %s: %d warnings in %s", rule, count, file)
	return b.create(title, warningPriority)
}

func (b *BeadCreator) create(title string, priority int) (string, error) {
	args := []string{"create", "--title", title, "--priority", strconv.Itoa(priority)}

	// REASON: BeadCreator currently has no caller context API; timeout bounds command lifetime.
	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	out, err := b.run(ctx, b.command, args...)
	if err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("bd create timed out: %w", ctx.Err())
		}
		return "", fmt.Errorf("bd create failed: %w", err)
	}

	id := strings.TrimSpace(string(out))
	if id == "" {
		return "", fmt.Errorf("bd create returned empty output")
	}
	return id, nil
}
