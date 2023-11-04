package correlation

import (
	"bytes"
	"html/template"
	"log"
	"time"

	"github.com/utmstack/UTMStack/correlation/cache"
	"github.com/utmstack/UTMStack/correlation/rules"
	"github.com/utmstack/UTMStack/correlation/search"
)

func Finder(rule rules.Rule) {
	sleep := rule.Frequency * time.Second
	for {
		start := time.Now()
		if len(rule.Cache) != 0 {
			findInCache(rule)
		} else if len(rule.Search) != 0 {
			findInSearch(rule)
		}
		log.Printf("Process rule '%s' took: %s", rule.Name, time.Since(start))
		
		switch sleep{
		case 0:
			time.Sleep(5 * time.Minute)
		default:
			time.Sleep(sleep)
		}
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
					log.Printf("Error while trying to process the query %v of the rule %s: %v", step+1, rule.Name, err)
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
						log.Printf("Error while trying to process the query %v of the rule %s: %v", step+1, rule.Name, err)
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
						log.Printf("Error while trying to process the query %v of the rule %s: %v", step+1, rule.Name, err)
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
