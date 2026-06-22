import {Component, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {resolveRangeByTime} from '../../../shared/util/resolve-date';
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
    this.overviewService.countAlertsByStatus(range.timeFrom, range.timeTo)
      .pipe(takeUntil(this.destroy$))
      .subscribe(
        entries => {
          this.entries = entries || [];
          this.loading = false;
        },
        () => {
          this.entries = [];
          this.errorMessage = 'Could not load federation overview.';
          this.loading = false;
        }
      );
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
