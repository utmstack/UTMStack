package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/tidwall/gjson"
	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
)

type relatedLogsRepo struct {
	connectors.AlertRepository
	raw           json.RawMessage
	echoes        []domain.UtmAlert
	offsets       []int
	echoErr       error
	refs          []domain.LogRef
	refsTruncated bool
	refsAnchor    time.Time
}

func (r *relatedLogsRepo) GetRawByID(context.Context, string) (json.RawMessage, error) {
	return r.raw, nil
}
func (r *relatedLogsRepo) ListEchoes(_ context.Context, parent string, offset, size int, sortBy, order string) ([]domain.UtmAlert, int64, error) {
	if parent != "parent" || sortBy != "@timestamp" || order != "asc" {
		panic("unexpected echo query")
	}
	r.offsets = append(r.offsets, offset)
	if r.echoErr != nil {
		return nil, 0, r.echoErr
	}
	if offset >= len(r.echoes) {
		return nil, int64(len(r.echoes)), nil
	}
	end := offset + size
	if end > len(r.echoes) {
		end = len(r.echoes)
	}
	return r.echoes[offset:end], int64(len(r.echoes)), nil
}
func (r *relatedLogsRepo) RelatedLogRefs(_ context.Context, _ []domain.CorrelationStep, anchor time.Time, _ int) ([]domain.LogRef, bool, error) {
	r.refsAnchor = anchor
	return r.refs, r.refsTruncated, nil
}

type relatedResolver struct{ raw json.RawMessage }

func (r relatedResolver) AfterEventsByRuleName(string) (json.RawMessage, bool) { return r.raw, true }

func parentLog() json.RawMessage {
	return json.RawMessage(`{"id":"parent","name":"Login","@timestamp":"2026-01-02 10:00:00.123456789","dataType":"test","events":[{"id":"first","@timestamp":"2026-01-02 09:59:00.987654321"}],"lastEvent":{"id":"first","dataType":"test","timestamp":"2026-01-02 10:00:00.123456789"}}`)
}
func echoLog(id, ts string) domain.UtmAlert {
	return domain.UtmAlert{Timestamp: ts, Events: []domain.AlertEvent{{ID: id, Timestamp: ts}}, LastEvent: json.RawMessage(fmt.Sprintf(`{"id":%q,"timestamp":%q}`, id, ts))}
}

func TestRelatedLogsIncludesEchoes(t *testing.T) {
	for _, after := range []string{"", "null", "[]", "invalid"} {
		t.Run(after, func(t *testing.T) {
			r := &relatedLogsRepo{raw: parentLog(), echoes: []domain.UtmAlert{echoLog("second", "2026-01-02 10:20:00.123456789"), echoLog("second", "2026-01-02T10:21:00Z")}}
			u := &alertUsecase{repo: r}
			if after != "" {
				u.resolver = relatedResolver{json.RawMessage(after)}
			}
			got, err := u.RelatedLogs(context.Background(), "parent")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.IDs, []string{"first", "second"}) {
				t.Fatalf("IDs=%v", got.IDs)
			}
			if got.Total != 2 || got.Truncated || got.DataType != "test" {
				t.Fatalf("response=%+v", got)
			}
			if got.TimeFrom != "2026-01-02T08:59:00Z" || got.TimeTo != "2026-01-02T10:26:00Z" {
				t.Fatalf("range=%s - %s", got.TimeFrom, got.TimeTo)
			}
		})
	}
}
func TestRelatedLogsEchoPagination(t *testing.T) {
	r := &relatedLogsRepo{raw: parentLog()}
	for i := 0; i < 101; i++ {
		r.echoes = append(r.echoes, echoLog(fmt.Sprint(i), "2026-01-02T10:10:00Z"))
	}
	got, err := (&alertUsecase{repo: r}).RelatedLogs(context.Background(), "parent")
	if err != nil {
		t.Fatal(err)
	}
	if got.Total != 102 || got.Truncated || !reflect.DeepEqual(r.offsets, []int{0, 100}) {
		t.Fatalf("total=%d truncated=%v offsets=%v", got.Total, got.Truncated, r.offsets)
	}
}
func TestRelatedLogsNoEchoes(t *testing.T) {
	got, err := (&alertUsecase{repo: &relatedLogsRepo{raw: parentLog()}}).RelatedLogs(context.Background(), "parent")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.IDs, []string{"first"}) || got.Truncated {
		t.Fatalf("response=%+v", got)
	}
}
func TestRelatedLogsEchoError(t *testing.T) {
	want := errors.New("echo query failed")
	_, err := (&alertUsecase{repo: &relatedLogsRepo{raw: parentLog(), echoErr: want}}).RelatedLogs(context.Background(), "parent")
	if !errors.Is(err, want) {
		t.Fatalf("error=%v", err)
	}
}
func TestRelatedLogsLimits(t *testing.T) {
	for _, count := range []int{998, 999, 1000, 1001} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			r := &relatedLogsRepo{raw: parentLog()}
			for i := 0; i < count; i++ {
				r.echoes = append(r.echoes, echoLog(fmt.Sprint(i), "2026-01-02T10:10:00Z"))
			}
			got, err := (&alertUsecase{repo: r}).RelatedLogs(context.Background(), "parent")
			if err != nil {
				t.Fatal(err)
			}
			want := count + 1
			if want > relatedMaxSize {
				want = relatedMaxSize
			}
			if got.Total != want || got.Truncated != (count+1 > relatedMaxSize) {
				t.Fatalf("total=%d truncated=%v", got.Total, got.Truncated)
			}
		})
	}
}
func TestRelatedLogsKeepsCorrelationReferences(t *testing.T) {
	r := &relatedLogsRepo{raw: parentLog(), echoes: []domain.UtmAlert{echoLog("second", "2026-01-02T10:20:00Z")}, refs: []domain.LogRef{{ID: "second"}, {ID: "correlated"}}, refsTruncated: true}
	u := &alertUsecase{repo: r, resolver: relatedResolver{json.RawMessage(`[{"within":"2h","count":1}]`)}}
	got, err := u.RelatedLogs(context.Background(), "parent")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.IDs, []string{"first", "second", "correlated"}) || !got.Truncated {
		t.Fatalf("response=%+v", got)
	}
	if got.TimeFrom != "2026-01-02T07:59:00Z" {
		t.Fatal(got.TimeFrom)
	}
	if r.refsAnchor.Format(time.RFC3339Nano) != "2026-01-02T10:00:00.123456789Z" {
		t.Fatal(r.refsAnchor)
	}
}
func TestParseRelatedTime(t *testing.T) {
	want := time.Date(2026, 1, 2, 10, 0, 0, 123456789, time.UTC)
	for _, s := range []string{"2026-01-02T10:00:00.123456789Z", "2026-01-02 10:00:00.123456789", "2026-01-02T10:00:00.123456789", "2026-01-02T13:00:00.123456789+03:00"} {
		if got := parseRelatedTime(s); !got.Equal(want) {
			t.Errorf("%q: %s", s, got)
		}
	}
	for _, s := range []string{"", "invalid"} {
		if !parseRelatedTime(s).IsZero() {
			t.Errorf("accepted %q", s)
		}
	}
	if got := eventTime(gjson.Parse(`{"@timestamp":"invalid","timestamp":"2026-01-02 10:00:00.123456789"}`)); !got.Equal(want) {
		t.Fatal(got)
	}
}
