import {Component, Input, OnInit} from '@angular/core';
import * as moment from 'moment/moment';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable} from 'rxjs';
import {concatMap, filter, map, tap} from 'rxjs/operators';
import {UtmRenderVisualization} from '../../../../dashboard/shared/services/utm-render-visualization.service';
import {RunVisualizationService} from '../../../../graphic-builder/shared/services/run-visualization.service';
import {VisualizationType} from '../../../../shared/chart/types/visualization.type';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {TimeWindowsService} from '../../../shared/components/utm-cp-section/time-windows.service';
import {ComplianceReportType} from '../../../shared/type/compliance-report.type';
import {ExportPdfService} from '../../../../shared/services/util/export-pdf.service';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ComplianceStatusEnum} from '../../../shared/enums/compliance-status.enum';
import {UtmDashboardVisualizationType} from '../../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';

@Component({
  selector: 'app-compliance-report-detail',
  templateUrl: './compliance-report-detail.component.html',
  styleUrls: ['./compliance-report-detail.component.css']
})
export class ComplianceReportDetailComponent implements OnInit {
  request = {
    page: 0,
    size: 10,
    sort: 'order,asc',
    'idDashboard.equals': 0
  };
  _report: ComplianceReportType;
  compliance$!: Observable<any>;
  csvExport = false;
  printFormat = false;
  vis: VisualizationType;
  ComplianceStatusEnum = ComplianceStatusEnum;

  constructor(private utmRenderVisualization: UtmRenderVisualization,
              private runVisualization: RunVisualizationService,
              private timeWindowsService: TimeWindowsService,
              private spinner: NgxSpinnerService,
              private exportPdfService: ExportPdfService,
              private toastService: UtmToastService) { }

  ngOnInit() {
    this.compliance$ = this.utmRenderVisualization.onRefresh$
      .pipe(
        filter((refresh) => !!refresh),
        concatMap(() => this.runVisualization.run(this.vis, {
          page: 0,
          size: 5,
        })),
        map(run => run[0] && run[0] ? run[0] : {
          rows: []
        }),
      );
  }

  @Input() set report(report: ComplianceReportType) {
    if (report) {
      this._report = report;
      const visualizationType: UtmDashboardVisualizationType = report.dashboard.find(vis =>
        vis.visualization.chartType === ChartTypeEnum.TABLE_CHART || vis.visualization.chartType === ChartTypeEnum.LIST_CHART);

      if (visualizationType) {
        this.vis = visualizationType.visualization;
      }
      const time = visualizationType.visualization.filterType.find(filterType => filterType.field === '@timestamp');
      if (time) {
        this.timeWindowsService.changeTimeWindows({
          reportId: report.id,
          time: time.value[0]
        });
      }

      this.utmRenderVisualization.notifyRefresh(true);
    }
  }

  get report() {
    return this._report;
  }

  exportToCsv(compliance: any) {
    this.csvExport = true;
    const columns = compliance.columns.map(column => column.split('->')[1]);
    const rows = compliance.rows.map(row => row.map(cell => cell.value));
    const csvData = [
      ['Report Name:', this.report.associatedDashboard.name || 'No name'],
      ['Description:', this.report.associatedDashboard.description || 'No description'],
      [],
      columns,
      ...rows
    ];

    const csvContent = this.convertToCsv(csvData);

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', 'UTM_COMPLIANCE_DETAILS-' + moment(new Date()).format('YYYY-MM-DD') + '.csv');
    link.style.visibility = 'hidden';

    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);

    this.csvExport = false;
  }

  convertToCsv(data: any[][]): string {
    return data
      .map(row => row.map(item => {
        if (Array.isArray(item)) {
          return `"${item.join(', ').replace(/"/g, '""')}"`;
        }

        return `"${String(item).replace(/"/g, '""')}"`;
      }).join(','))
      .join('\n');
  }

  exportToPdf() {
    this.spinner.show('buildPrintPDF');
    const url = '/dashboard/export-compliance/' + this.report.id;
    const fileName = this.report.associatedDashboard.name.replace(/ /g, '_');
    this.exportPdfService.getPdf(url, fileName, 'PDF_TYPE_TOKEN').subscribe(response => {
      this.spinner.hide('buildPrintPDF').then(() =>
        this.exportPdfService.handlePdfResponse(response));
    }, error => {
      this.spinner.hide('buildPrintPDF').then(() =>
        this.toastService.showError('Error', 'An error occurred while creating a PDF.'));
    });
  }

  isComplaint() {
    return this.report.configReportStatus === ComplianceStatusEnum.COMPLIANT
      || (this.report.configReportNote && this.report.configReportNote !== '');
  }
}
