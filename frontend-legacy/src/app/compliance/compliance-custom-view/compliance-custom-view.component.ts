import {Component, OnInit, QueryList, ViewChildren} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {TableBuilderResponseType} from '../../shared/chart/types/response/table-builder-response.type';
import {SortableDirective} from '../../shared/directives/sortable/sortable.directive';
import {SortDirection} from '../../shared/directives/sortable/type/sort-direction.type';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {ChartTypeEnum} from '../../shared/enums/chart-type.enum';
import {ComplianceParamsEnum} from '../shared/enums/compliance-params.enum';
import {CpReportsService} from '../shared/services/cp-reports.service';
import {ComplianceReportType} from '../shared/type/compliance-report.type';

@Component({
  selector: 'app-compliance-custom-view',
  templateUrl: './compliance-custom-view.component.html',
  styleUrls: ['./compliance-custom-view.component.scss']
})
export class ComplianceCustomViewComponent implements OnInit {
  printFormat = false;
  loading = true;
  reportConfig: ComplianceReportType;
  reportId: number;
  standardId: number;
  sectionId: number;
  data: TableBuilderResponseType;
  direction: SortDirection = '';
  @ViewChildren(SortableDirective) headers: QueryList<SortableDirective>;
  chartTypeEnum = ChartTypeEnum;
  private responseRows: Array<{ value: any; metric: boolean }[]>;

  constructor(private activeRoute: ActivatedRoute,
              private cpReportsService: CpReportsService) {

    this.activeRoute.queryParams.subscribe((params) => {
      this.reportId = params[ComplianceParamsEnum.TEMPLATE];
      this.standardId = params[ComplianceParamsEnum.STANDARD_ID];
      this.sectionId = params[ComplianceParamsEnum.SECTION_ID];
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
    this.cpReportsService.find(this.reportId).subscribe(response => {
      this.reportConfig = response.body;
      this.getReportData(this.reportConfig.configUrl);
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
    this.cpReportsService.getCustomReportData(endpoint).subscribe(response => {
      this.loading = false;
      this.data = response.body[0];
      this.responseRows = this.data.rows;
    });
  }
}
