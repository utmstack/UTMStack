import {Component, Input, OnInit} from '@angular/core';
import {Observable} from 'rxjs';
import {TimezoneFormatService} from '../../../shared/services/utm-timezone.service';
import {DatePipeDefaultOptions} from '../../../shared/types/date-pipe-default-options';
import {ComplianceStatusExtendedEnum} from '../../shared/enums/compliance-status.enum';
import {ComplianceControlEvaluationType} from '../../shared/type/compliance-control-evaluation.type';


@Component({
  selector: 'app-compliance-evaluation-view-detail',
  templateUrl: './compliance-evaluation-view-detail.component.html',
  styleUrls: ['./compliance-evaluation-view-detail.component.css']
})
export class ComplianceEvaluationViewDetailComponent implements OnInit {
  @Input() control: ComplianceControlEvaluationType;
  dateFormat$: Observable<DatePipeDefaultOptions>;
  ComplianceStatusExtendedEnum = ComplianceStatusExtendedEnum;
  constructor(private timezoneFormatService: TimezoneFormatService) { }

  ngOnInit() {
    this.dateFormat$ = this.timezoneFormatService.getDateFormatSubject();
  }
}
