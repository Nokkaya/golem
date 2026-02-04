package tools

import (
    "context"
    "fmt"
    "os/exec"
    "runtime"
    "strings"
    "time"

    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/components/tool/utils"
)

// ExecInput parameters for exec tool
type ExecInput struct {
    Command    string `json:"command" jsonschema:"required,description=Shell command to execute"`
    WorkingDir string `json:"working_dir" jsonschema:"description=Working directory for the command"`
}

// ExecOutput result of exec tool
type ExecOutput struct {
    Stdout   string `json:"stdout"`
    Stderr   string `json:"stderr"`
    ExitCode int    `json:"exit_code"`
}

// Dangerous commands to block
var dangerousCommands = []string{
    "rm -rf /",
    "rm -rf ~",
    "mkfs",
    "dd if=",
    ":(){:|:&};:",
    "format c:",
    "del /f /s /q",
}

type execToolImpl struct {
    timeout             time.Duration
    restrictToWorkspace bool
    workspaceDir        string
}

func (e *execToolImpl) execute(ctx context.Context, input *ExecInput) (*ExecOutput, error) {
    cmdLower := strings.ToLower(input.Command)
    for _, dangerous := range dangerousCommands {
        if strings.Contains(cmdLower, dangerous) {
            return &ExecOutput{
                Stderr:   fmt.Sprintf("Blocked dangerous command: %s", dangerous),
                ExitCode: 1,
            }, nil
        }
    }

    var cmd *exec.Cmd
    if runtime.GOOS == "windows" {
        cmd = exec.CommandContext(ctx, "cmd", "/C", input.Command)
    } else {
        cmd = exec.CommandContext(ctx, "sh", "-c", input.Command)
    }

    if input.WorkingDir != "" {
        cmd.Dir = input.WorkingDir
    } else if e.workspaceDir != "" {
        cmd.Dir = e.workspaceDir
    }

    timeoutCtx, cancel := context.WithTimeout(ctx, e.timeout)
    defer cancel()
    cmd = exec.CommandContext(timeoutCtx, cmd.Path, cmd.Args[1:]...)
    cmd.Dir = input.WorkingDir

    var stdout, stderr strings.Builder
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    exitCode := 0
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            exitCode = exitErr.ExitCode()
        } else {
            return &ExecOutput{
                Stderr:   err.Error(),
                ExitCode: 1,
            }, nil
        }
    }

    return &ExecOutput{
        Stdout:   stdout.String(),
        Stderr:   stderr.String(),
        ExitCode: exitCode,
    }, nil
}

// NewExecTool creates the exec tool
func NewExecTool(timeoutSec int, restrictToWorkspace bool, workspaceDir string) (tool.InvokableTool, error) {
    impl := &execToolImpl{
        timeout:             time.Duration(timeoutSec) * time.Second,
        restrictToWorkspace: restrictToWorkspace,
        workspaceDir:        workspaceDir,
    }
    return utils.InferTool("exec", "Execute a shell command", impl.execute)
}
