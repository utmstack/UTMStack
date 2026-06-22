import {Component, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {forkJoin, of, Subject} from 'rxjs';
import {catchError, takeUntil} from 'rxjs/operators';
import {resolveRangeByTime} from '../../../shared/util/resolve-date';
import {CountAlertsBySeverityEntry} from '../../domain/count-alerts-by-severity.model';
import {CountAlertsByStatusEntry} from '../../domain/count-alerts-by-status.model';
import {FederationInstanceStateService} from '../../services/federation-instance-state.service';
import {FederationOverviewService} from '../../services/federation-overview.service';

@Component({
  selector: 'app-federation-overview-grid',
  templateUrl: './federation-overview-grid.component.html',
  styleUrls: ['./federation-overview-grid.component.scss']
})
export class FederationOverviewGridComponent implements OnInit, OnDestroy {
  entries: CountAlertsByStatusEntry[] = [];
  severityByInstance: Record<number, CountAlertsBySeverityEntry> = {};
  loading = false;
  errorMessage: string | null = null;

  private destroy$ = new Subject<void>();

  constructor(
    private overviewService: FederationOverviewService,
    private instanceState: FederationInstanceStateService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.loadData();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  loadData(): void {
    const range = resolveRangeByTime('week');
    if (!range.timeFrom || !range.timeTo) {
      return;
    }
    this.loading = true;
    this.errorMessage = null;
    forkJoin([
      this.overviewService.countAlertsByStatus(range.timeFrom, range.timeTo),
      this.overviewService.countAlertsBySeverity(range.timeFrom, range.timeTo)
        .pipe(catchError(() => of([] as CountAlertsBySeverityEntry[])))
    ])
      .pipe(takeUntil(this.destroy$))
      .subscribe(
        ([status, severity]: [CountAlertsByStatusEntry[], CountAlertsBySeverityEntry[]]) => {
          this.entries = status || [];
          this.severityByInstance = {};
          (severity || []).forEach(entry => {
            this.severityByInstance[entry.instanceId] = entry;
          });
          this.loading = false;
        },
        () => {
          this.entries = [];
          this.severityByInstance = {};
          this.errorMessage = 'Could not load federation overview.';
          this.loading = false;
        }
      );
  }

  severityFor(entry: CountAlertsByStatusEntry): CountAlertsBySeverityEntry | null {
    return this.severityByInstance[entry.instanceId] || null;
  }

  onSelect(entry: CountAlertsByStatusEntry, route: string): void {
    const target = this.instanceState.instances.find(i => i.id === entry.instanceId);
    if (target) {
      this.instanceState.setActive(target, true);
    }
    this.router.navigate([route]);
  }

  trackByInstance(_index: number, entry: CountAlertsByStatusEntry): number {
    return entry.instanceId;
  }
}
