import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {IndexPatternService} from '../../../../../shared/services/elasticsearch/index-pattern.service';
import {UtmIndexPattern} from '../../../../../shared/types/index-pattern/utm-index-pattern';
import {UtmComplianceQueryConfigType} from '../../../type/compliance-query-config.type';


@Component({
  selector: 'app-utm-compliance-query-list',
  templateUrl: './utm-compliance-query-list.component.html',
  styleUrls: ['./utm-compliance-query-list.component.scss']
})
export class UtmComplianceQueryListComponent implements OnInit {

  queries: UtmComplianceQueryConfigType[] = [];

  @Output() add = new EventEmitter<UtmComplianceQueryConfigType>();
  @Output() edit = new EventEmitter<{ index: number, query: UtmComplianceQueryConfigType }>();
  @Output() delete = new EventEmitter<number>();

  pattern: UtmIndexPattern;
  patterns: UtmIndexPattern[];
  indexPatternNames = [];

  editingIndex: number = null;

  ngOnInit() {
    this.getIndexPatterns();
  }

  constructor(private indexPatternService: IndexPatternService) {

  }

  startEdit(index: number) {
    this.editingIndex = index;
  }

  finishEdit(query: UtmComplianceQueryConfigType) {
    if (this.editingIndex !== null && this.queries) {
      this.queries[this.editingIndex] = query;
    }

    this.editingIndex = null;
  }


  cancelEdit() {
    this.editingIndex = null;
  }

  remove(index: number) {
    this.queries = this.queries.filter((_, i) => i !== index);
  }

  onQueryAdd(query: UtmComplianceQueryConfigType): void {
    this.queries = [...this.queries, query];
  }

  //TODO: ELENA revisar de aqui para abajo
  onIndexPatternChange($event: any) {
    this.pattern = $event;
    //this.indexPatternChange.emit($event);
  }

  getIndexPatterns() {
    const req = {
      page: 0,
      size: 1000,
      sort: 'id,asc',
      'isActive.equals': true,
    };
    this.indexPatternService.query(req).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  private onSuccess(data, headers) {
    this.patterns = data;
    this.indexPatternNames = this.getListPatterns().map(f => f.name);
  }

  getListPatterns() {
    return (this.patterns || [])
      .filter(pattern => pattern && pattern.id != null)
      .map(pattern => ({
        id: pattern.id,
        name: pattern.pattern,
      }));
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }


  //TODO: ELENA revisar si puedo reutilizar los nombres
  getIndexPatternName(id: number) {
    if (!this.patterns || !id) {
      return '';
    }

    const pattern = this.patterns.find(p => p.id === id);
    return pattern ? pattern.pattern : '';
  }


}
