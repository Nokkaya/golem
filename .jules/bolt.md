## 2025-02-18 - [Optimizing Tool File Access]
**Learning:** Tools in this agent architecture often need to read small slices of potentially large files (e.g., logs). Using `os.ReadFile` followed by `strings.Split` forces the entire file into memory twice (bytes + string slice).
**Action:** Prefer streaming file access (checking `\n` manually) and `ReadAt` to extract only the needed segment. This reduces memory pressure from O(File) to O(Buffer). Be careful to replicate `strings.Split` edge cases (e.g., trailing newlines).

## 2025-02-18 - [Optimizing Loop Performance]
**Learning:** Scalar loops iterating byte-by-byte (`for i, b := range chunk`) are significantly slower than optimized standard library functions (`bytes.IndexByte`, `bytes.Count`) which utilize assembly/SIMD instructions.
**Action:** Replace manual byte scanning loops with `bytes` package functions whenever possible for high-throughput I/O operations.
