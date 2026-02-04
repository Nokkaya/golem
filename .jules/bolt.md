## 2026-02-04 - Replicate strings.Split behavior with streaming
**Learning:** When optimizing file reading by replacing `os.ReadFile` + `strings.Split` with `bufio.Reader`, replicating `strings.Split` behavior for edge cases (empty file -> `[""]`, trailing newline -> extra empty line) is non-trivial and requires explicit checks.
**Action:** When implementing streaming file readers that must match legacy split behavior, always verify "empty file" and "trailing newline" cases with dedicated tests.
