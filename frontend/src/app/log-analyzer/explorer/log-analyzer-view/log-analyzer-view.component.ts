import {HttpResponse} from '@angular/common/http';
import {Component, HostListener, Input, OnDestroy, OnInit, Type} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import * as moment from 'moment';
import {Observable, Subject} from 'rxjs';
import {filter, takeUntil} from 'rxjs/operators';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {
  ElasticFilterDefaultTime
} from '../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {
  UtmFilterBehavior
} from '../../../shared/components/utm/filters/utm-elastic-filter/shared/behavior/utm-filter.behavior';
import {
  UtmTableDetailComponent
} from '../../../shared/components/utm/table/utm-table/utm-table-detail/utm-table-detail.component';
import {ADMIN_ROLE} from '../../../shared/constants/global.constant';
import {LOG_ANALYZER_TOTAL_ITEMS} from '../../../shared/constants/log-analyzer.constant';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {ElasticDataTypesEnum} from '../../../shared/enums/elastic-data-types.enum';
import {ElasticOperatorsEnum} from '../../../shared/enums/elastic-operators.enum';
import {DataNatureTypeEnum, NatureDataPrefixEnum} from '../../../shared/enums/nature-data.enum';
import {ElasticDataExportService} from '../../../shared/services/elasticsearch/elastic-data-export.service';
import {ElasticDataService} from '../../../shared/services/elasticsearch/elastic-data.service';
import {TimezoneFormatService} from '../../../shared/services/utm-timezone.service';
import {DatePipeDefaultOptions} from '../../../shared/types/date-pipe-default-options';
import {ElasticSearchFieldInfoType} from '../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {UtmIndexPattern} from '../../../shared/types/index-pattern/utm-index-pattern';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {parseQueryParamsToFilter} from '../../../shared/util/query-params-to-filter.util';
import {
  LogAnalyzerQueryCreateComponent
} from '../../queries/log-analyzer-query-create/log-analyzer-query-create.component';
import {IndexFieldController} from '../../shared/behaviors/index-field-controller.behavior';
import {IndexPatternBehavior} from '../../shared/behaviors/index-pattern.behavior';
import {LogFilterBehavior} from '../../shared/behaviors/log-filter.behavior';
import {QueryRunBehavior} from '../../shared/behaviors/query-run.behavior';
import {TabService} from '../../shared/services/tab.service';
import {LogAnalyzerQueryType} from '../../shared/type/log-analyzer-query.type';

@Component({
  selector: 'app-log-analyzer-view',
  templateUrl: './log-analyzer-view.component.html',
  styleUrls: ['./log-analyzer-view.component.scss']
})
export class LogAnalyzerViewComponent implements OnInit, OnDestroy {
  @Input() data: LogAnalyzerQueryType;
  @Input() uuid: string;
  fields: UtmFieldType[] = [];
  rows: any[] = [];
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  totalItems = LOG_ANALYZER_TOTAL_ITEMS;
  view: 'table' | 'chart' = 'table';
  filters: ElasticFilterType[] = [{
    field: '@timestamp',
    operator: ElasticOperatorsEnum.IS_BETWEEN,
    value: ['now-24h', 'now']
  }];
  selectedFields: ElasticSearchFieldInfoType[] = [{name: '@timestamp', type: ElasticDataTypesEnum.DATE}];
  dataNature: string = DataNatureTypeEnum.EVENT;
  queryParams: any;
  counter: any;
  runningQuery: boolean;
  pattern: UtmIndexPattern;
  error = false;
  loading = true;
  componentDetail: Type<any> = UtmTableDetailComponent;
  fieldWidth = '300px';
  detailWidth: number;
  pageWidth = window.innerWidth;
  //
  csvExport = false;
  admin = ADMIN_ROLE;
  private timer: number;
  private sortBy = NatureDataPrefixEnum.TIMESTAMP + ',' + 'desc';
  patterns: UtmIndexPattern[];
  paramLoaded = false;
  defaultTime: ElasticFilterDefaultTime = new ElasticFilterDefaultTime('now-24h', 'now');
  dateFormat$: Observable<DatePipeDefaultOptions>;
  destroy$ = new Subject<void>();
  filterWidth: number;
  tableWidth: number;

  constructor(private indexPatternBehavior: IndexPatternBehavior,
              private logAnalyzerService: ElasticDataService,
              private indexFieldController: IndexFieldController,
              private modalService: NgbModal,
              private queryRunBehavior: QueryRunBehavior,
              private activatedRoute: ActivatedRoute,
              private utmToastService: UtmToastService,
              private utmFilterBehavior: UtmFilterBehavior,
              private elasticDataExportService: ElasticDataExportService,
              private timezoneFormatService: TimezoneFormatService,
              private logFilterBehavior: LogFilterBehavior,
              private router: Router,
              private tabService: TabService) {

    this.detailWidth = (this.pageWidth - 310);
  }

  ngOnInit() {
    this.setInitialWidth();
    this.activatedRoute.queryParams
      .subscribe(params => {
          this.queryParams = Object.entries(params).length > 0 ? params : null;
          if (this.queryParams) {
            if (this.queryParams['@timestamp']) {
              const range = this.queryParams['@timestamp'].split('->')[1].split(',');
              this.defaultTime = new ElasticFilterDefaultTime(range[0], range[1]);
              this.filters[0] = {
                field: '@timestamp',
                operator: ElasticOperatorsEnum.IS_BETWEEN,
                value: [range[0], range[1]]
              };
            }
          }
      });
    this.dateFormat$ = this.timezoneFormatService.getDateFormatSubject();
    this.initExplorer();
  }

  initExplorer() {
    this.dataNature = this.data.pattern.pattern;
    this.pattern = this.data.pattern;
    this.indexPatternBehavior.changePattern({pattern: this.pattern, tabUUID: this.uuid});
    this.logFilterBehavior.$logFilter.next({filter: this.filters, sort: this.sortBy});

    this.resolveParams().then((dataNature) => {
      this.indexPatternBehavior.pattern$
        .pipe(
          takeUntil(this.destroy$),
          filter((nature) => !!nature && this.uuid === nature.tabUUID)
        )
        .subscribe(nature => {
          this.setFilterSearchOnNatureChange();
          this.pattern = nature.pattern;
          if (!this.data) {
            this.selectedFields = [{name: '@timestamp', type: ElasticDataTypesEnum.DATE}];
          }
          this.rows = [];
          this.loading = true;
          setTimeout(() => {
            this.getData();
          }, 1000);
        });
    });
  }


  /**
   * Resolve params and data nature depending on action
   */
  resolveParams(): Promise<any> {
    return new Promise<any>(resolve => {
      let origin: any = DataNatureTypeEnum.ALERT;
      // If query params exist
      if (this.queryParams) {
        origin = this.dataNature;
        // get filters from url and add to current filter
        parseQueryParamsToFilter(this.queryParams).then((filters) => {
          filters.forEach(filterType => {
            const indexFilter = this.filters.findIndex(value => value.field === filterType.field);
            if (indexFilter !== -1) {
              this.filters[indexFilter] = filterType;
            } else {
              if (filterType.field !== 'refreshRoute') {
                this.filters.push(filterType);
              }
            }
          });
          // this.filters = this.filters.concat(filters);
        });
      }
      // If does not have query params and data assign data nature ALERT as default
      if (!this.data && !this.queryParams) {
        origin = DataNatureTypeEnum.EVENT;
      }
      // If data exist assign current and selected filter to data input
      if (this.data) {
        this.selectedFields = [];
        this.fields = this.data.columnsType;
        this.filters = this.data.filtersType ? this.data.filtersType : this.filters;
        origin = this.data.pattern.pattern;
        this.pattern = this.data.pattern;
        if (this.fields) {
          this.fields.forEach(value => {
            this.selectedFields.push({name: value.field, type: value.type});
          });
        } else {
          this.selectedFields.push(
            {name: '@timestamp', type: ElasticDataTypesEnum.DATE},
          );
        }
      }

      resolve(origin + (new Date().getTime()));
    });
  }

  onPageChange($event: number) {
    this.page = $event;
    this.getData();
  }

  onSizeChange($event: number) {
    this.itemsPerPage = $event;
    this.page = 1;
    this.getData();
  }

  onColumnChange($event: ElasticSearchFieldInfoType[]) {
    this.fields = [];
    $event.forEach(value => {
      this.fields.push({field: value.name, visible: true, type: value.type});
    });
  }

  onFilterChange($event: ElasticFilterType[]) {
    this.utmFilterBehavior.$filterChange.next(null);
    this.utmFilterBehavior.$filterExistChange.next(null);
    this.filters = $event;
    this.logFilterBehavior.$logFilter.next({filter: this.filters, sort: this.sortBy});
    this.page = 1;
    this.getData();
  }

  getData() {
    const dateStart = new Date();
    if (!this.runningQuery) {
      this.runningQuery = true;
      this.loading = true;
      this.logAnalyzerService.search(
        this.page,
        this.itemsPerPage,
        LOG_ANALYZER_TOTAL_ITEMS,
        this.pattern.pattern,
        this.filters,
        this.sortBy).subscribe(
        (res: HttpResponse<any>) => {
          this.counter = moment(new Date()).diff(dateStart, 'seconds', true);
          this.totalItems = Number(res.headers.get('X-Total-Count'));
          this.rows = res.body;
          this.loading = false;
          this.error = false;
          this.runningQuery = false;
          this.queryRunBehavior.$runQueryFinished.next(true);
        },
        (res: HttpResponse<any>) => {
          this.counter = moment(new Date()).diff(dateStart, 'seconds', true);
          this.loading = false;
          this.runningQuery = false;
          this.error = true;
          this.rows = [];
        }
      );
    } else {
      this.utmToastService.showInfo('Processing', 'Query still running,please wait to finish');
    }
  }

  onSortBy($event: string) {
    this.sortBy = $event;
    this.logFilterBehavior.$logFilter.next({filter: this.filters, sort: this.sortBy});
    this.getData();
  }

  onRemoveColumn($event: UtmFieldType) {
    this.indexFieldController.$field.next($event.field);
  }

  setInitialWidth() {
    if (this.pageWidth > 4000) {
      this.filterWidth = 400;
    } else if (this.pageWidth > 2500) {
      this.filterWidth = 350;
    } else if (this.pageWidth > 1980) {
      this.filterWidth = 350;
    } else {
      this.filterWidth = 300;
    }

    this.tableWidth = this.pageWidth - this.filterWidth - 51;
  }

  changeFields(pattern: UtmIndexPattern) {
    /*this.utmFilterBehavior.$filterChange.next(null);
    this.utmFilterBehavior.$filterExistChange.next(null);
    this.filters = this.filters.filter(value => value.operator === ElasticOperatorsEnum.IS_IN_FIELD || value.field === '@timestamp');
    //this.indexPatternBehavior.changePattern({pattern: this.pattern, tabUUID: this.uuid});
    this.selectedFields = [{name: '@timestamp', type: ElasticDataTypesEnum.DATE}];
    this.sortBy = NatureDataPrefixEnum.TIMESTAMP + ',' + 'desc';
    this.pattern = pattern;
    this.fields = [{field: '@timestamp', type: ElasticDataTypesEnum.DATE, visible: true, label: '@timestamp'}];*/

    this.router.navigate([], {
      relativeTo: this.activatedRoute,
      queryParams: {
        patternId: pattern.id,
        indexPattern: pattern.pattern,
        refreshRoute: true
      },
    });
  }

  saveQuery() {
    const modal = this.modalService.open(LogAnalyzerQueryCreateComponent, {centered: true});
    modal.componentInstance.columns = this.fields;
    modal.componentInstance.query = this.data;
    modal.componentInstance.filters = this.filters;
    modal.componentInstance.pattern = this.pattern;
  }

  exportToCsv() {
    this.csvExport = true;
    const params = {
      columns: this.fields,
      indexPattern: this.pattern.pattern,
      filters: this.filters,
      top: LOG_ANALYZER_TOTAL_ITEMS
    };
    this.elasticDataExportService.exportCsv(params, 'UTM LOG EXPLORER').then(() => {
      this.csvExport = false;
    });
  }

  onSearchInAll($event: string) {
    const indexValue = this.filters.findIndex(value => value.operator === ElasticOperatorsEnum.IS_IN_FIELD);
    if (indexValue !== -1) {
      if ($event || $event !== '') {
        this.filters[indexValue].value = $event;
      } else {
        this.filters.splice(indexValue, 1);
      }
    } else {
      const filter: ElasticFilterType = {
        field: this.resolvePrefix(),
        operator: ElasticOperatorsEnum.IS_IN_FIELD,
        value: $event
      };
      this.filters.push(filter);
    }
    this.logFilterBehavior.$logFilter.next({filter: this.filters, sort: this.sortBy});
    this.getData();
  }

  /**
   * Set prefix is data nature change
   */
  setFilterSearchOnNatureChange() {
    const indexFieldIn = this.filters.findIndex(value => value.operator === ElasticOperatorsEnum.IS_IN_FIELD);
    if (indexFieldIn !== -1) {
      this.filters[indexFieldIn].field = this.resolvePrefix();
    }
  }

  resolvePrefix(): string {
    switch (this.dataNature) {
      case DataNatureTypeEnum.ALERT:
        return NatureDataPrefixEnum.ALERT + '*';
      case DataNatureTypeEnum.EVENT:
        return NatureDataPrefixEnum.EVENT + '*';
      case DataNatureTypeEnum.VULNERABILITY:
        return NatureDataPrefixEnum.VULNERABILITY + '*';
    }
  }

  onResize($event: ResizeEvent) {
    if ($event.rectangle.width >= 250) {
      this.detailWidth = (this.pageWidth - $event.rectangle.width - 10);
      this.fieldWidth = $event.rectangle.width + 'px';
    }
  }

  @HostListener('window:resize', ['$event'])
  onResizeWindows(event: any) {
    this.pageWidth = event.target.innerWidth;
    this.setInitialWidth();
  }

  loadQuery() {
    this.router.navigate(['/discover/log-analyzer-queries']);
  }

  ngOnDestroy(): void {
    this.filters = [];
    this.utmFilterBehavior.$filterChange.next(null);
    this.utmFilterBehavior.$filterExistChange.next(null);
    this.destroy$.next();
    this.destroy$.complete();
  }
}
