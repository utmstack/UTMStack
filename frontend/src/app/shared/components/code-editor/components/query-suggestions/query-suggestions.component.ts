import { Component, EventEmitter, Output } from '@angular/core';

@Component({
  selector: 'app-query-suggestions',
  templateUrl: './query-suggestions.component.html',
  styleUrls: ['./query-suggestions.component.scss']
})
export class QuerySuggestionsComponent {

  @Output() selectQuery = new EventEmitter<string>();

  queries = [
    {
      label: 'Filtering by Time Range (Last 30 Days)',
      query: `SELECT @timestamp FROM v11-log-* WHERE @timestamp BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW() ORDER BY @timestamp DESC LIMIT 5;`
    },
    {
      label: 'Selecting Nested Field',
      query: `SELECT lastEvent.log.action AS action FROM v11-alert-* LIMIT 5;`
    },
    {
      label: 'Top 10 errores',
      query: ``
    },
    {
      label: 'Using Aggregations',
      query: `SELECT lastEvent.log.action AS action, COUNT(*) AS total FROM v11-alert-* WHERE @timestamp BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW() GROUP BY action ORDER BY total DESC LIMIT 10;`
    },
    {
      label: 'Using Subqueries',
      query: `SELECT action, total FROM ( SELECT lastEvent.log.action AS action, COUNT(*) AS total FROM v11-alert-* WHERE @timestamp BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW() GROUP BY action ) AS sub WHERE total > 50 ORDER BY total DESC;`
    },
  ];

  onSelect(q: string) {
    this.selectQuery.emit(q);
  }
}
