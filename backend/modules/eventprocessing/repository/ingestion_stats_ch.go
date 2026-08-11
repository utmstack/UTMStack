package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/threatwinds/go-sdk/store"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

var ErrNoTenantScope = errors.New("eventprocessing: no tenant in scope")

// The statistics dataset is already aggregated: the stats plugin writes one row
// per (tenant, topic, source, data type) every cycle with how many events and
// bytes it saw. Answering "how much was ingested" is therefore a sum over those
// counters, and the store's reader only counts records — so these three reads
// go to SQL directly, the same escape hatch the log explorer's SQL mode uses.
//
// Everything that varies is a bound parameter. The only text interpolated into
// a statement is the grouping column and the bucket size, and both come from a
// whitelist below rather than from the request.
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

func (r *chIngestionStatsRepository) TotalsByField(ctx context.Context, field string, q connectors.IngestionStatsQuery, top int) ([]dto.IngestionStatsBucket, dto.IngestionTotals, error) {
	col, ok := groupColumns[field]
	if !ok {
		return nil, dto.IngestionTotals{}, fmt.Errorf("eventprocessing: cannot group ingestion by %q", field)
	}

	where, args, err := r.predicate(ctx, q)
	if err != nil {
		return nil, dto.IngestionTotals{}, err
	}
	if top <= 0 {
		top = 100
	}

	sql := fmt.Sprintf(`SELECT %s AS k, sum(count) AS events, sum(bytes) AS bytes, max(`+"`@timestamp`"+`) AS last
		FROM %s WHERE %s GROUP BY k ORDER BY events DESC LIMIT ?`,
		col, r.store.TableName(eventstore.DatasetStats), where)

	rows, err := r.store.Conn.Query(readOnly(ctx), sql, append(args, top)...)
	if err != nil {
		return nil, dto.IngestionTotals{}, err
	}
	defer rows.Close()

	buckets := make([]dto.IngestionStatsBucket, 0, top)
	for rows.Next() {
		var (
			key           string
			events, bytes uint64
			last          time.Time
		)
		if err := rows.Scan(&key, &events, &bytes, &last); err != nil {
			return nil, dto.IngestionTotals{}, err
		}
		if key == "" {
			continue
		}
		buckets = append(buckets, dto.IngestionStatsBucket{
			Key:      key,
			Count:    int64(events),
			Bytes:    int64(bytes),
			LastSeen: last.UTC().Format(time.RFC3339),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, dto.IngestionTotals{}, err
	}

	// The total is the whole window, not the sum of the top N shown.
	totals, err := r.totals(ctx, where, args)
	if err != nil {
		return nil, dto.IngestionTotals{}, err
	}
	return buckets, totals, nil
}

func (r *chIngestionStatsRepository) totals(ctx context.Context, where string, args []any) (dto.IngestionTotals, error) {
	sql := fmt.Sprintf("SELECT sum(count), sum(bytes) FROM %s WHERE %s",
		r.store.TableName(eventstore.DatasetStats), where)

	var events, bytes uint64
	if err := r.store.Conn.QueryRow(readOnly(ctx), sql, args...).Scan(&events, &bytes); err != nil {
		return dto.IngestionTotals{}, err
	}
	return dto.IngestionTotals{Events: int64(events), Bytes: int64(bytes)}, nil
}

func (r *chIngestionStatsRepository) Timeline(ctx context.Context, q connectors.IngestionStatsQuery, interval string) ([]dto.TimelinePoint, error) {
	where, args, err := r.predicate(ctx, q)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s AS t, sum(count) AS events, sum(bytes) AS bytes
		FROM %s WHERE %s GROUP BY t ORDER BY t`,
		bucketExpr(interval), r.store.TableName(eventstore.DatasetStats), where)

	rows, err := r.store.Conn.Query(readOnly(ctx), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]dto.TimelinePoint, 0)
	for rows.Next() {
		var (
			at            time.Time
			events, bytes uint64
		)
		if err := rows.Scan(&at, &events, &bytes); err != nil {
			return nil, err
		}
		points = append(points, dto.TimelinePoint{
			Timestamp: at.UTC().Format(time.RFC3339),
			Count:     int64(events),
			Bytes:     int64(bytes),
		})
	}
	return points, rows.Err()
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
	keys := make([]string, 0, len(leaders))
	for _, b := range leaders {
		keys = append(keys, b.Key)
	}

	where, args, err := r.predicate(ctx, q)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s AS k, %s AS t, sum(count) AS events, sum(bytes) AS bytes
		FROM %s WHERE %s AND %s IN (?) GROUP BY k, t ORDER BY k, t`,
		col, bucketExpr(interval), r.store.TableName(eventstore.DatasetStats), where, col)

	rows, err := r.store.Conn.Query(readOnly(ctx), sql, append(args, keys)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byKey := make(map[string][]dto.TimelinePoint, len(keys))
	for rows.Next() {
		var (
			key           string
			at            time.Time
			events, bytes uint64
		)
		if err := rows.Scan(&key, &at, &events, &bytes); err != nil {
			return nil, err
		}
		byKey[key] = append(byKey[key], dto.TimelinePoint{
			Timestamp: at.UTC().Format(time.RFC3339),
			Count:     int64(events),
			Bytes:     int64(bytes),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Ordered by how much each line carries, like the buckets it came from.
	out := make([]dto.TimelineSeries, 0, len(keys))
	for _, k := range keys {
		out = append(out, dto.TimelineSeries{Key: k, Points: byKey[k]})
	}
	return out, nil
}

// predicate builds the WHERE every read shares. The tenant is a bound
// parameter and comes from the session: a caller cannot name one.
func (r *chIngestionStatsRepository) predicate(ctx context.Context, q connectors.IngestionStatsQuery) (string, []any, error) {
	parts := make([]string, 0, 5)
	args := make([]any, 0, 5)

	tenant := authz.TenantIDFromContext(ctx)
	switch {
	case tenant != "":
		parts = append(parts, "tenantId = ?")
		args = append(args, tenant)
	case tenancy.Enabled():
		return "", nil, ErrNoTenantScope
	default:
		// A single-tenant install reads everything it holds, which is its own.
		_ = store.AllTenants
	}

	if !q.From.IsZero() {
		parts = append(parts, "`@timestamp` >= ?")
		args = append(args, q.From.UTC())
	}
	if !q.To.IsZero() {
		parts = append(parts, "`@timestamp` <= ?")
		args = append(args, q.To.UTC())
	}
	if q.Type != "" {
		parts = append(parts, "type = ?")
		args = append(args, q.Type)
	}
	if q.DataSource != "" {
		parts = append(parts, "`dataSource` = ?")
		args = append(args, q.DataSource)
	}

	if len(parts) == 0 {
		return "1", args, nil
	}
	return strings.Join(parts, " AND "), args, nil
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

// readOnly is the same guard the analyzer's SQL mode runs under: a read that
// cannot write, cannot run forever, and cannot drag the whole table back.
func readOnly(ctx context.Context) context.Context {
	return clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{
		"readonly":           1,
		"max_execution_time": 30,
	}))
}
