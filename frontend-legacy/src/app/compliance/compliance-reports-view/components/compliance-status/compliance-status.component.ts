import {ChangeDetectionStrategy, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {ModalService} from '../../../../core/modal/modal.service';
import {UtmDashboardVisualizationType} from '../../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {VisualizationType} from '../../../../shared/chart/types/visualization.type';
import {ModalAddNoteComponent} from '../../../../shared/components/utm/util/modal-add-note/modal-add-note.component';
import {ChartTypeEnum} from '../../../../shared/enums/chart-type.enum';
import {TimeWindowsService} from '../../../shared/components/utm-cp-section/time-windows.service';
import {ComplianceStatusEnum} from '../../../shared/enums/compliance-status.enum';
import {ComplianceReportType} from '../../../shared/type/compliance-report.type';


export type ComplianceStatus = 'complaint' | 'non_complaint';

@Component({
  selector: 'app-compliance-status',
  templateUrl: './compliance-status.component.html',
  styleUrls: ['./compliance-status.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ComplianceStatusComponent implements OnInit {
  @Input() template: 'default' | 'dropdown' = 'default';
  private _report: ComplianceReportType;
  vis: VisualizationType;

  @Output() visualization = new EventEmitter<any>();
  changing: any;
  status: ComplianceStatus = 'complaint';
  loading = false;
  ComplianceStatus = ComplianceStatusEnum;

  constructor(private timeWindowsService: TimeWindowsService,
              private modalService: ModalService) {
  }

  ngOnInit() {}

  @Input() set report(value: ComplianceReportType) {
    if (value) {
      this._report = value;
      const visualizationType: UtmDashboardVisualizationType = value.dashboard.find(vis =>
        vis.visualization.chartType === ChartTypeEnum.TABLE_CHART || vis.visualization.chartType === ChartTypeEnum.LIST_CHART);

      if (visualizationType) {
        this.visualization.emit(visualizationType);
        this.vis = visualizationType.visualization;
      }
      const time = visualizationType.visualization.filterType.find(filterType => filterType.field === '@timestamp');
      if (time) {
        this.timeWindowsService.changeTimeWindows({
          reportId: value.id,
          time: time.value[0]
        });
      }
    }
  }

  get report() {
    return this._report;
  }

  onClick(event: Event) {
    event.stopPropagation();
  }

  changeStatusTo(status: ComplianceStatusEnum) {
    const modalRef = this.modalService.open(ModalAddNoteComponent, {centered: true});
    modalRef.componentInstance.report = this.report;
    modalRef.componentInstance.isComplaint = this.isComplaint();
    modalRef.componentInstance.header = 'Confirm status change';
    modalRef.componentInstance.message = this.getModalMessage(status);
    modalRef.componentInstance.confirmBtnText = 'Confirm';
    modalRef.componentInstance.confirmBtnIcon = 'icon-checkmark-circle';
    modalRef.componentInstance.confirmBtnType = 'confirm';
    modalRef.componentInstance.cancelBtnText = 'Cancel';
    modalRef.result.then(() => {
      // this.reportService.loadReportNote(this.report);
    });
  }

  getModalMessage(status: ComplianceStatusEnum) {
    return !this.isComplaint()
      ? `You are about to change the compliance status to <b>Compliant (External Tool)</b>.
       <br><br>
       Please note that you must provide a detailed note explaining where and how this compliance
       is being fulfilled using the external tool.
       <br><br>
       Are you sure you want to proceed?`
      : this.isComplaint()
        ? `You are about to change the compliance status to <b>Non Compliant</b>.
       <br><br>
       Please note that the note associated with this compliance will be permanently removed.
       <br><br>
       Are you sure you want to proceed?`
        : '';
  }

  isComplaint() {
    return this.report.configReportStatus === ComplianceStatusEnum.COMPLIANT
      || (this.report.configReportNote && this.report.configReportNote !== '');
  }
}
