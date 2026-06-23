import {Component, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {forkJoin, of, Subject} from 'rxjs';
import {catchError, takeUntil} from 'rxjs/operators';
import {resolveRangeByTime} from '../../../shared/util/resolve-date';
import {CountAlertsBySeverityEntry} from '../../domain/count-alerts-by-severity.model';
import {CountAlertsByStatusEntry} from '../../domain/count-alerts-by-status.model';
import {FederationInstanceStateService} from '../../services/federation-instance-state.service';
import {FederationOverviewService} from '../../services/federation-overview.service';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {FederationInstance} from '../../domain/federation-instance.model';
import {FederationInstancesService} from '../../services/federation-instances.service';
import {InstanceFormModalComponent} from '../instance-form-modal/instance-form-modal.component';
import {
  ModalConfirmationComponent
} from '../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';

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
  pendingDeleteId: number | null = null;

  private destroy$ = new Subject<void>();

  constructor(
    private overviewService: FederationOverviewService,
    private instanceState: FederationInstanceStateService,
    private router: Router,
    private instancesService: FederationInstancesService,
    private modalService: NgbModal
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

    this.router.navigateByUrl(route);
  }

  trackByInstance(_index: number, entry: CountAlertsByStatusEntry): number {
    return entry.instanceId;
  }



  openCreate(event?: Event): void {
    if (event) {
      event.stopPropagation();
    }
    this.errorMessage = null;
    const ref = this.modalService.open(InstanceFormModalComponent, {
      centered: true,
      backdrop: 'static'
    });
    ref.componentInstance.saved.subscribe((instance: FederationInstance) => {
      ref.close();
      this.reload(instance.id, true);
    });
  }

  openEdit(instance: CountAlertsByStatusEntry, event?: Event): void {
    if (event) {
      event.stopPropagation();
    }
    this.errorMessage = null;
    const ref = this.modalService.open(InstanceFormModalComponent, {
      centered: true,
      backdrop: 'static'
    });
    const target = this.instanceState.instances.find(i => i.id === instance.instanceId);
    ref.componentInstance.instance = target;
    ref.componentInstance.saved.subscribe((updated: FederationInstance) => {
      ref.close();
      this.reload(updated.id, false);
    });
  }

  remove(instance: CountAlertsByStatusEntry, event?: Event): void {
    if (event) {
      event.stopPropagation();
    }
    if (this.pendingDeleteId !== null) {
      return;
    }

    const target = this.instanceState.instances.find(i => i.id === instance.instanceId);
    if(!target)return
    const modalRef = this.modalService.open(ModalConfirmationComponent, {backdrop: 'static', centered: true});
    modalRef.componentInstance.header = 'Remove federation instance';
    modalRef.componentInstance.message = `Are you sure that you want to remove the instance: \n${target.name}?`;
    modalRef.componentInstance.confirmBtnText = 'Remove';
    modalRef.componentInstance.confirmBtnIcon = 'icon-cancel-circle2';
    modalRef.componentInstance.confirmBtnType = 'delete';
    modalRef.componentInstance.textDisplay = 'Requests routed to this instance will stop working' +
      ' and its locally stored selection will be cleared.';
    modalRef.componentInstance.textType = 'warning';
    modalRef.result.then(() => this.confirmRemove(target), () => undefined);
  }

  trackById(_index: number, item: FederationInstance): number {
    return item.id;
  }

  private confirmRemove(instance: FederationInstance): void {
    this.pendingDeleteId = instance.id;
    this.errorMessage = null;
    this.instancesService.delete(instance.id).subscribe({
      next: () => {
        this.pendingDeleteId = null;
        this.reload(null, false);
      },
      error: err => {
        this.pendingDeleteId = null;
        this.errorMessage = (err && err.error && err.error.message)
          || 'Failed to remove instance.';
      }
    });
  }

  private reload(preferredId: number | null, forceSwitch: boolean): void {
    this.instancesService.list().subscribe({
      next: items => {
        const list = items || [];
        this.instanceState.setInstances(list);
        if (list.length === 0) {
          window.location.reload();
          return;
        }
        if (preferredId !== null) {
          const target = list.find(i => i.id === preferredId);
          if (target) {
            this.instanceState.setActive(target, false);
            return;
          }
        }
        if (forceSwitch) {
          window.location.reload();
        }
        this.loadData();
      },
      error: () => {
        this.errorMessage = 'Failed to refresh instances.';
      }
    });
  }

}
