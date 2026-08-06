package usecase

import (
	"context"

	"gorm.io/gorm"
)

// PipelineBootstrap migrates utm_tenant_config and utm_regex_pattern from the
// the pipeline YAML the event processor reads.
type PipelineBootstrap struct {
	writer *pipelineWriter
	db     *gorm.DB
}

func NewPipelineBootstrap(writer *pipelineWriter, db *gorm.DB) *PipelineBootstrap {
	return &PipelineBootstrap{writer: writer, db: db}
}

// Run is idempotent and safe to call on every boot.
// Run writes the named patterns the filter engine resolves {{.patternName}}
// against. They are baked in rather than stored, so this is a plain write.
func (b *PipelineBootstrap) Run(ctx context.Context) error {
	patterns := make(map[string]string, len(systemPatterns))
	for k, v := range systemPatterns {
		patterns[k] = v
	}
	return b.writer.WritePatterns(patterns)
}

// systemPatterns are the 26 canonical named patterns the filter engine uses as
// {{.patternName}} references. Previously seeded into utm_regex_pattern; now
// baked into the bootstrap so the patterns.yaml always contains them even on a
// fresh install (no DB table needed).
var systemPatterns = map[string]string{
	"ciscoMacAddr":    `(?:(?:[A-Fa-f0-9]{4}\.){2}[A-Fa-f0-9]{4})`,
	"syslogDate":      `[A-Z][a-z]{2} \d{1,2} \d{2}:\d{2}:\d{2}`,
	"winMacAddr":      `(?:(?:[A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2})`,
	"commonMacAddr":   `(?:(?:[A-Fa-f0-9]{2}:){5}[A-Fa-f0-9]{2})|(?:(?:[A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2})`,
	"integer":         `(?:[+-]?(?:[0-9]+))`,
	"day":             `(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)`,
	"word":            `\b\w+\b`,
	"greedy":          `.*`,
	"space":           `\s+`,
	"notSpace":        `\S+`,
	"monthName":       `\b(?:[Jj]an(?:uary|uar)?|[Ff]eb(?:ruary|ruar)?|[Mm](?:a|ä)?r(?:ch|z)?|[Aa]pr(?:il)?|[Mm]a(?:y|i)?|[Jj]un(?:e|i)?|[Jj]ul(?:y|i)?|[Aa]ug(?:ust)?|[Ss]ep(?:tember)?|[Oo](?:c|k)?t(?:ober)?|[Nn]ov(?:ember)?|[Dd]e(?:c|z)(?:ember)?)\b`,
	"ipv4":            `(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.)){3}((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))`,
	"email":           `((?P<name>[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+)@(?P<domain>[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*))`,
	"domain":          `((?:[_a-z0-9](?:[_a-z0-9-]{0,61}[a-z0-9])?\.)+(?:[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?)?)`,
	"hostname":        `(\b(?:[0-9A-Za-z][0-9A-Za-z-]{0,62})(?:\.(?:[0-9A-Za-z][0-9A-Za-z-]{0,62}))*(\.?|\b))`,
	"data":            `(.*?)`,
	"ipv6":            `([0-9a-fA-F]{1,4}(:[0-9a-fA-F]{0,4}){1,7}|::[0-1]?)`,
	"uuid":            `([A-Fa-f0-9]{8}-(?:[A-Fa-f0-9]{4}-){3}[A-Fa-f0-9]{12})`,
	"monthNumber":     `(?:0[1-9]|1[0-2])`,
	"monthDay":        `(?:(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9])`,
	"year":            `(([1-9])[0-9]{1,3})`,
	"hour":            `(([01][0-9])|2[0-4])`,
	"minute":          `(?:[0-5][0-9])`,
	"seconds":         `(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)`,
	"time":            `((([01][0-9])|2[0-4]):(?:[0-5][0-9])(?::(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)))`,
	"iso8601Timezone": `(Z|([+-](([01][0-9])|2[0-4]):?([0-5][0-9])))`,
}
