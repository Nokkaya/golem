package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "runtime"
    "strings"
    "testing"
)

func TestExecTool_UsesWorkspaceDirWhenWorkingDirEmpty(t *testing.T) {
    tmpDir := t.TempDir()
    tool, err := NewExecTool(60, false, tmpDir)
    if err != nil {
        t.Fatalf("NewExecTool error: %v", err)
    }

    cmd := "pwd"
    if runtime.GOOS == "windows" {
        cmd = "cd"
    }

    ctx := context.Background()
    argsJSON := fmt.Sprintf(`{"command": %q}`, cmd)

    result, err := tool.InvokableRun(ctx, argsJSON)
    if err != nil {
        t.Fatalf("InvokableRun error: %v", err)
    }

    stdout := result
    var out ExecOutput
    if err := json.Unmarshal([]byte(result), &out); err == nil {
        stdout = out.Stdout
    }

    if !strings.Contains(stdout, tmpDir) {
        if runtime.GOOS == "windows" {
            escaped := strings.ReplaceAll(tmpDir, "\\", "\\\\")
            if strings.Contains(stdout, escaped) {
                return
            }
        }
        t.Fatalf("expected command to run in workspace dir %q, got output: %s", tmpDir, stdout)
    }
}
