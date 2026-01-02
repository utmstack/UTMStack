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
      query: `SELECT lastEvent.log.action AS action FROM v11-alert-*;`
    },
    {
      label: 'Top 10 errores',
      query: ``
    },
    {
      label: 'Subqueries',
      query: ``
    },
    {
      label: 'Agregacion',
      query: ``
    },
  ];

  onSelect(q: string) {
    this.selectQuery.emit(q);
  }
}
