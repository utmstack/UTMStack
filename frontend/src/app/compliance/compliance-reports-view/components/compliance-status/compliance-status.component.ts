import {ChangeDetectionStrategy, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import { Observable } from 'rxjs';
import {concatMap, filter, map, tap} from 'rxjs/operators';
import { UtmRenderVisualization } from '../../../../dashboard/shared/services/utm-render-visualization.service';
import { RunVisualizationService } from '../../../../graphic-builder/shared/services/run-visualization.service';
import { VisualizationType } from '../../../../shared/chart/types/visualization.type';
import { ChartTypeEnum } from '../../../../shared/enums/chart-type.enum';
import { ComplianceReportType } from '../../../shared/type/compliance-report.type';
import {TimeWindowsService} from '../../../shared/components/utm-cp-section/time-windows.service';

@Component({
  selector: 'app-compliance-status',
  templateUrl: './compliance-status.component.html',
  styleUrls: ['./compliance-status.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ComplianceStatusComponent implements OnInit {
  private _report: ComplianceReportType;
  compliance$!: Observable<{status: boolean}>;
  request = {
    page: 0,
    size: 10000,
    sort: 'order,asc',
    'idDashboard.equals': 0
  };

  @Output() isCompliant = new EventEmitter<boolean>();

  constructor(private utmRenderVisualization: UtmRenderVisualization,
              private runVisualization: RunVisualizationService,
              private timeWindowsService: TimeWindowsService) {}

  ngOnInit() {
    this.compliance$ = this.utmRenderVisualization.onRefresh$
      .pipe(
        filter((refresh) => !!refresh),
        concatMap(() => this.utmRenderVisualization.query(this.request)
          .pipe(
            tap(response => console.log('Response:', response)),
            map(response => response.body.filter(vis =>
              vis.visualization.chartType === ChartTypeEnum.TABLE_CHART || vis.visualization.chartType === ChartTypeEnum.LIST_CHART
            )),
            filter(vis => vis.length > 0),
            map(vis => vis[0].visualization),
            tap( vis => {
              const time = vis.filterType.find( filterType => filterType.field === '@timestamp');
              if (time) {
                this.timeWindowsService.changeTimeWindows({
                  reportId: this.report.id,
                  time: time.value[0]
                });
              }
            }),
            concatMap((vis: VisualizationType) => this.runVisualization.run(vis)),
            map(run => {
              const isCompliant = run[0] && run[0].rows.length > 0;
              this.isCompliant.emit(isCompliant);
              return {
                status: isCompliant
              };
            })
          ))
      );
  }

  @Input() set report(value: ComplianceReportType) {
    if (value) {
      this._report = value;
      this.request = {
        ...this.request,
        'idDashboard.equals': value.dashboardId,
      };
      this.utmRenderVisualization.notifyRefresh(true);
    }
  }

  get report() {
    return this._report;
  }
}
