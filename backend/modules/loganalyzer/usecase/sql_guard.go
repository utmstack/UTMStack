package usecase

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reStartsWithSelect = regexp.MustCompile(`(?is)^\s*(with|select)\b`)
	reLineComment      = regexp.MustCompile(`--`)
	reBlockComment     = regexp.MustCompile(`/\*`)
)

var forbidden = []string{
	"utmstack.", "system.", "information_schema", "default.",
	"remote", "remotesecure", "cluster", "clusterallreplicas", "statistics",
	"url", "file", "s3", "hdfs", "mysql", "postgresql", "mongodb", "jdbc", "odbc",
	"merge", "executable", "input", "dictionary",
	"insert", "alter", "create", "drop", "attach", "detach", "truncate",
	"rename", "optimize", "grant", "revoke", "set ", "system ", "kill",
}

func ValidateSQL(sql string) error {
	q := strings.TrimSpace(sql)
	if q == "" {
		return fmt.Errorf("query is required")
	}
	if !reStartsWithSelect.MatchString(q) {
		return fmt.Errorf("the query must start with SELECT or WITH")
	}
	// A comment can hide anything from the checks below.
	if reLineComment.MatchString(q) || reBlockComment.MatchString(q) {
		return fmt.Errorf("comments are not allowed")
	}

	stripped := strings.TrimSpace(strings.TrimRight(q, ";"))
	if strings.Contains(stripped, ";") {
		return fmt.Errorf("only one statement at a time")
	}

	lower := strings.ToLower(stripped)
	// Quoting can hide a forbidden token (e.g. `'utmstack'` or `` `utmstack` ``);
	// match against the query with quotes stripped so quoted names are caught too.
	matchable := strings.ReplaceAll(lower, "'", "")
	matchable = strings.ReplaceAll(matchable, "`", "")
	for _, tok := range forbidden {
		if strings.Contains(matchable, tok) {
			return fmt.Errorf("%q is not allowed here: the query may read the logs and alerts datasets and nothing else", strings.TrimSpace(tok))
		}
	}
	return nil
}
