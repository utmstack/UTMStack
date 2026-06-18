package models

import (
	"encoding/json"
	"regexp"
	"strings"
)

func NewContainerFilter(rules []FilterRule) *ContainerFilter {
	return &ContainerFilter{rules: rules}
}

func (f *ContainerFilter) ShouldCollect(container Container) bool {
	defaultExcludes := []string{
		"k8s.gcr.io",
		"registry.k8s.io",
		"quay.io/coreos",
		"gcr.io/google-containers",
		"docker.io/library/pause",
	}

	for _, exclude := range defaultExcludes {
		if strings.HasPrefix(container.Image, exclude) {
			return false
		}
	}

	if container.Labels != nil {
		if exclude, exists := container.Labels["logs.exclude"]; exists && exclude == "true" {
			return false
		}
	}

	for _, rule := range f.rules {
		if f.matchesRule(container, rule) {
			return rule.Action == "include"
		}
	}

	return container.State == "running" ||
		strings.Contains(container.Status, "Up") ||
		(container.State == "exited" && strings.Contains(container.Status, "Exited (0)")) // Contenedores que terminaron exitosamente
}

func (f *ContainerFilter) matchesRule(container Container, rule FilterRule) bool {
	switch rule.Type {
	case "image":
		if rule.Pattern == "*" {
			return true
		}
		if matched, _ := regexp.MatchString(rule.Pattern, container.Image); matched {
			return true
		}
	case "name":
		if matched, _ := regexp.MatchString(rule.Pattern, container.Name); matched {
			return true
		}
	case "label":
		parts := strings.SplitN(rule.Pattern, "=", 2)
		if len(parts) == 2 {
			key, value := parts[0], parts[1]
			if container.Labels[key] == value {
				return true
			}
		}
	}
	return false
}

func ParseLogLine(line string, maxLength int) (string, error) {
	cleaned := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(line, "")

	if len(cleaned) > maxLength {
		cleaned = cleaned[:maxLength]
	}

	if strings.TrimSpace(cleaned) == "" {
		return "", ErrEmptyLogLine
	}

	return cleaned, nil
}

func EnrichLogWithContainer(logLine, containerName string) string {
	if !isValidJSON(logLine) {
		return logLine
	}

	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(logLine), &logData); err != nil {
		return logLine
	}

	cleanName := CleanContainerName(containerName)
	logData["containerName"] = cleanName

	enrichedBytes, err := json.Marshal(logData)
	if err != nil {
		return logLine
	}

	return string(enrichedBytes)
}

func isValidJSON(str string) bool {
	str = strings.TrimSpace(str)
	return (strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) ||
		(strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]"))
}

func CleanContainerName(containerName string) string {
	swarmPattern := `\.\d+\.[a-zA-Z0-9]+$`
	re := regexp.MustCompile(swarmPattern)
	cleaned := re.ReplaceAllString(containerName, "")

	composePattern := `[_-]\d+$`
	re2 := regexp.MustCompile(composePattern)
	cleaned = re2.ReplaceAllString(cleaned, "")

	return cleaned
}

var (
	ErrEmptyLogLine = NewError("empty log line after cleaning")
)

type Error struct {
	Message string
}

func NewError(msg string) *Error {
	return &Error{Message: msg}
}

func (e *Error) Error() string {
	return e.Message
}

func SafeContainerID(id string) string {
	if len(id) >= 12 {
		return id[:12]
	}
	return id
}
