export interface QuerySuggestion {
  label: string;
  query: string;
}

export const QUERY_SUGGESTIONS: QuerySuggestion[] = [
  {
    label: 'Selecting Fields',
    query: `SELECT lastEvent.log.action, severity
            FROM v11-alert-*
            LIMIT 20;`
  },
  {
    label: 'Filtering by Field Value',
    query: `SELECT *
            FROM v11-alert-*
            WHERE severityLabel = 'Low';`
  },
  {
    label: 'Filtering by Time Range (Last 24 Hours)',
    query: `SELECT *
            FROM v11-alert-*
            WHERE @timestamp BETWEEN DATE_SUB(NOW(), INTERVAL 24 HOUR) AND NOW();`
  },
  {
    label: 'Filtering by Time Range (Last 30 Days)',
    query: `SELECT @timestamp
            FROM v11-log-*
            WHERE @timestamp BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
            ORDER BY @timestamp DESC
            LIMIT 5;`
  },
  {
    label: 'Selecting Nested Field',
    query: `SELECT lastEvent.log.action AS action
            FROM v11-alert-*
            LIMIT 5;`
  },
  {
    label: 'Using Aggregations (COUNT)',
    query: `SELECT lastEvent.log.action AS action, COUNT(*) AS total
            FROM v11-alert-*
            WHERE @timestamp BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
            GROUP BY action
            ORDER BY total DESC
            LIMIT 10;`
  },
  {
    label: 'Using Aggregations (MAX)',
    query: `SELECT lastEvent.log.action AS action, MAX(@timestamp) AS lastSeen
            FROM v11-alert-*
            GROUP BY action;`
  },
  {
    label: 'Grouping Events by Day',
    query: `SELECT DATE_FORMAT(@timestamp, 'yyyy-MM-dd') AS day, COUNT(*) AS total
            FROM v11-alert-*
            GROUP BY day
            ORDER BY day DESC;`
  },
  {
    label: 'Filtering by Match',
    query: `SELECT severityLabel
            FROM v11-alert-*
            WHERE MATCH(severityLabel, 'Low');`
  },
  {
    label: 'Using Subqueries',
    query: `SELECT action, total
            FROM (
              SELECT lastEvent.log.action AS action, COUNT(*) AS total
              FROM v11-alert-*
              WHERE @timestamp BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
              GROUP BY action
            ) AS sub
            WHERE total > 50
            ORDER BY total DESC;`
  }
];
