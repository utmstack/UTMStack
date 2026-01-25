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
  // tslint:disable-next-line:variable-name
  private _queries: UtmComplianceQueryConfigType[] = [];
  @Input() set queries(value: UtmComplianceQueryConfigType[]) {
    this._queries = value || [];
  }

  get queries() {
    return this._queries;
  }

  @Output() add = new EventEmitter<{
    query: UtmComplianceQueryConfigType;
    index: number | null;
  }>();
  @Output() remove = new EventEmitter<number>();

  patterns: UtmIndexPattern[];
  indexPatternNames = [];
  editingIndex: number = null;

  constructor(private indexPatternService: IndexPatternService) {}

  ngOnInit() {
    this.getIndexPatterns();
  }

  startEdit(index: number) {
    this.editingIndex = index;
  }

  cancelEdit() {
    this.editingIndex = null;
  }

  onQueryAdd(query: UtmComplianceQueryConfigType): void {
    this.add.emit({
      query,
      index: this.editingIndex
    });

    this.editingIndex = null;
  }

  onRemove(index: number) {
    this.remove.emit(index);
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
    //this.alertService.error(error.error, error.message, null);
  }

  getIndexPatternName(id: number) {
    if (!this.patterns || !id) {
      return '';
    }

    const pattern = this.patterns.find(p => p.id === id);
    return pattern ? pattern.pattern : '';
  }
}
