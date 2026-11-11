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

// TestWriteFile_DanglingSymlinkOutside proves that a broken symlink inside the
// workspace, whose target does not exist and lies outside the workspace, cannot
// be used to create that outside file. os.Stat follows the link and reports
// "not exist", so a parent-only containment check sees the link path itself
// (which is inside the workspace) and os.WriteFile then follows the link.
func TestWriteFile_DanglingSymlinkOutside(t *testing.T) {
	for _, useAbs := range []bool{false, true} {
		name := "workspace-relative"
		if useAbs {
			name = "absolute"
		}
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()
			outsideFile := filepath.Join(tmpDir, "outside_new.txt")

			workspaceDir := filepath.Join(tmpDir, "workspace")
			if err := os.Mkdir(workspaceDir, 0755); err != nil {
				t.Fatalf("failed to create workspace dir: %v", err)
			}

			symlinkPath := filepath.Join(workspaceDir, "dangling")
			if err := os.Symlink(outsideFile, symlinkPath); err != nil {
				t.Fatalf("failed to create symlink: %v", err)
			}

			tool, err := NewWriteFileTool(workspaceDir)
			if err != nil {
				t.Fatalf("NewWriteFileTool error: %v", err)
			}

			path := "dangling"
			if useAbs {
				path = symlinkPath
			}
			argsJSON := fmt.Sprintf(`{"path": %q, "content": "CREATED_OUTSIDE"}`, path)
			_, err = tool.InvokableRun(context.Background(), argsJSON)
			if err == nil {
				t.Fatal("SECURITY VULNERABILITY: write through dangling symlink was allowed")
			}
			if !strings.Contains(err.Error(), "access denied") {
				t.Fatalf("expected access denied, got: %v", err)
			}
			if data, readErr := os.ReadFile(outsideFile); readErr == nil {
				t.Fatalf("SECURITY VULNERABILITY: created %s via dangling symlink with content %q", outsideFile, data)
			} else if !os.IsNotExist(readErr) {
				t.Fatalf("stat outside file: %v", readErr)
			}
		})
	}
}

// TestWriteFile_DanglingRelativeSymlinkOutside is the same bypass with a
// relative link target (../outside), which WriteFile resolves from the
// symlink's directory.
func TestWriteFile_DanglingRelativeSymlinkOutside(t *testing.T) {
	tmpDir := t.TempDir()
	outsideFile := filepath.Join(tmpDir, "outside_rel.txt")

	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.Mkdir(workspaceDir, 0755); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	symlinkPath := filepath.Join(workspaceDir, "dangling_rel")
	if err := os.Symlink("../outside_rel.txt", symlinkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	tool, err := NewWriteFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	_, err = tool.InvokableRun(context.Background(), `{"path": "dangling_rel", "content": "CREATED_OUTSIDE"}`)
	if err == nil {
		t.Fatal("SECURITY VULNERABILITY: write through relative dangling symlink was allowed")
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Fatalf("expected access denied, got: %v", err)
	}
	if data, readErr := os.ReadFile(outsideFile); readErr == nil {
		t.Fatalf("SECURITY VULNERABILITY: created %s via relative dangling symlink with content %q", outsideFile, data)
	} else if !os.IsNotExist(readErr) {
		t.Fatalf("stat outside file: %v", readErr)
	}
}

// TestWriteFile_DanglingSymlinkChainOutside rejects a link that points at
// another dangling link whose final target is outside the workspace.
func TestWriteFile_DanglingSymlinkChainOutside(t *testing.T) {
	tmpDir := t.TempDir()
	outsideFile := filepath.Join(tmpDir, "outside_chain.txt")

	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.Mkdir(workspaceDir, 0755); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	inner := filepath.Join(workspaceDir, "inner")
	if err := os.Symlink(outsideFile, inner); err != nil {
		t.Fatalf("failed to create inner symlink: %v", err)
	}
	outer := filepath.Join(workspaceDir, "outer")
	if err := os.Symlink(inner, outer); err != nil {
		t.Fatalf("failed to create outer symlink: %v", err)
	}

	tool, err := NewWriteFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	_, err = tool.InvokableRun(context.Background(), `{"path": "outer", "content": "CREATED_OUTSIDE"}`)
	if err == nil {
		t.Fatal("SECURITY VULNERABILITY: write through dangling symlink chain was allowed")
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Fatalf("expected access denied, got: %v", err)
	}
	if data, readErr := os.ReadFile(outsideFile); readErr == nil {
		t.Fatalf("SECURITY VULNERABILITY: created %s via dangling symlink chain with content %q", outsideFile, data)
	} else if !os.IsNotExist(readErr) {
		t.Fatalf("stat outside file: %v", readErr)
	}
}

// TestWriteFile_ParentSymlinkOutside covers a not-yet-existing file whose
// parent directory is a symlink to a directory outside the workspace.
// PR #32 already canonicalizes that parent; this must keep failing closed.
func TestWriteFile_ParentSymlinkOutside(t *testing.T) {
	tmpDir := t.TempDir()
	outsideDir := filepath.Join(tmpDir, "outside_dir")
	if err := os.Mkdir(outsideDir, 0755); err != nil {
		t.Fatalf("failed to create outside dir: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.Mkdir(workspaceDir, 0755); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}
	if err := os.Symlink(outsideDir, filepath.Join(workspaceDir, "escape")); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	tool, err := NewWriteFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	_, err = tool.InvokableRun(context.Background(), `{"path": "escape/new.txt", "content": "PWNED"}`)
	if err == nil {
		t.Fatal("SECURITY VULNERABILITY: write through parent symlink was allowed")
	}
	if !strings.Contains(err.Error(), "access denied") && !strings.Contains(err.Error(), "outside workspace") {
		t.Fatalf("expected access denied, got: %v", err)
	}
	outsideFile := filepath.Join(outsideDir, "new.txt")
	if data, readErr := os.ReadFile(outsideFile); readErr == nil {
		t.Fatalf("SECURITY VULNERABILITY: created %s via parent symlink with content %q", outsideFile, data)
	} else if !os.IsNotExist(readErr) {
		t.Fatalf("stat outside file: %v", readErr)
	}
}

// TestWriteFile_ParentSymlinkNestedOutside is a not-yet-existing file nested
// under a real directory that is only reachable through a parent symlink
// pointing outside the workspace.
func TestWriteFile_ParentSymlinkNestedOutside(t *testing.T) {
	tmpDir := t.TempDir()
	outsideDir := filepath.Join(tmpDir, "outside_dir")
	nestedDir := filepath.Join(outsideDir, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create outside dir: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.Mkdir(workspaceDir, 0755); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}
	if err := os.Symlink(outsideDir, filepath.Join(workspaceDir, "escape")); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	tool, err := NewWriteFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	_, err = tool.InvokableRun(context.Background(), `{"path": "escape/nested/new.txt", "content": "PWNED"}`)
	if err == nil {
		t.Fatal("SECURITY VULNERABILITY: write through nested parent symlink was allowed")
	}
	outsideFile := filepath.Join(nestedDir, "new.txt")
	if data, readErr := os.ReadFile(outsideFile); readErr == nil {
		t.Fatalf("SECURITY VULNERABILITY: created %s via nested parent symlink with content %q", outsideFile, data)
	} else if !os.IsNotExist(readErr) {
		t.Fatalf("stat outside file: %v", readErr)
	}
}

// TestWriteFile_DanglingSymlinkInsideRejected documents that every broken
// symlink is denied, including one whose lexical target is inside the
// workspace. The destination is not canonicalized by EvalSymlinks, and
// write_file would follow the link.
func TestWriteFile_DanglingSymlinkInsideRejected(t *testing.T) {
	workspaceDir := t.TempDir()
	if err := os.Symlink("missing.txt", filepath.Join(workspaceDir, "dangling")); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	tool, err := NewWriteFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	_, err = tool.InvokableRun(context.Background(), `{"path": "dangling", "content": "nope"}`)
	if err == nil {
		t.Fatal("expected broken symlink to be rejected")
	}
	if !strings.Contains(err.Error(), "broken symlink") {
		t.Fatalf("expected broken symlink error, got: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(workspaceDir, "missing.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("write through in-workspace dangling symlink created the target: %v", statErr)
	}
}

// TestWriteFile_ExistingSymlinkOutside ensures a symlink to an existing file
// outside the workspace cannot overwrite that file (PR #32 behavior).
func TestWriteFile_ExistingSymlinkOutside(t *testing.T) {
	tmpDir := t.TempDir()
	outsideFile := filepath.Join(tmpDir, "outside_existing.txt")
	if err := os.WriteFile(outsideFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}

	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.Mkdir(workspaceDir, 0755); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}
	if err := os.Symlink(outsideFile, filepath.Join(workspaceDir, "link_to_secret")); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	tool, err := NewWriteFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	_, err = tool.InvokableRun(context.Background(), `{"path": "link_to_secret", "content": "OVERWRITTEN"}`)
	if err == nil {
		t.Fatal("SECURITY VULNERABILITY: overwrite through symlink was allowed")
	}
	got, readErr := os.ReadFile(outsideFile)
	if readErr != nil {
		t.Fatalf("read outside file: %v", readErr)
	}
	if string(got) != "initial" {
		t.Fatalf("outside file was modified: %q", got)
	}
}

// TestWriteFile_NewFileInsideWorkspace is the non-symlink create path: a
// missing file whose parent is the workspace itself must still succeed.
func TestWriteFile_NewFileInsideWorkspace(t *testing.T) {
	workspaceDir := t.TempDir()
	tool, err := NewWriteFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}

	_, err = tool.InvokableRun(context.Background(), `{"path": "notes.txt", "content": "hello"}`)
	if err != nil {
		t.Fatalf("write inside workspace failed: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(workspaceDir, "notes.txt"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("content = %q, want hello", got)
	}
}

// TestReadWrite_SymlinkInsideWorkspace keeps legitimate symlinks that resolve
// inside the workspace working for both read and write.
func TestReadWrite_SymlinkInsideWorkspace(t *testing.T) {
	workspaceDir := t.TempDir()
	realPath := filepath.Join(workspaceDir, "real.txt")
	if err := os.WriteFile(realPath, []byte("inside"), 0644); err != nil {
		t.Fatalf("failed to create real file: %v", err)
	}
	if err := os.Symlink("real.txt", filepath.Join(workspaceDir, "alias.txt")); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	ctx := context.Background()
	readTool, err := NewReadFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewReadFileTool error: %v", err)
	}
	result, err := readTool.InvokableRun(ctx, `{"path": "alias.txt"}`)
	if err != nil {
		t.Fatalf("read via in-workspace symlink failed: %v", err)
	}
	if !strings.Contains(result, "inside") {
		t.Fatalf("read via in-workspace symlink = %s, want contents of real.txt", result)
	}

	writeTool, err := NewWriteFileTool(workspaceDir)
	if err != nil {
		t.Fatalf("NewWriteFileTool error: %v", err)
	}
	if _, err := writeTool.InvokableRun(ctx, `{"path": "alias.txt", "content": "updated"}`); err != nil {
		t.Fatalf("write via in-workspace symlink failed: %v", err)
	}
	got, err := os.ReadFile(realPath)
	if err != nil {
		t.Fatalf("read real file: %v", err)
	}
	if string(got) != "updated" {
		t.Fatalf("real file = %q, want updated", got)
	}
}
