## 2025-02-04 - Path Traversal in Filesystem Tools
**Vulnerability:** The `read_file`, `write_file`, and `list_dir` tools accepted absolute paths without any validation against a workspace root, allowing arbitrary file system access.
**Learning:** Tools in an agentic system must be sandboxed by default. Relying on "safe mode" flags or user trust is insufficient for file operations.
**Prevention:** Always wrap file operations in a path validation helper that resolves paths relative to a strict workspace root using `filepath.Rel` and checking for `..` components.

## 2025-02-18 - Shell Execution Path Traversal
**Vulnerability:** The `exec` tool honored `WorkingDir` parameter even when `restrictToWorkspace` was enabled, allowing execution of commands in arbitrary directories.
**Learning:** Configuration flags like `restrictToWorkspace` must be enforced on all input parameters that affect file system access, not just the command itself.
**Prevention:** Validate `WorkingDir` against the workspace root using the same path validation logic as filesystem tools.

## 2025-05-23 - Path Traversal via Symbolic Links
**Vulnerability:** Filesystem tools (`read_file`, `write_file`, etc.) were vulnerable to path traversal because they only checked for `..` in the relative path without resolving symbolic links. An attacker could create a symlink inside the workspace pointing to a sensitive file outside (e.g., `/etc/passwd`) and read it.
**Learning:** Checking for `..` is insufficient for path validation in environments where symlinks are possible. `filepath.Clean` and `filepath.Abs` do not resolve symlinks.
**Prevention:** Always use `filepath.EvalSymlinks` to resolve the canonical path of both the workspace root and the target file before verifying that the target is within the workspace. Ensure to handle cases where the target file does not exist (e.g., for write operations) by validating the parent directory.
