import {Component, OnDestroy, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {
  ModalConfirmationComponent
} from '../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {FederationInstance} from '../../domain/federation-instance.model';
import {FederationInstanceStateService} from '../../services/federation-instance-state.service';
import {FederationInstancesService} from '../../services/federation-instances.service';
import {InstanceFormModalComponent} from '../instance-form-modal/instance-form-modal.component';

@Component({
  selector: 'app-federation-instance-switcher',
  templateUrl: './instance-switcher.component.html',
  styleUrls: ['./instance-switcher.component.scss']
})
export class InstanceSwitcherComponent implements OnInit, OnDestroy {
  instances: FederationInstance[] = [];
  active: FederationInstance | null = null;
  pendingDeleteId: number | null = null;
  errorMessage: string | null = null;
  private destroy$ = new Subject<void>();

  constructor(private instanceState: FederationInstanceStateService,
              private instancesService: FederationInstancesService,
              private modalService: NgbModal) {}

  ngOnInit(): void {
    this.instanceState.instances$
      .pipe(takeUntil(this.destroy$))
      .subscribe(items => this.instances = items);
    this.instanceState.active$
      .pipe(takeUntil(this.destroy$))
      .subscribe(item => this.active = item);
  }

  select(instance: FederationInstance, event?: Event): void {
    if (event) {
      event.stopPropagation();
    }
    if (this.active && this.active.id === instance.id) {
      return;
    }
    this.instanceState.setActive(instance);
    window.location.reload();
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

  openEdit(instance: FederationInstance, event?: Event): void {
    if (event) {
      event.stopPropagation();
    }
    this.errorMessage = null;
    const ref = this.modalService.open(InstanceFormModalComponent, {
      centered: true,
      backdrop: 'static'
    });
    ref.componentInstance.instance = instance;
    ref.componentInstance.saved.subscribe((updated: FederationInstance) => {
      ref.close();
      this.reload(updated.id, false);
    });
  }

  remove(instance: FederationInstance, event?: Event): void {
    if (event) {
      event.stopPropagation();
    }
    if (this.pendingDeleteId !== null) {
      return;
    }
    const modalRef = this.modalService.open(ModalConfirmationComponent, {backdrop: 'static', centered: true});
    modalRef.componentInstance.header = 'Remove federation instance';
    modalRef.componentInstance.message = `Are you sure that you want to remove the instance: \n${instance.name}?`;
    modalRef.componentInstance.confirmBtnText = 'Remove';
    modalRef.componentInstance.confirmBtnIcon = 'icon-cancel-circle2';
    modalRef.componentInstance.confirmBtnType = 'delete';
    modalRef.componentInstance.textDisplay = 'Requests routed to this instance will stop working' +
      ' and its locally stored selection will be cleared.';
    modalRef.componentInstance.textType = 'warning';
    modalRef.result.then(() => this.confirmRemove(instance), () => undefined);
  }

  private confirmRemove(instance: FederationInstance): void {
    this.pendingDeleteId = instance.id;
    this.errorMessage = null;
    this.instancesService.delete(instance.id).subscribe({
      next: () => {
        this.pendingDeleteId = null;
        const wasActive = this.active && this.active.id === instance.id;
        this.reload(null, wasActive);
      },
      error: err => {
        this.pendingDeleteId = null;
        this.errorMessage = (err && err.error && err.error.message)
          || 'Failed to remove instance.';
      }
    });
  }

  trackById(_index: number, item: FederationInstance): number {
    return item.id;
  }

  private reload(preferredId: number | null, forceSwitch: boolean): void {
    this.instancesService.list().subscribe({
      next: items => {
        const list = items || [];
        const previousActiveId = this.active ? this.active.id : null;
        this.instanceState.setInstances(list);
        if (list.length === 0) {
          window.location.reload();
          return;
        }
        if (preferredId !== null) {
          const target = list.find(i => i.id === preferredId);
          if (target) {
            this.instanceState.setActive(target, false);
            if (forceSwitch || target.id !== previousActiveId) {
              window.location.reload();
            }
            return;
          }
        }
        if (forceSwitch) {
          window.location.reload();
        }
      },
      error: () => {
        this.errorMessage = 'Failed to refresh instances.';
      }
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
