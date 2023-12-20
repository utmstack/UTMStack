import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
// tslint:disable-next-line:max-line-length
import {
  IndexPatternCreateComponent
} from '../../../shared/components/utm/index-pattern/index-pattern-create/index-pattern-create.component';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {IndexPatternService} from '../../../shared/services/elasticsearch/index-pattern.service';
import {UtmIndexPattern} from '../../../shared/types/index-pattern/utm-index-pattern';
import {SortByType} from '../../../shared/types/sort-by.type';
import {IndexPatternDeleteComponent} from '../index-pattern-delete/index-pattern-delete.component';
import {IndexPatternHelpComponent} from '../shared/components/index-pattern-help/index-pattern-help.component';

@Component({
  selector: 'app-index-pattern-list',
  templateUrl: './index-pattern-list.component.html',
  styleUrls: ['./index-pattern-list.component.scss']
})
export class IndexPatternListComponent implements OnInit {
  patterns: UtmIndexPattern[] = [];
  loading = true;
  fields: SortByType[] = [{field: 'pattern', fieldName: 'Index Pattern'}];
  totalItems: any;
  page = 0;
  itemsPerPage = ITEMS_PER_PAGE;
  indexSearch = '';
  searching = false;
  private sortBy = 'id,asc';

  constructor(private modalService: NgbModal,
              private indexPatternService: IndexPatternService) {
  }

  ngOnInit() {
    this.getIndexPatterns();
  }

  viewIndexHelp() {
    const modal = this.modalService.open(IndexPatternHelpComponent, {centered: true});
  }

  createIndexPattern() {
    const modal = this.modalService.open(IndexPatternCreateComponent, {centered: true});
    modal.componentInstance.indexPatternCreated.subscribe(created => {
      this.getIndexPatterns();
    });
  }

  editIndexPattern(pattern: UtmIndexPattern) {
    const modal = this.modalService.open(IndexPatternCreateComponent, {centered: true});
    modal.componentInstance.indexPattern = pattern;
    modal.componentInstance.indexPatternCreated.subscribe(edited => {
      this.getIndexPatterns();
    });
  }

  deleteIndexPattern(indexPattern: UtmIndexPattern) {
    const modal = this.modalService.open(IndexPatternDeleteComponent, {centered: true});
    modal.componentInstance.indexPattern = indexPattern;
    modal.componentInstance.indexPatternDeleted.subscribe(deleted => {
      this.getIndexPatterns();
    });
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.getIndexPatterns();
  }

  loadPage(page: number) {
    this.page = page - 1;
    this.getIndexPatterns();
  }

  searchIndex($event: string) {
    this.searching = true;
    this.indexSearch = $event;
    this.getIndexPatterns();
  }

  getIndexPatterns() {
    const req = {
      page: this.page,
      size: this.itemsPerPage,
      'pattern.contains': this.indexSearch,
      'isActive.equals': true,
      sort: this.sortBy
    };
    this.indexPatternService.query(req).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.patterns = data;
    this.loading = false;
    this.searching = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  getModules(pattern: UtmIndexPattern) {
    return pattern.patternModule ? pattern.patternModule.split(',') : [];
  }
}
