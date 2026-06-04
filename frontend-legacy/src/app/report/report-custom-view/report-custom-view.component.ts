import {Component, OnInit, QueryList, ViewChildren} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {TableBuilderResponseType} from '../../shared/chart/types/response/table-builder-response.type';
import {SortableDirective} from '../../shared/directives/sortable/sortable.directive';
import {SortDirection} from '../../shared/directives/sortable/type/sort-direction.type';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {ChartTypeEnum} from '../../shared/enums/chart-type.enum';
import {ReportParamsEnum} from '../shared/enums/report-params.enum';
import {ReportService} from '../shared/service/report.service';
import {ReportType} from '../shared/type/report.type';

@Component({
  selector: 'app-report-custom-view',
  templateUrl: './report-custom-view.component.html',
  styleUrls: ['./report-custom-view.component.css']
})
export class ReportCustomViewComponent implements OnInit {
  printFormat = false;
  loading = true;
  reportConfig: ReportType;
  reportId: number;
  standardId: number;
  sectionId: number;
  data: TableBuilderResponseType;
  direction: SortDirection = '';
  @ViewChildren(SortableDirective) headers: QueryList<SortableDirective>;
  chartTypeEnum = ChartTypeEnum;
  private responseRows: Array<{ value: any; metric: boolean }[]>;

  constructor(private activeRoute: ActivatedRoute,
              private reportService: ReportService) {

    this.activeRoute.queryParams.subscribe((params) => {
      this.reportId = params[ReportParamsEnum.TEMPLATE_ID];
      this.sectionId = params[ReportParamsEnum.SECTION_ID];
    });
  }


  ngOnInit() {
    window.addEventListener('beforeprint', (event) => {
      this.printFormat = true;
    });
    window.addEventListener('afterprint', (event) => {
      this.printFormat = false;
    });
    this.getTemplate();
  }

  getTemplate() {
    this.reportService.find(this.reportId).subscribe(response => {
      this.reportConfig = response.body;
      this.getReportData(this.reportConfig.repUrl);
    });
  }


  print() {
    this.printFormat = true;
    setTimeout(() => {
      window.print();
    }, 2000);
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

  private getReportData(endpoint: string) {
    this.reportService.getReportListData(endpoint).subscribe(response => {
      this.loading = false;
      this.data = response.body[0];
      this.responseRows = this.data.rows;
    });
  }

}
