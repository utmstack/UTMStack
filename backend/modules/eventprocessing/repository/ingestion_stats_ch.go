package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

var ErrNoTenantScope = errors.New("eventprocessing: no tenant in scope")

type chIngestionStatsRepository struct {
	store *eventstore.Store
}

func NewIngestionStatsRepository(s *eventstore.Store) connectors.IngestionStatsRepository {
	return &chIngestionStatsRepository{store: s}
}

// groupColumns is the closed set a widget may break ingestion down by.
var groupColumns = map[string]string{
	"dataSource": "`dataSource`",
	"dataType":   "`dataType`",
}

const tsFormat = "'%Y-%m-%dT%H:%M:%SZ'"

func (r *chIngestionStatsRepository) TotalsByField(ctx context.Context, field string, q connectors.IngestionStatsQuery, top int) ([]dto.IngestionStatsBucket, dto.IngestionTotals, error) {
	col, ok := groupColumns[field]
	if !ok {
		return nil, dto.IngestionTotals{}, fmt.Errorf("eventprocessing: cannot group ingestion by %q", field)
	}

	where, err := r.predicate(ctx, q)
	if err != nil {
		return nil, dto.IngestionTotals{}, err
	}
	if top <= 0 {
		top = 100
	}

	sql := fmt.Sprintf(`SELECT %s AS key, sum(count) AS count, sum(bytes) AS bytes, formatDateTime(max(`+"`@timestamp`"+`), %s) AS lastSeen
		FROM %s WHERE %s GROUP BY key ORDER BY count DESC LIMIT %d`,
		col, tsFormat, r.store.TableName(eventstore.DatasetStats), where, top)

	rows, err := r.fetch(ctx, sql)
	if err != nil {
		return nil, dto.IngestionTotals{}, err
	}
	buckets := make([]dto.IngestionStatsBucket, 0, len(rows))
	for _, row := range rows {
		var b dto.IngestionStatsBucket
		if err := json.Unmarshal(row, &b); err != nil {
			return nil, dto.IngestionTotals{}, err
		}
		buckets = append(buckets, b)
	}

	// The total is the whole window, not the sum of the top N shown.
	totals, err := r.totals(ctx, where)
	if err != nil {
		return nil, dto.IngestionTotals{}, err
	}
	return buckets, totals, nil
}

func (r *chIngestionStatsRepository) totals(ctx context.Context, where string) (dto.IngestionTotals, error) {
	sql := fmt.Sprintf(`SELECT coalesce(sum(count), 0) AS events, coalesce(sum(bytes), 0) AS bytes
		FROM %s WHERE %s`, r.store.TableName(eventstore.DatasetStats), where)

	rows, err := r.fetch(ctx, sql)
	if err != nil {
		return dto.IngestionTotals{}, err
	}
	var t dto.IngestionTotals
	if len(rows) > 0 {
		if err := json.Unmarshal(rows[0], &t); err != nil {
			return dto.IngestionTotals{}, err
		}
	}
	return t, nil
}

func (r *chIngestionStatsRepository) Timeline(ctx context.Context, q connectors.IngestionStatsQuery, interval string) ([]dto.TimelinePoint, error) {
	where, err := r.predicate(ctx, q)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT formatDateTime(%s, %s) AS timestamp, sum(count) AS count, sum(bytes) AS bytes
		FROM %s WHERE %s GROUP BY timestamp ORDER BY timestamp`,
		bucketExpr(interval), tsFormat, r.store.TableName(eventstore.DatasetStats), where)

	rows, err := r.fetch(ctx, sql)
	if err != nil {
		return nil, err
	}
	points := make([]dto.TimelinePoint, 0, len(rows))
	for _, row := range rows {
		var p dto.TimelinePoint
		if err := json.Unmarshal(row, &p); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, nil
}

func (r *chIngestionStatsRepository) TimelineByField(ctx context.Context, field string, q connectors.IngestionStatsQuery, interval string, top int) ([]dto.TimelineSeries, error) {
	col, ok := groupColumns[field]
	if !ok {
		return nil, fmt.Errorf("eventprocessing: cannot group ingestion by %q", field)
	}

	// Which lines to draw is decided over the whole window, so a series does
	// not appear and vanish depending on the bucket.
	leaders, _, err := r.TotalsByField(ctx, field, q, top)
	if err != nil {
		return nil, err
	}
	if len(leaders) == 0 {
		return []dto.TimelineSeries{}, nil
	}
	quoted := make([]string, 0, len(leaders))
	for _, b := range leaders {
		quoted = append(quoted, chQuote(b.Key))
	}

	where, err := r.predicate(ctx, q)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s AS key, formatDateTime(%s, %s) AS timestamp, sum(count) AS count, sum(bytes) AS bytes
		FROM %s WHERE %s AND %s IN (%s) GROUP BY key, timestamp ORDER BY key, timestamp`,
		col, bucketExpr(interval), tsFormat, r.store.TableName(eventstore.DatasetStats), where, col, strings.Join(quoted, ", "))

	rows, err := r.fetch(ctx, sql)
	if err != nil {
		return nil, err
	}

	byKey := make(map[string][]dto.TimelinePoint, len(leaders))
	for _, row := range rows {
		var p struct {
			Key string `json:"key"`
			dto.TimelinePoint
		}
		if err := json.Unmarshal(row, &p); err != nil {
			return nil, err
		}
		byKey[p.Key] = append(byKey[p.Key], p.TimelinePoint)
	}

	// Ordered by how much each line carries, like the buckets it came from.
	out := make([]dto.TimelineSeries, 0, len(leaders))
	for _, b := range leaders {
		out = append(out, dto.TimelineSeries{Key: b.Key, Points: byKey[b.Key]})
	}
	return out, nil
}

// fetch runs a read through the store: the query already carries the tenant
// predicate from predicate(), so Store.Query's AllTenants scope applies.
func (r *chIngestionStatsRepository) fetch(ctx context.Context, sql string) ([]json.RawMessage, error) {
	return r.store.Query(ctx, sql)
}

// predicate builds the WHERE every read shares. The tenant comes from the
// session: a caller cannot name one.
func (r *chIngestionStatsRepository) predicate(ctx context.Context, q connectors.IngestionStatsQuery) (string, error) {
	parts := make([]string, 0, 4)

	tenant := authz.TenantIDFromContext(ctx)
	switch {
	case tenant != "":
		parts = append(parts, "tenantId = "+chQuote(tenant))
	case tenancy.Enabled():
		return "", ErrNoTenantScope
	default:
		// A single-tenant install reads everything it holds, which is its own.
	}

	if !q.From.IsZero() {
		parts = append(parts, "`@timestamp` >= "+chQuote(q.From.UTC().Format(time.RFC3339)))
	}
	if !q.To.IsZero() {
		parts = append(parts, "`@timestamp` <= "+chQuote(q.To.UTC().Format(time.RFC3339)))
	}
	if q.Type != "" {
		parts = append(parts, "type = "+chQuote(q.Type))
	}
	if q.DataSource != "" {
		parts = append(parts, "`dataSource` = "+chQuote(q.DataSource))
	}

	if len(parts) == 0 {
		return "1", nil
	}
	return strings.Join(parts, " AND "), nil
}

// chQuote turns a value into a ClickHouse string literal: backslash and quote
// are the two characters that can end a literal early.
func chQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return "'" + s + "'"
}

// bucketExpr turns a bucket size into the expression that groups by it. The
// unit cannot be a bound parameter, so anything unrecognised becomes an hour
// rather than reaching the statement.
func bucketExpr(interval string) string {
	n, unit := parseBucket(interval)
	return fmt.Sprintf("toStartOfInterval(`@timestamp`, INTERVAL %d %s)", n, unit)
}

var bucketUnits = map[byte]string{
	's': "SECOND",
	'm': "MINUTE",
	'h': "HOUR",
	'd': "DAY",
	'w': "WEEK",
}

func parseBucket(interval string) (int, string) {
	s := strings.TrimSpace(interval)
	if s == "1M" || strings.EqualFold(s, "month") {
		return 1, "MONTH"
	}
	if len(s) < 2 {
		return 1, "HOUR"
	}

	unit, ok := bucketUnits[strings.ToLower(s)[len(s)-1]]
	if !ok {
		return 1, "HOUR"
	}
	n, err := strconv.Atoi(s[:len(s)-1])
	if err != nil || n <= 0 {
		return 1, "HOUR"
	}
	return n, unit
}
