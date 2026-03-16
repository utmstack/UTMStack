import {Component, Input, OnInit} from '@angular/core';
import {Observable} from 'rxjs';
import {TimezoneFormatService} from '../../../shared/services/utm-timezone.service';
import {DatePipeDefaultOptions} from '../../../shared/types/date-pipe-default-options';
import {ComplianceStatusExtendedEnum} from '../../shared/enums/compliance-status.enum';
import {ComplianceControlLatestEvaluationType} from '../../shared/type/compliance-control-latest-evaluation.type';

@Component({
  selector: 'app-compliance-latest-evaluation-view-detail',
  templateUrl: './compliance-latest-evaluation-view-detail.component.html',
  styleUrls: ['./compliance-latest-evaluation-view-detail.component.css']
})
export class ComplianceLatestEvaluationViewDetailComponent implements OnInit {
  @Input() control: ComplianceControlLatestEvaluationType;
  dateFormat$: Observable<DatePipeDefaultOptions>;
  ComplianceStatusExtendedEnum = ComplianceStatusExtendedEnum;
  constructor(private timezoneFormatService: TimezoneFormatService) { }

  ngOnInit() {
    this.dateFormat$ = this.timezoneFormatService.getDateFormatSubject();
  }
}
