package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathTraversal_ReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	outsideFile := filepath.Join(tmpDir, "outside.txt")
	err := os.WriteFile(outsideFile, []byte("secret content"), 0644)
	if err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	err = os.Mkdir(workspaceDir, 0755)
	if err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	// Now we pass workspaceDir to the tool
	tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	ctx := context.Background()
	// Attempt path traversal using absolute path to outside file
	argsJSON := fmt.Sprintf(`{"path": %q}`, outsideFile)

	result, err := tool.InvokableRun(ctx, argsJSON)

	// Now we expect an ERROR because we fixed it.
	if err == nil {
		t.Fatalf("SECURITY FAILURE: Was able to read file outside workspace! Result: %s", result)
	}

	if !strings.Contains(err.Error(), "access denied") {
	    t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}

func TestPathTraversal_RelativePath(t *testing.T) {
    tmpDir := t.TempDir()
	outsideFile := filepath.Join(tmpDir, "outside.txt")
	err := os.WriteFile(outsideFile, []byte("secret content"), 0644)
	if err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	err = os.Mkdir(workspaceDir, 0755)
	if err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

    tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

    ctx := context.Background()
    // Attempt path traversal using relative path
    argsJSON := fmt.Sprintf(`{"path": "../outside.txt"}`)

    result, err := tool.InvokableRun(ctx, argsJSON)
    if err == nil {
		t.Fatalf("SECURITY FAILURE: Was able to read file outside workspace via relative path! Result: %s", result)
	}
    if !strings.Contains(err.Error(), "access denied") {
	    t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}
