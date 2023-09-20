package correlation

import (
	"bytes"
	"html/template"
	"time"

	"github.com/AtlasInsideCorp/UTMStackCorrelationEngine/cache"
	"github.com/AtlasInsideCorp/UTMStackCorrelationEngine/rules"
	"github.com/AtlasInsideCorp/UTMStackCorrelationEngine/search"
	"github.com/AtlasInsideCorp/UTMStackCorrelationEngine/utils"
	"github.com/quantfall/holmes"
)

var h = holmes.New(utils.GetConfig().ErrorLevel, "CORRELATION")

func Finder(rule rules.Rule) {
	for {
		start := time.Now()
		if len(rule.Cache) != 0 {
			findInCache(rule)
		} else if len(rule.Search) != 0 {
			findInSearch(rule)
		}
		h.Debug("Process rule '%s' took: %s", rule.Name, time.Since(start))
		time.Sleep(rule.Frequency * time.Second)
	}
}

func findInSearch(rule rules.Rule) {
	var tmpLogs [20][]map[string]string

	for step, query := range rule.Search {
		if step == 0 {
			l := search.Search(query.Query)
			processResponse(l, rule, query.Save, &tmpLogs, len(rule.Search), step, query.MinCount)
		} else {
			for _, fields := range tmpLogs[step-1] {
				var q bytes.Buffer
				t := template.Must(template.New("query").Parse(query.Query))
				err := t.Execute(&q, fields)
				if err != nil {
					h.Error("Error while trying to process the query %v of the rule %s: %v", step+1, rule.Name, err)
				} else {
					l := search.Search(q.String())
					processResponse(l, rule, query.Save, &tmpLogs, len(rule.Search), step, query.MinCount)
				}
			}
		}
	}
}

func findInCache(rule rules.Rule) {
	var tmpLogs [20][]map[string]string
	for step, query := range rule.Cache {
		if step == 0 {
			l := cache.Search(query.AllOf, query.OneOf, query.TimeLapse)
			processResponse(l, rule, query.Save, &tmpLogs, len(rule.Cache), step, query.MinCount)
		} else {
			for _, fields := range tmpLogs[step-1] {
				var allOfList []rules.AllOf
				for _, allOf := range query.AllOf {
					var value bytes.Buffer
					t := template.Must(template.New("allOf").Parse(allOf.Value))
					err := t.Execute(&value, fields)
					if err != nil {
						h.Error("Error while trying to process the query %v of the rule %s: %v", step+1, rule.Name, err)
					} else {
						allOfList = append(allOfList, rules.AllOf{Field: allOf.Field, Operator: allOf.Operator, Value: value.String()})
					}
				}

				var oneOfList []rules.OneOf
				for _, oneOf := range query.OneOf {
					var value bytes.Buffer
					t := template.Must(template.New("oneOf").Parse(oneOf.Value))
					err := t.Execute(&value, fields)
					if err != nil {
						h.Error("Error while trying to process the query %v of the rule %s: %v", step+1, rule.Name, err)
					} else {
						oneOfList = append(oneOfList, rules.OneOf{Field: oneOf.Field, Operator: oneOf.Operator, Value: value.String()})
					}
				}

				l := cache.Search(allOfList, oneOfList, query.TimeLapse)
				processResponse(l, rule, query.Save, &tmpLogs, len(rule.Cache), step, query.MinCount)
			}
		}
	}
}
