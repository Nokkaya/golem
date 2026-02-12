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
	err := os.WriteFile(outsideFile, []byte("SUPER SECRET"), 0644)
	if err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	err = os.Mkdir(workspaceDir, 0755)
	if err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	// Create a symlink inside the workspace pointing to the outside file
	symlinkPath := filepath.Join(workspaceDir, "link_to_secret")
	err = os.Symlink(outsideFile, symlinkPath)
	if err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	ctx := context.Background()
	// Attempt to read the symlink
	argsJSON := fmt.Sprintf(`{"path": "link_to_secret"}`)

	result, err := tool.InvokableRun(ctx, argsJSON)

	if err == nil {
		t.Fatalf("SECURITY FAILURE: Was able to read file outside workspace via symlink! Result: %s", result)
	}

    // Expect "access denied" error
    if !strings.Contains(err.Error(), "access denied") {
        t.Fatalf("Expected 'access denied' error, got: %v", err)
    }
}
