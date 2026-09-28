package storage

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
)

// DesktopLocationOpener implements LocationOpener for desktop operating systems.
type DesktopLocationOpener struct{}

// NewDesktopLocationOpener creates a DesktopLocationOpener.
func NewDesktopLocationOpener() *DesktopLocationOpener {
	return &DesktopLocationOpener{}
}

func (o *DesktopLocationOpener) OpenLocation(ctx context.Context, loc Location) error {
	if loc.Kind != LocationDesktopPath || loc.Path == "" {
		return errors.New("cannot open non-desktop location in file manager")
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", loc.Path)
	case "windows":
		cmd = exec.CommandContext(ctx, "explorer", loc.Path)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", loc.Path)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file manager: %w", err)
	}
	return nil
}
