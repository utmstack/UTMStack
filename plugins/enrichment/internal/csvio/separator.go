package csvio

import "strings"

func DetectSeparator(firstLine string) rune {
	candidates := []rune{',', ';', '\t', '|'}
	bestCount := 0
	bestSep := ','
	for _, c := range candidates {
		count := strings.Count(firstLine, string(c))
		if count > bestCount {
			bestCount = count
			bestSep = c
		}
	}
	return bestSep
}

func ResolveSeparator(formValue string, content []byte) rune {
	switch formValue {
	case ",":
		return ','
	case ";":
		return ';'
	case "\\t", "tab", "\t":
		return '\t'
	case "|":
		return '|'
	}
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) != "" {
			return DetectSeparator(line)
		}
	}
	return ','
}
