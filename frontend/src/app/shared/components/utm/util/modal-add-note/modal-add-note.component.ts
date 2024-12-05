import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {ComplianceReportType} from '../../../../../compliance/shared/type/compliance-report.type';
import {CpReportsService} from "../../../../../compliance/shared/services/cp-reports.service";
import {UtmToastService} from "../../../../alert/utm-toast.service";

@Component({
  selector: 'app-modal-add-note',
  templateUrl: './modal-add-note.component.html',
  styleUrls: ['./modal-add-note.component.scss']
})
export class ModalAddNoteComponent implements OnInit {
  @Input() report: ComplianceReportType;
  header: string;
  message: string;
  confirmBtnText: string;
  confirmBtnIcon: string;
  confirmBtnType: 'delete' | 'default';
  textDisplay: string;
  textType: 'warning' | 'danger';
  hideBtnCancel = false;
  note = '';

  constructor(private activeModal: NgbActiveModal,
              private reportsService: CpReportsService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
  }

  confirm() {
    //this.creating = true;
    this.setReport();
    this.reportsService.update({
      ...this.report,
    }).subscribe(response => {
      this.toastService.showSuccessBottom('Comment added successfully');
     // this.creating = false;
      this.reportsService.notifyRefresh({
        loading: true,
        sectionId: this.report.section.id,
        reportSelected: 0
      });
    }, error => {
      this.toastService.showError('Error adding note',
        'Error adding comment, please try again');
      // this.creating = false;
    });
    this.activeModal.close('ok');
  }

  cancel() {
    this.activeModal.dismiss('cancel');
  }

  setReport(){
    this.report.status === 'non_complaint' ? this.report.note = this.note : this.report.note = '';
  }
}
