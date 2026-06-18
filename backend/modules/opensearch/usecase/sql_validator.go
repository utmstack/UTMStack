package usecase

import (
	"fmt"
	"regexp"
	"strings"
)

func ValidateSQL(sql string) error {
	trimmed := strings.TrimSpace(sql)

	if trimmed == "" {
		return nil
	}

	if !reMustStartWithSelect.MatchString(trimmed) {
		return fmt.Errorf("SQL must start with SELECT")
	}

	for _, tok := range forbiddenTokens {
		if reForbiddenToken(tok).MatchString(trimmed) {
			return fmt.Errorf("forbidden SQL keyword: %s", tok)
		}
	}

	if reLineComment.MatchString(trimmed) || reBlockComment.MatchString(trimmed) {
		return fmt.Errorf("SQL comments are not allowed")
	}

	stripped := strings.TrimRight(trimmed, ";")
	stripped = strings.TrimSpace(stripped)
	if strings.Contains(stripped, ";") {
		return fmt.Errorf("SQL must not contain internal semicolons")
	}

	if !quotesBalanced(stripped, '\'') {
		return fmt.Errorf("unbalanced single quotes in SQL")
	}

	if !quotesBalanced(stripped, '"') {
		return fmt.Errorf("unbalanced double quotes in SQL")
	}

	if !parensBalanced(stripped) {
		return fmt.Errorf("unbalanced parentheses in SQL")
	}

	if reBadFrom.MatchString(stripped) {
		return fmt.Errorf("FROM clause must not be empty")
	}

	if reSelectComma.MatchString(stripped) ||
		reDoubleComma.MatchString(stripped) ||
		reCommaFrom.MatchString(stripped) {
		return fmt.Errorf("misplaced comma in SQL")
	}

	if reBareSubquery.MatchString(stripped) {
		return fmt.Errorf("subquery in FROM must have an alias")
	}

	if bad := findDisallowedFunction(stripped); bad != "" {
		return fmt.Errorf("disallowed function call: %s", bad)
	}

	if reHaving.MatchString(stripped) && !reGroupBy.MatchString(stripped) {
		return fmt.Errorf("HAVING clause requires GROUP BY")
	}

	return nil
}

// ---------------------------------------------------------------------------
// Precompiled regexps
// ---------------------------------------------------------------------------

var (
	reMustStartWithSelect = regexp.MustCompile(`(?is)^\s*select\b`)
	reLineComment         = regexp.MustCompile(`--`)
	reBlockComment        = regexp.MustCompile(`/\*`)
	reBadFrom             = regexp.MustCompile(`(?i)\bFROM\s*(?:GROUP|WHERE|ORDER|$)`)
	reSelectComma         = regexp.MustCompile(`(?i)SELECT\s*,`)
	reDoubleComma         = regexp.MustCompile(`,,`)
	reCommaFrom           = regexp.MustCompile(`(?i),\s*FROM`)
	reBareSubquery        = regexp.MustCompile(`(?i)\bFROM\s*\([^)]*\)\s*(?:WHERE|GROUP|ORDER|LIMIT|HAVING|$)`)
	reHaving              = regexp.MustCompile(`(?i)\bHAVING\b`)
	reGroupBy             = regexp.MustCompile(`(?i)\bGROUP\s+BY\b`)
)

var forbiddenTokens = []string{
	"insert", "update", "delete", "drop", "alter", "create",
	"replace", "truncate", "merge", "grant", "revoke",
	"exec", "execute", "commit", "rollback", "into",
}

func reForbiddenToken(tok string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b` + tok + `\b`)
}

var allowedFunctions = map[string]struct{}{
	"count": {},
	"avg":   {},
	"min":   {},
	"max":   {},
	"sum":   {},
}

func findDisallowedFunction(sql string) string {
	re := regexp.MustCompile(`(?i)\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	for _, match := range re.FindAllStringSubmatch(sql, -1) {
		name := strings.ToLower(match[1])
		if _, ok := allowedFunctions[name]; !ok {
			return match[1]
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func quotesBalanced(s string, q rune) bool {
	count := 0
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' {
			i++ // skip escaped character
			continue
		}
		if runes[i] == q {
			count++
		}
	}
	return count%2 == 0
}

func parensBalanced(s string) bool {
	depth := 0
	for _, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth < 0 {
				return false
			}
		}
	}
	return depth == 0
}
