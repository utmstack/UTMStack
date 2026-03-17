import {Component, OnDestroy, OnInit} from '@angular/core';
import {Subject} from 'rxjs';
import {filter, takeUntil} from 'rxjs/operators';
import {ComplianceStatusExtendedEnum} from '../shared/enums/compliance-status.enum';
import {ComplianceStrategyEnum} from '../shared/enums/compliance-strategy.enum';
import {CpControlConfigService} from '../shared/services/cp-control-config.service';
import {ComplianceControlEvaluationHistoryType} from '../shared/type/compliance-control-evaluation-history.type';
import {ComplianceControlLatestEvaluationType} from '../shared/type/compliance-control-latest-evaluation.type';

@Component({
  selector: 'app-compliance-evaluation-history-view',
  templateUrl: './compliance-evaluation-history-view.component.html',
  styleUrls: ['./compliance-evaluation-history-view.component.scss']
})
export class ComplianceEvaluationHistoryViewComponent implements OnInit, OnDestroy {
  controlId: number;
  currentEvaluation: ComplianceControlLatestEvaluationType;
  evaluationsHistory: ComplianceControlEvaluationHistoryType[];
  loading = false;
  showDetails = false;
  showRemediation = false;
  showSolution = false;
  selectedEvaluation: ComplianceControlEvaluationHistoryType;
  destroy$: Subject<void> = new Subject<void>();

  ComplianceStatusExtendedEnum = ComplianceStatusExtendedEnum;
  ComplianceStrategyEnum = ComplianceStrategyEnum;

  constructor(private cpControlConfigService: CpControlConfigService) {
  }

  ngOnInit() {
    this.loading = true;
    this.cpControlConfigService.onLoadControl$
      .pipe(takeUntil(this.destroy$),
            filter(params => !!params),
      ).subscribe(params => {
        this.currentEvaluation = params.template;
        if (this.currentEvaluation) {
          this.loadReport(params);
        }
      });
  }

  loadReport(params: any) {
    this.showDetails = false;
    this.controlId = params.template && params.template.id ? params.template.id : null;
    this.getEvaluations();
  }

  getEvaluations() {
    if (this.controlId) {
      this.loading = true;
      this.cpControlConfigService.evaluationsByControl(this.controlId)
        .subscribe(response => {
          this.evaluationsHistory = response.body.evaluations;
          this.loading = false;
        });
    }
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }

  showEvaluationDetails(evaluation: ComplianceControlEvaluationHistoryType) {
    this.showDetails = true;
    this.selectedEvaluation = evaluation;
  }
}
