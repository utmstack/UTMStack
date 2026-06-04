import { Component, EventEmitter, Output } from '@angular/core';
import {QUERY_SUGGESTIONS, QuerySuggestion} from './query-suggestions.constants';

@Component({
  selector: 'app-query-suggestions',
  templateUrl: './query-suggestions.component.html',
  styleUrls: ['./query-suggestions.component.scss']
})
export class QuerySuggestionsComponent {

  @Output() selectQuery = new EventEmitter<string>();
  queries: QuerySuggestion[] = QUERY_SUGGESTIONS;

  onSelect(q: string) {
    this.selectQuery.emit(q);
  }
}
