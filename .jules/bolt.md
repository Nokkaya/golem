## 2025-02-18 - [Optimizing Tool File Access]
**Learning:** Tools in this agent architecture often need to read small slices of potentially large files (e.g., logs). Using `os.ReadFile` followed by `strings.Split` forces the entire file into memory twice (bytes + string slice).
**Action:** Prefer streaming file access (checking `\n` manually) and `ReadAt` to extract only the needed segment. This reduces memory pressure from O(File) to O(Buffer). Be careful to replicate `strings.Split` edge cases (e.g., trailing newlines).

## 2025-05-21 - [Efficient Byte Scanning in Go]
**Learning:** Manual byte-by-byte iteration in Go (`for i, b := range chunk`) to find delimiters is significantly slower than using SIMD-optimized `bytes` package functions.
**Action:** Replace manual loops with `bytes.IndexByte` or `bytes.Count` when scanning buffers. In `read_file`, this yielded a ~3.6x speedup (11.6ms -> 3.2ms for 10MB file).
