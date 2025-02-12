import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {ComplianceStatusEnum} from '../../../../../compliance/shared/enums/compliance-status.enum';
import {CpReportsService} from '../../../../../compliance/shared/services/cp-reports.service';
import {ComplianceReportType} from '../../../../../compliance/shared/type/compliance-report.type';
import {UtmToastService} from '../../../../alert/utm-toast.service';

@Component({
  selector: 'app-modal-add-note',
  templateUrl: './modal-add-note.component.html',
  styleUrls: ['./modal-add-note.component.scss']
})
export class ModalAddNoteComponent implements OnInit {
  @Input() report: ComplianceReportType;
  @Input() isComplaint: boolean;
  header: string;
  message: string;
  confirmBtnText: string;
  confirmBtnIcon: string;
  confirmBtnType: 'delete' | 'default';
  textDisplay: string;
  textType: 'warning' | 'danger';
  hideBtnCancel = false;
  note = '';
  ComplianceStatusEnum = ComplianceStatusEnum;

  constructor(private activeModal: NgbActiveModal,
              private reportsService: CpReportsService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
  }

  confirm() {

    this.reportsService.update({
      ...this.report,
      configReportNote: !this.isComplaint ? this.note : null,
      dashboard: [],
    }).subscribe(response => {
      this.toastService.showSuccessBottom('Note added successfully');
      this.reportsService.notifyRefresh({
        loading: true,
        sectionId: this.report.section.id,
        reportSelected: 0
      });
    }, error => {
      this.toastService.showError('Error adding note',
        'Error adding comment, please try again');
    });
    this.activeModal.close('ok');
  }

  cancel() {
    this.activeModal.dismiss('cancel');
  }
}
