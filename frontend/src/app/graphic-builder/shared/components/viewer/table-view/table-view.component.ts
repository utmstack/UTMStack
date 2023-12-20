import {
  AfterViewInit,
  ChangeDetectorRef,
  Component,
  EventEmitter,
  Input,
  OnChanges,
  OnInit,
  Output,
  QueryList,
  SimpleChanges,
  ViewChildren
} from '@angular/core';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {DashboardBehavior} from '../../../../../shared/behaviors/dashboard.behavior';
import {EchartClickAction} from '../../../../../shared/chart/types/action/echart-click-action';
import {UtmTableOptionType} from '../../../../../shared/chart/types/charts/table/utm-table-option.type';
import {TableBuilderResponseType} from '../../../../../shared/chart/types/response/table-builder-response.type';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {ElasticFilterDefaultTime} from '../../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {SortableDirective} from '../../../../../shared/directives/sortable/sortable.directive';
import {SortDirection} from '../../../../../shared/directives/sortable/type/sort-direction.type';
import {SortEvent} from '../../../../../shared/directives/sortable/type/sort-event';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {ChartValueSeparator} from '../../../../../shared/enums/chart-value-separator';
import {TimeFilterType} from '../../../../../shared/types/time-filter.type';
import {mergeParams, sanitizeFilters} from '../../../../../shared/util/elastic-filter.util';
import {RunVisualizationBehavior} from '../../../behavior/run-visualization.behavior';
import {RunVisualizationService} from '../../../services/run-visualization.service';
import {UtmChartClickActionService} from '../../../services/utm-chart-click-action.service';
import {rebuildVisualizationFilterTime} from '../../../util/chart-filter/chart-filter.util';
import {resolveDefaultVisualizationTime} from '../../../util/visualization/visualization-render.util';

@Component({
  selector: 'app-table-view',
  templateUrl: './table-view.component.html',
  styleUrls: ['./table-view.component.scss']
})
export class TableViewComponent implements OnInit, AfterViewInit, OnChanges {
  @ViewChildren(SortableDirective) headers: QueryList<SortableDirective>;
  @Input() visualization: VisualizationType;
  @Input() building: boolean;
  /**
   * Define is table will export or not to remove fixing height and items per page
   */
  @Input() exportFormat: boolean;
  @Output() runned = new EventEmitter<string>();
  @Input() loadingOption = true;
  @Input() chart: ChartTypeEnum;
  @Input() width: string;
  @Input() height: string;
  @Input() showTime: boolean;
  @Input() timeByDefault: any;
  @Input() chartId: number;
  totalItems: number;
  page = 1;
  itemsPerPage = 5;
  pageStart = 0;
  pageEnd = 5;
  data: TableBuilderResponseType;
  direction: SortDirection = '';
  tableOptions: UtmTableOptionType;
  mapMetric: Map<number, number> = new Map<number, number>();
  chartTypeEnum = ChartTypeEnum;
  responseRows: Array<{ value: any; metric: boolean }[]>;
  exporting = false;
  error = false;
  defaultTime: ElasticFilterDefaultTime;
  screenResolution = window.screen.width;
  searching = false;
  empty = false;

  constructor(private runVisualizationService: RunVisualizationService,
              private utmChartClickActionService: UtmChartClickActionService,
              private runVisualizationBehavior: RunVisualizationBehavior,
              private cdr: ChangeDetectorRef,
              private dashboardBehavior: DashboardBehavior,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.runVisualization();
    this.defaultTime = resolveDefaultVisualizationTime(this.visualization);
    this.runVisualizationBehavior.$run.subscribe(id => {
      if (id && this.chartId === id) {
        this.runVisualization();
        this.defaultTime = resolveDefaultVisualizationTime(this.visualization);
      }
    });
    this.dashboardBehavior.$filterDashboard.subscribe(dashboardFilter => {
      if (dashboardFilter && dashboardFilter.indexPattern === this.visualization.pattern.pattern) {
        mergeParams(dashboardFilter.filter, this.visualization.filterType).then(newFilters => {
          this.visualization.filterType = sanitizeFilters(newFilters);
          this.runVisualization();
        });
      }
    });
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes.height && changes.height.currentValue !== 'NaNpx') {
      const currentGridHeight = Number(changes.height.currentValue.replace('px', ''));
      if (!isNaN(currentGridHeight)) {
        this.itemsPerPage = this.calcItemsPerPage();
        this.pageEnd = this.page * this.itemsPerPage;
        this.pageStart = this.pageEnd - this.itemsPerPage;
      }
    }
  }

  ngAfterViewInit(): void {
    this.cdr.detectChanges();
  }

  calcItemsPerPage(): number {
    this.tableOptions = this.visualization.chartConfig;
    if (this.exportFormat) {
      return this.data.rows.length - 1;
    } else if (this.tableOptions.dynamicPageSize) {
      const gridHeight = Number(this.height.replace('px', ''));
      const calc = Math.round((gridHeight / 40));
      return gridHeight < 390 ? (calc - 3) : (calc - 2);
    } else {
      return this.tableOptions.itemsPerPage;
    }
  }

  runVisualization() {
    this.loadingOption = true;
    this.runVisualizationService.run(this.visualization).subscribe(data => {
      this.empty = data.length === 0 || data[0].rows.length === 0;
      this.data = data.length > 0 ? data[0] : [];
      if (!this.empty) {
        this.tableOptions = this.visualization.chartConfig;
        this.itemsPerPage = this.itemsPerPage = this.calcItemsPerPage();
        this.pageStart = 0;
        this.pageEnd = this.itemsPerPage;
        this.responseRows = this.data.rows;
        this.totalItems = this.data.rows.length;
        this.mapMetric.clear();
        if (this.tableOptions && this.tableOptions.totalFunction) {
          this.applyFunction();
        }
      }
      this.loadingOption = false;
      this.error = false;
      this.runned.emit('runned');
    }, error => {
      this.loadingOption = false;
      this.error = true;
      this.runned.emit('runned');
      this.toastService.showError('Error',
        'Error occurred while running visualization');
    });
  }

  loadPage(page: number) {
    this.pageEnd = page * this.itemsPerPage;
    this.pageStart = this.pageEnd - this.itemsPerPage;
  }

  onSort({column, direction}: SortEvent) {
    this.direction = direction;
    this.headers.forEach(header => {
      if (header.sortable !== column) {
        header.direction = '';
      }
    });
    if (direction === '') {
      this.data.rows = this.responseRows;
    } else {
      this.data.rows = this.data.rows.sort((a, b) => {
        return a[1][column] > b[1][column] ? 1 : -1;
      });
    }
  }

  getFirstMetric(): number {
    return this.data.rows[0].findIndex(value => value.metric === true);
  }

  applyFunction() {
    switch (this.tableOptions.totalFunction) {
      case 'sum':
        this.applySum();
        break;
      case 'count':
        this.applyCount();
        break;
      case 'min':
        this.applyMin();
        break;
      case 'max':
        this.applyMax();
        break;
      case 'avg':
        this.applyAvg();
        break;


    }
  }

  applySum() {
    for (const row of this.data.rows) {
      for (let i = 0; i < row.length; i++) {
        if (row[i].metric) {
          if (this.mapMetric.get(i)) {
            const val = this.mapMetric.get(i) + row[i].value;
            this.mapMetric.set(i, val);
          } else {
            this.mapMetric.set(i, row[i].value);
          }
        }
      }
    }
  }

  applyCount() {
    for (const row of this.data.rows) {
      for (let i = 0; i < row.length; i++) {
        if (row[i].metric) {
          this.mapMetric.set(i, this.data.rows.length);
        }
      }
    }
  }

  applyMin() {
    for (const row of this.data.rows) {
      for (let i = 0; i < row.length; i++) {
        if (row[i].metric) {
          if (this.mapMetric.get(i)) {
            const val = this.mapMetric.get(i) > row[i].value ? row[i].value : this.mapMetric.get(i);
            this.mapMetric.set(i, val);
          } else {
            this.mapMetric.set(i, row[i].value);
          }
        }
      }
    }
  }

  applyMax() {
    for (const row of this.data.rows) {
      for (let i = 0; i < row.length; i++) {
        if (row[i].metric) {
          if (this.mapMetric.get(i)) {
            const val = this.mapMetric.get(i) < row[i].value ? row[i].value : this.mapMetric.get(i);
            this.mapMetric.set(i, val);
          } else {
            this.mapMetric.set(i, row[i].value);
          }
        }
      }
    }
  }

  applyAvg() {
    for (const row of this.data.rows) {
      for (let i = 0; i < row.length; i++) {
        if (row[i].metric) {
          if (this.mapMetric.get(i)) {
            const val = Number(((this.mapMetric.get(i) + row[i].value) / this.data.rows.length).toFixed(2));
            this.mapMetric.set(i, val);
          } else {
            this.mapMetric.set(i, row[i].value);
          }
        }
      }
    }
  }

  getMapMetric(): number[] {
    const values: number[] = [];
    this.mapMetric.forEach((value, key) => values.push(value));
    return values;
  }

  onTimeFilterChange($event: TimeFilterType) {
    rebuildVisualizationFilterTime($event, this.visualization.filterType).then(filters => {
      this.visualization.filterType = filters;
      this.runVisualization();
    });
  }

  chartEvent($event: { value: any, metric: boolean }[]) {
    if (typeof this.visualization.chartAction === 'string') {
      this.visualization.chartAction = JSON.parse(this.visualization.chartAction);
    }
    const echartClickAction: EchartClickAction = {
      seriesName: this.buildSeriesNameFromRow($event)
    };
    if (!this.building) {
      this.utmChartClickActionService.onClickNavigate(this.visualization, echartClickAction);
    }
  }

  buildSeriesNameFromRow(rowSelected: { value: any, metric: boolean }[]): string {
    let str = '';
    // tslint:disable-next-line:prefer-for-of
    for (let i = 0; i < rowSelected.length; i++) {
      if (!rowSelected[i].metric) {
        str += this.data.columns[i].split('->')[0] +
          ChartValueSeparator.VALUE_SEPARATOR +
          rowSelected[i].value + ChartValueSeparator.BUCKET_SEPARATOR;
      }
    }

    if (str.endsWith(ChartValueSeparator.BUCKET_SEPARATOR.toString())) {
      str = str.substring(0, (str.length - ChartValueSeparator.BUCKET_SEPARATOR.toString().length));
    }
    return str;
  }

  exportToCSV() {
    this.exporting = true;
    this.processToCsv().then(rows => {
      setTimeout(() => {
        const csvContent = 'data:text/csv;charset=utf-8,'
          + rows.map(e => e.join(',')).join('\n');
        const encodedUri = encodeURI(csvContent);
        const link = document.createElement('a');
        link.setAttribute('href', encodedUri);
        link.setAttribute('download', this.processVisNameToCsvName() + '.csv');
        document.body.appendChild(link); // Required for FF
        link.click(); // This will download the data file named "my_data.csv".
        this.exporting = false;
      }, 3000);
    });
  }


  /**
   * Return clean name for csv based on visualization name
   */
  processVisNameToCsvName(): string {
    let str = this.visualization.name;
    str = str.replace(/\W+(?!$)/g, '-').toLowerCase();
    str = str.replace(/\W$/, '').toLowerCase();
    return str;
  }

  /**
   * Process current data to csv
   */
  processToCsv(): Promise<Array<any[]>> {
    return new Promise<Array<any[]>>(resolve => {
      const dataExport: Array<any[]> = [];
      // First extract columns to set the first element of exported array to csv, this
      // way get csv headers
      this.addColumnsToCsv().then(csvHeader => {
        dataExport.push(csvHeader);
        // after header as been added process row and extract data
        this.addRowsToCsv().then(csvRows => {
          csvRows.forEach(row => {
            dataExport.push(row);
          });
        });
      });
      resolve(dataExport);
    });
  }

  /**
   * Process row to convert to acceptable data to export to csv
   */
  addRowsToCsv(): Promise<any[]> {
    return new Promise<any[]>(resolve => {
      const rows: any[] = [];
      this.data.rows.forEach((rowsData) => {
        this.convertRowTableTypeToStringArray(rowsData).then(data => {
          rows.push(data);
        });
      });
      resolve(rows);
    });
  }

  /**
   * Extract data value for each row
   * @param rowsData Row object of TableBuilderResponseType row property
   */
  convertRowTableTypeToStringArray(rowsData: { value: any, metric: boolean }[]): Promise<string[]> {
    return new Promise<string[]>(resolve => {
      const dataArr: string[] = [];
      rowsData.forEach(row => {
        this.addQuoteToData(row.value).then(value => {
          dataArr.push(value);
        });
      });
      resolve(dataArr);
    });
  }

  addQuoteToData(value): Promise<any> {
    return new Promise<any>(resolve => {
      if (typeof value === 'string' || typeof value === 'object') {
        if (isNaN(Date.parse(value))) {
          value = '\"' + value + '\"';
        }
      }
      resolve(value);
    });
  }

  /**
   * Extract table header and covert to acceptable format to export
   */
  addColumnsToCsv(): Promise<string[]> {
    return new Promise<string[]>(resolve => {
      const columns: string[] = [];
      this.data.columns.forEach((col) => {
        const columnName = col.split('->')[1];
        columns.push((columnName === '' || columnName === null || columnName === undefined) ? '\"\"' : columnName);
      });
      resolve(columns);
    });
  }

}
