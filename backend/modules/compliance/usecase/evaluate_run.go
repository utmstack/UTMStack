package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

const checkConcurrency = 8

type checkSig struct {
	dataset  domain.Dataset
	dataType string
	rule     domain.CheckRule
	ruleVal  int
	filters  string
}

type dataKey struct {
	dataset  domain.Dataset
	dataType string
}

type checkOutcome struct {
	outcome domain.CheckOutcome
	hits    int64
	err     string
}

func sigOf(ch domain.Check) checkSig {
	rv := 0
	if ch.RuleValue != nil {
		rv = *ch.RuleValue
	}
	f, _ := json.Marshal(ch.Filters)
	return checkSig{
		dataset:  effectiveDataset(ch),
		dataType: ch.DataType,
		rule:     ch.Rule,
		ruleVal:  rv,
		filters:  string(f),
	}
}

func (e *evaluator) runChecks(ctx context.Context, checks map[checkSig]domain.Check, from, to time.Time) map[checkSig]checkOutcome {
	out := make(map[checkSig]checkOutcome, len(checks))

	present := map[dataKey]bool{}
	for sig := range checks {
		present[dataKey{sig.dataset, sig.dataType}] = false
	}
	for k := range present {
		ok, err := e.events.HasData(ctx, k.dataset, k.dataType, from, to)
		if err != nil {
			_ = catcher.Error("compliance: data presence probe failed", err,
				map[string]any{"dataset": string(k.dataset), "dataType": k.dataType})
			ok = true
		}
		present[k] = ok
	}

	runnable := make(map[checkSig]domain.Check, len(checks))
	for sig, ch := range checks {
		if present[dataKey{sig.dataset, sig.dataType}] {
			runnable[sig] = ch
			continue
		}
		out[sig] = checkOutcome{outcome: domain.CheckNotApplicable}
	}

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		sem = make(chan struct{}, checkConcurrency)
	)
	for sig, ch := range runnable {
		wg.Add(1)
		go func(sig checkSig, ch domain.Check) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			res := checkOutcome{}
			hits, err := e.events.Count(ctx, connectors.CheckQuery{
				Dataset:  effectiveDataset(ch),
				DataType: ch.DataType,
				Filters:  ch.Filters,
				From:     from,
				To:       to,
			})
			switch {
			case err != nil:
				res.outcome, res.err = domain.CheckError, err.Error()
			case passesRule(ch, hits):
				res.outcome, res.hits = domain.CheckPassed, hits
			default:
				res.outcome, res.hits = domain.CheckFailed, hits
			}

			mu.Lock()
			out[sig] = res
			mu.Unlock()
		}(sig, ch)
	}
	wg.Wait()
	return out
}

func passesRule(ch domain.Check, hits int64) bool {
	rv := int64(1)
	if ch.RuleValue != nil {
		rv = int64(*ch.RuleValue)
	}
	switch ch.Rule {
	case domain.RuleMinHitsRequired:
		return hits >= rv
	case domain.RuleThresholdMax:
		return hits <= rv
	default:
		return false
	}
}

func (e *evaluator) controlRow(ctx context.Context, id string, results map[checkSig]checkOutcome, activity map[string]int64) dto.ControlRow {
	row := dto.ControlRow{ControlID: id, Name: id}

	c, ok := e.controls.Get(ctx, id)
	if !ok {
		row.EngineStatus, row.Status = domain.StatusNotCovered, domain.StatusNotCovered
		row.Evidence = "control not found in the library"
		return row
	}
	row.Name = c.Name

	if effectiveScope(c) == domain.ScopeGovernance {
		row.EngineStatus, row.Status = domain.StatusOutOfScope, domain.StatusOutOfScope
		row.Evidence = "governance control — not provable from log data"
		return row
	}

	rules := e.coverage.Rules(id)
	row.Coverage = len(rules)
	for _, name := range rules {
		row.Activity += int(activity[name])
	}

	runnable := runnableChecks(c)
	if len(runnable) == 0 {
		switch {
		case len(c.Checks) > 0:
			row.EngineStatus = domain.StatusPending
			row.Evidence = "check declared but not yet written"
		case len(rules) > 0:
			row.EngineStatus = domain.StatusCompliant
			row.Evidence = fmt.Sprintf("covered by %d rule(s), %d alerts in window", len(rules), row.Activity)
		default:
			row.EngineStatus = domain.StatusNotCovered
			row.Evidence = "no check and no rule coverage"
		}
		row.Status = row.EngineStatus
		return row
	}

	passed, failed, skipped := 0, 0, 0
	var firstFail string
	for _, ch := range runnable {
		res := results[sigOf(ch)]
		row.Checks = append(row.Checks, dto.CheckResult{
			Key:      ch.Key,
			Name:     ch.Name,
			Dataset:  effectiveDataset(ch),
			DataType: ch.DataType,
			Rule:     ch.Rule,
			Required: ch.RuleValue,
			Outcome:  res.outcome,
			Hits:     res.hits,
			Error:    res.err,
		})
		switch res.outcome {
		case domain.CheckPassed:
			passed++
		case domain.CheckFailed:
			failed++
			if firstFail == "" {
				firstFail = fmt.Sprintf("%s (%d hits)", ch.Name, res.hits)
			}
		default: // NOT_APPLICABLE or ERROR — neither passes nor fails the control
			skipped++
		}
	}

	switch {
	case passed+failed == 0:
		row.EngineStatus = domain.StatusNotEvaluated
		row.Evidence = fmt.Sprintf("no data for %d check(s) in the window", skipped)

	case effectiveStrategy(c) == domain.StrategyAny:
		if passed > 0 {
			row.EngineStatus = domain.StatusCompliant
			row.Evidence = fmt.Sprintf("%d of %d checks passed", passed, len(runnable))
		} else {
			row.EngineStatus = domain.StatusNonCompliant
			row.Evidence = firstFail
		}

	case failed > 0:
		row.EngineStatus = domain.StatusNonCompliant
		row.Evidence = firstFail
		if failed > 1 {
			row.Evidence = fmt.Sprintf("%s, and %d more", firstFail, failed-1)
		}

	case skipped > 0:
		row.EngineStatus = domain.StatusAtRisk
		row.Evidence = fmt.Sprintf("%d of %d checks passed, %d could not run", passed, len(runnable), skipped)

	default:
		row.EngineStatus = domain.StatusCompliant
		row.Evidence = fmt.Sprintf("all %d checks passed", passed)
	}

	row.Status = row.EngineStatus
	return row
}

func (e *evaluator) collectChecks(ctx context.Context, controlIDs []string) map[checkSig]domain.Check {
	out := map[checkSig]domain.Check{}
	for _, id := range controlIDs {
		c, ok := e.controls.Get(ctx, id)
		if !ok || effectiveScope(c) == domain.ScopeGovernance {
			continue
		}
		for _, ch := range runnableChecks(c) {
			out[sigOf(ch)] = ch
		}
	}
	return out
}

func (e *evaluator) ruleNames(controlIDs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, id := range controlIDs {
		for _, n := range e.coverage.Rules(id) {
			if !seen[n] {
				seen[n] = true
				out = append(out, n)
			}
		}
	}
	return out
}

func controlIDsOf(fw *domain.Framework) []string {
	seen := map[string]bool{}
	var out []string
	for _, sec := range fw.Sections {
		for _, req := range sec.Requirements {
			for _, cid := range req.SatisfiedBy {
				if !seen[cid] {
					seen[cid] = true
					out = append(out, cid)
				}
			}
		}
	}
	return out
}
