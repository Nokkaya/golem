package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathTraversal_Symlink(t *testing.T) {
	tmpDir := t.TempDir()
	outsideFile := filepath.Join(tmpDir, "secret.txt")
	err := os.WriteFile(outsideFile, []byte("super secret"), 0644)
	if err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	err = os.Mkdir(workspaceDir, 0755)
	if err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	// Create a symlink in workspace pointing to outside file
	symlinkPath := filepath.Join(workspaceDir, "link_to_secret")
	err = os.Symlink(outsideFile, symlinkPath)
	if err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Initialize tool with workspace
	tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	ctx := context.Background()
	// Attempt to read the symlink
	argsJSON := fmt.Sprintf(`{"path": "link_to_secret"}`)

	result, err := tool.InvokableRun(ctx, argsJSON)

	// We expect this to FAIL with access denied
	// If it succeeds, it's a vulnerability
	if err == nil {
		t.Fatalf("SECURITY FAILURE: Was able to read file outside workspace via symlink! Result: %s", result)
	}

	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}
