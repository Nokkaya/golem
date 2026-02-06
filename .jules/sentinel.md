## 2025-02-04 - Path Traversal in Filesystem Tools
**Vulnerability:** The `read_file`, `write_file`, and `list_dir` tools accepted absolute paths without any validation against a workspace root, allowing arbitrary file system access.
**Learning:** Tools in an agentic system must be sandboxed by default. Relying on "safe mode" flags or user trust is insufficient for file operations.
**Prevention:** Always wrap file operations in a path validation helper that resolves paths relative to a strict workspace root using `filepath.Rel` and checking for `..` components.

## 2025-02-18 - Shell Execution Path Traversal
**Vulnerability:** The `exec` tool honored `WorkingDir` parameter even when `restrictToWorkspace` was enabled, allowing execution of commands in arbitrary directories.
**Learning:** Configuration flags like `restrictToWorkspace` must be enforced on all input parameters that affect file system access, not just the command itself.
**Prevention:** Validate `WorkingDir` against the workspace root using the same path validation logic as filesystem tools.

## 2025-05-23 - Symlink Path Traversal
**Vulnerability:** The `validatePath` function in filesystem tools did not resolve symbolic links, allowing access to files outside the workspace via a symlink created inside the workspace.
**Learning:** `filepath.Rel` only checks lexical path components. To prevent path traversal, one must always resolve symbolic links using `filepath.EvalSymlinks` to get the canonical path before validation.
**Prevention:** Always use `filepath.EvalSymlinks` on both the base directory and the target path, then check if the target is within the base.
