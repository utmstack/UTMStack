package config

import (
	"regexp"
	"sync"
)

// SensitivePattern defines a pattern for anonymizing sensitive data
type SensitivePattern struct {
	Regexp    string
	FakeValue string
	compiled  *regexp.Regexp
	once      sync.Once
}

// GetRegexp returns the compiled regexp, compiling it on first access
func (sp *SensitivePattern) GetRegexp() *regexp.Regexp {
	sp.once.Do(func() {
		sp.compiled = regexp.MustCompile(sp.Regexp)
	})
	return sp.compiled
}

// Fake values for anonymization
var (
	FakeUserName = "John Doe"
	FakeEmail    = "jhondoe@gmail.com"
)

// SensitivePatterns defines patterns to anonymize in alert data
var SensitivePatterns = map[string]*SensitivePattern{
	"email": {Regexp: `([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`, FakeValue: "jhondoe@gmail.com"},
	//"ipv4":  {Regexp: `(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`, FakeValue: "10.0.0.1"},
}

// init pre-compiles all sensitive patterns at startup to avoid
// latency during first alert processing
func init() {
	for _, pattern := range SensitivePatterns {
		_ = pattern.GetRegexp() // Trigger compilation via sync.Once
	}
}
