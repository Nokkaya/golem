package tools

import (
    "context"
    "os"
    "strings"

    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/components/tool/utils"
)

// ReadFileInput parameters for read_file tool
type ReadFileInput struct {
    Path   string `json:"path" jsonschema:"required,description=Absolute path to the file"`
    Offset int    `json:"offset" jsonschema:"description=Starting line number (0-based)"`
    Limit  int    `json:"limit" jsonschema:"description=Maximum number of lines to read"`
}

// ReadFileOutput result of read_file tool
type ReadFileOutput struct {
    Content    string `json:"content"`
    TotalLines int    `json:"total_lines"`
}

func readFile(ctx context.Context, input *ReadFileInput) (*ReadFileOutput, error) {
    data, err := os.ReadFile(input.Path)
    if err != nil {
        return nil, err
    }

    content := string(data)
    lines := strings.Split(content, "\n")
    totalLines := len(lines)

    if input.Offset > 0 {
        if input.Offset >= len(lines) {
            lines = []string{}
        } else {
            lines = lines[input.Offset:]
        }
    }

    if input.Limit > 0 && input.Limit < len(lines) {
        lines = lines[:input.Limit]
    }

    return &ReadFileOutput{
        Content:    strings.Join(lines, "\n"),
        TotalLines: totalLines,
    }, nil
}

// NewReadFileTool creates the read_file tool
func NewReadFileTool() (tool.InvokableTool, error) {
    return utils.InferTool("read_file", "Read the contents of a file", readFile)
}

// WriteFileInput parameters for write_file tool
type WriteFileInput struct {
    Path    string `json:"path" jsonschema:"required,description=Absolute path to the file"`
    Content string `json:"content" jsonschema:"required,description=Content to write"`
}

func writeFile(ctx context.Context, input *WriteFileInput) (string, error) {
    err := os.WriteFile(input.Path, []byte(input.Content), 0644)
    if err != nil {
        return "", err
    }
    return "File written successfully", nil
}

// NewWriteFileTool creates the write_file tool
func NewWriteFileTool() (tool.InvokableTool, error) {
    return utils.InferTool("write_file", "Write content to a file", writeFile)
}

// ListDirInput parameters for list_dir tool
type ListDirInput struct {
    Path string `json:"path" jsonschema:"required,description=Directory path to list"`
}

func listDir(ctx context.Context, input *ListDirInput) ([]string, error) {
    entries, err := os.ReadDir(input.Path)
    if err != nil {
        return nil, err
    }

    var result []string
    for _, entry := range entries {
        name := entry.Name()
        if entry.IsDir() {
            name += "/"
        }
        result = append(result, name)
    }
    return result, nil
}

// NewListDirTool creates the list_dir tool
func NewListDirTool() (tool.InvokableTool, error) {
    return utils.InferTool("list_dir", "List contents of a directory", listDir)
}
