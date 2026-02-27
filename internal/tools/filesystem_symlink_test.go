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
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// 1. Create a secret file OUTSIDE the workspace
	secretFile := filepath.Join(tmpDir, "secret.txt")
	secretContent := "SUPER_SECRET_PASSWORD"
	if err := os.WriteFile(secretFile, []byte(secretContent), 0644); err != nil {
		t.Fatalf("failed to create secret file: %v", err)
	}

	// 2. Create the workspace directory
	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.Mkdir(workspaceDir, 0755); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	// 3. Create a symlink INSIDE the workspace pointing to the secret file
	symlinkPath := filepath.Join(workspaceDir, "link_to_secret")
	if err := os.Symlink(secretFile, symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// 4. Initialize the tool with the workspace
	tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	// 5. Attempt to read the symlink
	// The path passed to the tool is relative to the workspace (or absolute pointing to inside workspace)
	// In this case, we ask to read "link_to_secret" which IS in the workspace.
	ctx := context.Background()
	argsJSON := fmt.Sprintf(`{"path": "link_to_secret"}`)

	result, err := tool.InvokableRun(ctx, argsJSON)

	// 6. Verification
	// If the tool is secure, it should REJECT reading the symlink because it resolves to outside.
	// If it returns the content, it's vulnerable.
	if err == nil {
		if strings.Contains(result, secretContent) {
			t.Fatalf("SECURITY VULNERABILITY: Successfully read file outside workspace via symlink! Result: %s", result)
		}
	} else {
		// If it errors, check if it's the expected access denied error
		t.Logf("Got error as expected (maybe): %v", err)
	}
}
