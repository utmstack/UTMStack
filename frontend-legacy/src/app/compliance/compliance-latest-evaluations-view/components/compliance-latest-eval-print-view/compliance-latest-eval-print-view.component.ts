import { HttpErrorResponse } from '@angular/common/http';
import {ChangeDetectionStrategy, Component, Input, OnDestroy, OnInit, Output} from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import {EMPTY, Observable} from 'rxjs';
import {catchError, concatMap, filter, map, tap} from 'rxjs/operators';
import { UtmToastService } from 'src/app/shared/alert/utm-toast.service';
import {TimezoneFormatService} from '../../../../shared/services/utm-timezone.service';
import {DatePipeDefaultOptions} from '../../../../shared/types/date-pipe-default-options';
import {CpControlConfigService} from '../../../shared/services/cp-control-config.service';
import {ComplianceControlLatestEvaluationType} from '../../../shared/type/compliance-control-latest-evaluation.type';


@Component({
  selector: 'app-compliance-latest-eval-print-view',
  templateUrl: './compliance-latest-eval-print-view.component.html',
  styleUrls: ['./compliance-latest-eval-print-view.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ComplianceLatestEvalPrintViewComponent implements OnInit, OnDestroy {
  controls$: Observable<ComplianceControlLatestEvaluationType[]>;
  dateFormat$: Observable<DatePipeDefaultOptions>;

  constructor(private controlsService: CpControlConfigService,
              private toastService: UtmToastService,
              private route: ActivatedRoute,
              private timezoneFormatService: TimezoneFormatService) { }

  ngOnInit() {
    this.dateFormat$ = this.timezoneFormatService.getDateFormatSubject();
    this.controls$ = this.route.queryParams
    .pipe(
      filter((params) => !!params.section),
      map((params) => JSON.parse(decodeURIComponent(params.section))),
      concatMap((params) => this.controlsService.fetchData({
        page: params.page,
        size: params.size,
        sectionId: params.id,
        sort: params.sort,
      })),
      map((res) => {
        return res.body.map((r, index) => {
          return {
            ...r,
          };
        });
      }),
      catchError((err: HttpErrorResponse) => {
        this.toastService.showError('Error',
          'Unable to retrieve the list of reports. Please try again or contact support.');
        return EMPTY;
      }));
  }

  ngOnDestroy(): void {
    throw new Error('Method not implemented.');
  }
}
