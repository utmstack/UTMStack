import {Component, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges, ViewChild} from '@angular/core';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {CpReportsService} from '../../services/cp-reports.service';
import {ComplianceReportType} from '../../type/compliance-report.type';
import {filter} from "rxjs/operators";
import {NgbPopover} from "@ng-bootstrap/ng-bootstrap";

@Component({
  selector: 'app-report-apply-note',
  templateUrl: './report-apply-note.component.html',
  styleUrls: ['./report-apply-note.component.scss']
})
export class ReportApplyNoteComponent implements OnInit, OnChanges {
  @Input() report: ComplianceReportType;
  @Input() showNote: boolean;
  @ViewChild('notePopover') notePopover!: NgbPopover;
  @Output() applyNote = new EventEmitter<string>();
  note: string;
  creating = false;

  constructor(private reportsService: CpReportsService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.reportsService.onLoadNote$
      .pipe(
        filter( report => !!report && this.report.id === report.id ))
      .subscribe(() => this.notePopover.open());
  }

  ngOnChanges(changes: SimpleChanges) {
    this.getNoteValue();
  }

  getNoteValue() {
    if (this.report) {
      this.note = this.report.configReportNote && this.report.configReportNote !== '' ? this.report.configReportNote : '';
    }
  }

  addNote() {
    this.creating = true;
    this.reportsService.update({
      ...this.report,
      dashboard: [],
      configReportNote: this.note
    }).subscribe(response => {
      this.toastService.showSuccessBottom('Comment added successfully');
      this.applyNote.emit('success');
      this.creating = false;
      this.reportsService.notifyRefresh({
        loading: true,
        sectionId: this.report.section.id,
        reportSelected: 0
      });
    }, error => {
      this.toastService.showError('Error adding note',
        'Error adding note, please try again');
      this.creating = false;
    });
  }

  onClick(event: Event) {
    event.stopPropagation();
  }
}
