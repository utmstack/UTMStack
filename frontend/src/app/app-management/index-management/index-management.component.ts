import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {ElasticSearchIndexService} from '../../shared/services/elasticsearch/elasticsearch-index.service';
import {ElasticsearchIndexInfoType} from '../../shared/types/elasticsearch/elasticsearch-index-info.type';
import {IndexDeleteComponent} from './index-delete/index-delete.component';
import {UtmAccountModule} from "../../account/account.module";
import {UtmToastService} from "../../shared/alert/utm-toast.service";

@Component({
  selector: 'app-index-management',
  templateUrl: './index-management.component.html',
  styleUrls: ['./index-management.component.scss']
})
export class IndexManagementComponent implements OnInit {
  searching: any;
  indexes: ElasticsearchIndexInfoType[] = [];
  totalItems: any;
  page = 0;
  itemsPerPage = ITEMS_PER_PAGE;
  loading = false;
  selected: string[] = [];
  req: any;

  constructor(private elasticIndexService: ElasticSearchIndexService,
              private modalService: NgbModal,
              private toastService: UtmToastService,) {
  }

  ngOnInit() {
    this.req = {
      includeSystemIndex: false,
      page: this.page,
      pattern: '*',
      size: this.itemsPerPage
    };
    this.getIndexes();
  }


  searchIndex($event: string) {
    this.req.pattern = $event ? ($event + '*') : '*';
    this.getIndexes();
  }

  onSortBy($event: SortEvent) {
    switch ($event.column) {
      case 'creationDate':
        this.req.sort = 'creation.date.string' + ',' + $event.direction;
        break;

      case 'docsCount':
        this.req.sort = 'docs.count' + ',' + $event.direction;
        break;

      case 'size':
        this.req.sort = 'store.size' + ',' + $event.direction;
        break;

      default:
        this.req.sort = $event.column + ',' + $event.direction;
        break;
    }
    this.getIndexes();
  }

  getIndexes() {
    this.loading = true;
    this.elasticIndexService.getElasticIndex(this.req).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  loadPage($event: number) {
    this.req.page = $event - 1;
    this.getIndexes();
  }

  deleteIndex(index: ElasticsearchIndexInfoType) {
    const modal = this.modalService.open(IndexDeleteComponent, {centered: true});
    modal.componentInstance.indexes = [index.index];
    modal.componentInstance.indexesDeleted.subscribe(() => {
       this.getIndexes();
    });
  }

  isSelected(index: string) {
    return this.selected.findIndex(value => value === index) !== -1;
  }

  addToSelected(index: ElasticsearchIndexInfoType) {
    const i = this.selected.findIndex(value => value === index.index);
    if (i === -1) {
      this.selected.push(index.index);
    } else {
      this.selected.splice(i, 1);
    }
  }

  bulkDelete() {
    const modal = this.modalService.open(IndexDeleteComponent, {centered: true});
    modal.componentInstance.indexes = this.selected;
    modal.componentInstance.indexesDeleted.subscribe(() => {
      this.selected = [];
      this.getIndexes();
    });
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.indexes = data;
    this.loading = false;
  }

  private onError(error) {
    this.toastService.showError('Error', 'An error occurred while listing the indexes');
    this.loading = false;
  }
}
