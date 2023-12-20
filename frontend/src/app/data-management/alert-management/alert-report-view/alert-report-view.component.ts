import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import * as moment from 'moment';
import {LOG_ANALYZER_TOTAL_ITEMS} from '../../../shared/constants/log-analyzer.constant';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {DataNatureTypeEnum} from '../../../shared/enums/nature-data.enum';
import {ElasticDataExportService} from '../../../shared/services/elasticsearch/elastic-data-export.service';
import {ElasticDataService} from '../../../shared/services/elasticsearch/elastic-data.service';
import {AlertReportType} from '../../../shared/types/alert/alert-report.type';
import {AlertReportService} from '../alert-reports/shared/services/report.service';

@Component({
  selector: 'app-alert-report-view',
  templateUrl: './alert-report-view.component.html',
  styleUrls: ['./alert-report-view.component.scss']
})
export class AlertReportViewComponent implements OnInit {
  reportId: any;
  printFormat = false;
  report: AlertReportType;
  csvExport: any;
  view: 'table' | 'grid' = 'table';
  data: any[];
  loading = true;
  generateReport: boolean;
  private sortBy: string;

  constructor(private activatedRoute: ActivatedRoute,
              private elasticDataService: ElasticDataService,
              private elasticDataExportService: ElasticDataExportService,
              private reportService: AlertReportService) {
  }

  ngOnInit() {
    this.activatedRoute.params.subscribe(params => {
      this.reportId = params.id;
      if (this.reportId) {
        this.reportService.find(this.reportId).subscribe(report => {
          this.report = report.body;
          this.getAlert(this.report);
        });
      }
    });

    window.addEventListener('beforeprint', (event) => {
      this.printFormat = true;
    });
    window.addEventListener('afterprint', (event) => {
      this.printFormat = false;
    });
  }

  getAlert(report: AlertReportType) {
    this.elasticDataService.search(1, report.repLimit,
      report.repLimit,
      DataNatureTypeEnum.ALERT,
      report.filters, this.sortBy).subscribe(
      (res: HttpResponse<any>) => {
        this.data = res.body;
        this.loading = false;
      },
      (res: HttpResponse<any>) => {
      }
    );
  }

  print() {
    this.printFormat = true;
    setTimeout(() => {
      window.print();
    }, 1000);
  }

  exportToCsv() {
    this.generateReport = true;
    const params = {
      columns: this.report.columns,
      dataOrigin: DataNatureTypeEnum.ALERT,
      filters: this.report.filters,
      top: LOG_ANALYZER_TOTAL_ITEMS
    };
    this.elasticDataExportService.exportCsv(params, 'UTM ALERTS').then(() => {
      this.generateReport = false;
    });
  }

  dateGenerated() {
    return moment(new Date()).format('YYYY-MM-DD HH:MM:ss');
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.getAlert(this.report);
  }
}
