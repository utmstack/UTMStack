import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {AlertReportType} from '../../../../../../shared/types/alert/alert-report.type';

@Component({
  selector: 'app-alert-report-filter',
  templateUrl: './alert-report-filter.component.html',
  styleUrls: ['./alert-report-filter.component.scss']
})
export class AlertReportFilterComponent implements OnInit {
  @Input() report: AlertReportType;

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }

}
