import {HttpResponse} from '@angular/common/http';
import {Component, Input, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ADMIN_ROLE} from '../../shared/constants/global.constant';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {LogstashService} from '../../shared/services/logstash/logstash.service';
import {SVG_MODULE_MAP} from '../shared/const/logstash.const';
import {LogstashFilterType} from '../shared/types/logstash-filter.type';
import {UtmPipeline} from '../shared/types/logstash-stats.type';

@Component({
  selector: 'app-logstash-filters',
  templateUrl: './logstash-filters.component.html',
  styleUrls: ['./logstash-filters.component.scss']
})
export class LogstashFiltersComponent implements OnInit {
  @Input() pipeline: UtmPipeline;
  filters: LogstashFilterType[] = [];
  loading = true;
  totalItems: any;
  requestParams =
    {
      page: 0,
      size: ITEMS_PER_PAGE,
      'filterName.contains': null,
      'isActive.equals': true,
      pipelineId: null,
      sort: 'id,asc'
    };
  openEditJson: any;
  filterEdit: LogstashFilterType;
  lineColor = 'green';
  svgMap = SVG_MODULE_MAP;
  roleAdmin = ADMIN_ROLE;

  constructor(private logstashFilterService: LogstashService,
              private utmToastService: UtmToastService,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    if (this.pipeline) {
      this.requestParams.pipelineId = this.pipeline.id;
      this.lineColor = this.pipeline.pipelineStatus === 'up' ? 'green' : 'red';
      this.getLogsFilters();
    }
  }

  loadPage(page: number) {
    this.requestParams.page = page - 1;
    this.getLogsFilters();
  }

  getLogsFilters() {
    this.logstashFilterService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  editFilter(filter: LogstashFilterType) {
    this.filterEdit = filter;
    this.openEditJson = true;
  }

  onFilterEditClose() {
    this.requestParams.page = 0;
    this.getLogsFilters();
    this.openEditJson = false;
  }

  deleteFilter(filter: LogstashFilterType) {
    const modalFilter = this.modalService.open(ModalConfirmationComponent, {centered: true});
    modalFilter.componentInstance.header = 'Delete asset';
    modalFilter.componentInstance.message = 'Are you sure that you want to delete this filter';
    modalFilter.componentInstance.confirmBtnText = 'Delete';
    modalFilter.componentInstance.confirmBtnIcon = 'icon-database-edit2';
    modalFilter.componentInstance.confirmBtnType = 'delete';
    modalFilter.result.then(() => {
      this.delete(filter);
    });
  }

  delete(filter: LogstashFilterType) {
    this.logstashFilterService.delete(filter.id).subscribe(() => {
      this.utmToastService.showSuccessBottom('Asset deleted successfully');
      this.requestParams.page = 0;
      this.getLogsFilters();
    }, () => {
      this.utmToastService.showError('Error deleting filter',
        'Error while trying to delete asset, please try again');
    });
  }

  private onSuccess(data: any[], headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.filters = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  onSortBy($event: SortEvent) {
    this.requestParams.sort = $event.column + ',' + $event.direction;
    this.getLogsFilters();
  }

  onSearch($event: string) {
    this.requestParams['filterName.contains'] = $event;
    this.requestParams.page = 0;
    this.getLogsFilters();
  }

  getModuleSvg(module: string): string {
    const icon = this.svgMap[module];
    return icon ? icon : 'generic.svg';
  }
}
