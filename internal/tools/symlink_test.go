package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathTraversal_Symlink(t *testing.T) {
	// 1. Create a temporary directory structure
	// /tmp/root/
	//   secret.txt (outside workspace)
	//   workspace/
	//     link_to_secret -> ../secret.txt

	tmpDir := t.TempDir()
	secretFile := filepath.Join(tmpDir, "secret.txt")
	secretContent := "CRITICAL_SECRET_DATA"
	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("failed to create secret file: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	err = os.Mkdir(workspaceDir, 0755)
	if err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	// Create symlink inside workspace pointing to outside file
	linkPath := filepath.Join(workspaceDir, "link_to_secret")
	err = os.Symlink(secretFile, linkPath)
	if err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// 2. Initialize the tool with the workspace
	tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	// 3. Attempt to read the symlink
	ctx := context.Background()
	// We use the relative path "link_to_secret" which is inside the workspace
	argsJSON := `{"path": "link_to_secret"}`

	result, err := tool.InvokableRun(ctx, argsJSON)

	// 4. Verification
	// If the vulnerability is fixed, we should get an "access denied" error.

	if err == nil {
		t.Fatalf("SECURITY FAILURE: Was able to read file outside workspace via symlink! Result: %s", result)
	}

	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}
