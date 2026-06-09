package mcp

const (
	defaultPageSize = 50
	maxPageSize     = 500
)

// clampPageSize normalises a caller-supplied page size: empty/zero falls back
// to the default; values above maxPageSize are capped so a buggy LLM cannot
// accidentally pull thousands of rows in a single tool call.
func clampPageSize(size int) int {
	if size <= 0 {
		return defaultPageSize
	}
	if size > maxPageSize {
		return maxPageSize
	}
	return size
}
