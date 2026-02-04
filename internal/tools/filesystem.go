package tools

import (
	"bufio"
	"context"
	"io"
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

func readLine(reader *bufio.Reader, keep bool) (string, bool, bool, error) {
	if keep {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", false, false, err
		}
		emptyRead := (err == io.EOF && len(line) == 0)
		hasNewline := strings.HasSuffix(line, "\n")
		return line, hasNewline, emptyRead, err
	}

	totalBytes := 0
	for {
		frag, err := reader.ReadSlice('\n')
		totalBytes += len(frag)
		if err == nil {
			return "", true, false, nil
		}
		if err == io.EOF {
			emptyRead := (totalBytes == 0)
			return "", false, emptyRead, io.EOF
		}
		if err != bufio.ErrBufferFull {
			return "", false, false, err
		}
	}
}

func readFile(ctx context.Context, input *ReadFileInput) (*ReadFileOutput, error) {
	f, err := os.Open(input.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	totalLines := 0
	currentLineIdx := 0

	reader := bufio.NewReader(f)
	var lastLineEndedWithNewline bool

	for {
		keep := false
		if currentLineIdx >= input.Offset {
			if input.Limit == 0 || (currentLineIdx-input.Offset) < input.Limit {
				keep = true
			}
		}

		lineContent, hasNewline, emptyRead, err := readLine(reader, keep)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if emptyRead {
			// End of stream immediately.
			if currentLineIdx == 0 || lastLineEndedWithNewline {
				totalLines++
				// Check if we need to keep this empty line
				if currentLineIdx >= input.Offset {
					if input.Limit == 0 || (currentLineIdx-input.Offset) < input.Limit {
						lines = append(lines, "")
					}
				}
			}
			break
		}

		lastLineEndedWithNewline = hasNewline

		if keep {
			content := strings.TrimSuffix(lineContent, "\n")
			lines = append(lines, content)
		}

		totalLines++
		currentLineIdx++

		if err == io.EOF {
			break
		}
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
