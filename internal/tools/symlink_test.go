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

	// Create a symlink inside workspace pointing to outside
	symlinkPath := filepath.Join(workspaceDir, "link_to_outside.txt")
	err = os.Symlink(outsideFile, symlinkPath)
	if err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	tool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}

	ctx := context.Background()
	// Attempt to read via symlink
	argsJSON := fmt.Sprintf(`{"path": "link_to_outside.txt"}`)

	result, err := tool.InvokableRun(ctx, argsJSON)

	// We expect an error if security is working.
	// If it succeeds, it's a vulnerability.
	if err == nil {
		t.Fatalf("SECURITY FAILURE: Was able to read file outside workspace via symlink! Result: %s", result)
	}

	if !strings.Contains(err.Error(), "access denied") && !strings.Contains(err.Error(), "symlink") {
		// Accept "access denied" or explicit symlink error
		t.Logf("Got error as expected: %v", err)
	}
}

func TestPathTraversal_WriteFile_Symlink(t *testing.T) {
	tmpDir := t.TempDir()
	outsideFile := filepath.Join(tmpDir, "outside_write.txt")
	err := os.WriteFile(outsideFile, []byte("initial content"), 0644)
	if err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	err = os.Mkdir(workspaceDir, 0755)
	if err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	// Create a symlink inside workspace pointing to outside
	symlinkPath := filepath.Join(workspaceDir, "link_to_outside.txt")
	err = os.Symlink(outsideFile, symlinkPath)
	if err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	tool, err := NewWriteFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	ctx := context.Background()
	// Attempt to write via symlink
	argsJSON := fmt.Sprintf(`{"path": "link_to_outside.txt", "content": "OVERWRITTEN"}`)

	_, err = tool.InvokableRun(ctx, argsJSON)

	if err == nil {
		// Check if the outside file was modified
		content, _ := os.ReadFile(outsideFile)
		if string(content) == "OVERWRITTEN" {
			t.Fatalf("SECURITY FAILURE: Was able to overwrite file outside workspace via symlink!")
		}
	}
}

func TestPathTraversal_WriteFile_BrokenSymlink(t *testing.T) {
	tmpDir := t.TempDir()

	// Target file does not exist initially
	outsideFile := filepath.Join(tmpDir, "outside_new.txt")

	workspaceDir := filepath.Join(tmpDir, "workspace")
	err := os.Mkdir(workspaceDir, 0755)
	if err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	// Create a BROKEN symlink inside workspace pointing to outside
	symlinkPath := filepath.Join(workspaceDir, "link_to_new.txt")
	err = os.Symlink(outsideFile, symlinkPath)
	if err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	tool, err := NewWriteFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	ctx := context.Background()
	// Attempt to write via broken symlink
	argsJSON := fmt.Sprintf(`{"path": "link_to_new.txt", "content": "CREATED_OUTSIDE"}`)

	// This should fail
	_, err = tool.InvokableRun(ctx, argsJSON)

	if err == nil {
		// Check if the outside file was created
		content, _ := os.ReadFile(outsideFile)
		if string(content) == "CREATED_OUTSIDE" {
			t.Fatalf("SECURITY FAILURE: Was able to create file outside workspace via broken symlink!")
		}
	}
}
