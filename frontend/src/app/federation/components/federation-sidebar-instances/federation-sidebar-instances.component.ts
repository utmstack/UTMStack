import {Component, OnDestroy, OnInit} from '@angular/core';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {SYSTEM_MENU_ICONS_PATH} from '../../../shared/constants/menu_icons.constants';
import {FederationInstance} from '../../domain/federation-instance.model';
import {FederationInstanceStateService} from '../../services/federation-instance-state.service';

@Component({
  selector: 'app-federation-sidebar-instances',
  templateUrl: './federation-sidebar-instances.component.html',
  styleUrls: ['./federation-sidebar-instances.component.scss']
})
export class FederationSidebarInstancesComponent implements OnInit, OnDestroy {
  iconPath = SYSTEM_MENU_ICONS_PATH;
  instances: FederationInstance[] = [];
  active: FederationInstance | null = null;
  pendingDeleteId: number | null = null;
  errorMessage: string | null = null;
  private destroy$ = new Subject<void>();

  constructor(private instanceState: FederationInstanceStateService,
) {}


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


  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
