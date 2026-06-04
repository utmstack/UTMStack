import {HttpErrorResponse} from '@angular/common/http';
import {
  ChangeDetectionStrategy,
  Component,
  EventEmitter,
  Input,
  OnChanges,
  OnInit,
  Output,
  SimpleChanges
} from '@angular/core';
import {EMPTY, Observable} from 'rxjs';
import {catchError, concatMap, filter, map, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {CpControlConfigService} from '../../services/cp-control-config.service';
import {ComplianceControlType} from '../../type/compliance-control.type';
import {ComplianceStandardSectionType} from '../../type/compliance-standard-section.type';

@Component({
  selector: 'app-utm-cp-section-config',
  templateUrl: './utm-cp-section-config.component.html',
  styleUrls: ['./utm-cp-section-config.component.css', ],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class UtmCpSectionConfigComponent implements OnInit, OnChanges {

  @Input() section: ComplianceStandardSectionType;
  @Input() index: number;
  @Output() isActive: EventEmitter<number> = new EventEmitter<number>();
  @Input() loadFirst = false;
  @Input() expandable = true;
  @Input() action: 'reports' | 'compliance' = 'compliance';

  controls$: Observable<ComplianceControlType[]>;
  selected: number;

  constructor(private cpControlConfigService: CpControlConfigService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.controls$ = this.cpControlConfigService.onRefresh$
      .pipe(filter(reportRefresh =>
          !!reportRefresh && reportRefresh.loading && reportRefresh.sectionId === this.section.id && this.expandable),
        tap((reportRefresh) => {
          this.selected = reportRefresh.reportSelected;
        }),
        concatMap(() => this.cpControlConfigService.fetchData({
          page: 0,
          size: 1000,
          standardId: this.section.standardId,
          sectionId: this.section.id,
        })),
        map((res) => {
          return res.body.map((r, index) => {
            return {
              ...r,
              selected: index === this.selected
            };
          });
        }),
        tap((controls) => {
          if (this.loadFirst) {
            this.loadReport(controls[0]);
            this.loadFirst = false;
          }
        }),
        catchError((err: HttpErrorResponse) => {
          this.toastService.showError('Error',
            'Unable to retrieve the list of reports. Please try again or contact support.');
          return EMPTY;
        }));
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (this.section.isActive) {
      this.cpControlConfigService.notifyRefresh({
        sectionId: this.section.id,
        loading: true,
        reportSelected: 0
      });
    }
  }

  loadControls() {
    if (!this.section.isActive) {
      this.isActive.emit(this.index);
    } else {
      this.section.isCollapsed = !this.section.isCollapsed;
    }
  }

  generateReport(control: ComplianceControlType, controls: ComplianceControlType[]) {
    if (this.section.isActive && control) {
      controls.forEach(r => r.selected = false);
      control.selected = true;
      this.loadReport(control);
    }
  }

  loadReport(report: ComplianceControlType) {
    this.cpControlConfigService.loadReport({
      template: report,
      sectionId: this.section.id,
      standardId: this.section.standardId
    });
  }
}
