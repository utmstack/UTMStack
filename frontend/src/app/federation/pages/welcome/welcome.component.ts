import {Component} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Observable} from 'rxjs';
import {map} from 'rxjs/operators';
import {InstanceFormModalComponent} from '../../components/instance-form-modal/instance-form-modal.component';
import {FederationInstance} from '../../domain/federation-instance.model';
import {FederationInstanceStateService} from '../../services/federation-instance-state.service';
import {FederationInstancesService} from '../../services/federation-instances.service';

@Component({
  selector: 'app-federation-welcome',
  templateUrl: './welcome.component.html',
  styleUrls: ['./welcome.component.scss']
})
export class WelcomeComponent {
  readonly hasInstances$: Observable<boolean> = this.instanceState.instances$
    .pipe(map(instances => instances.length > 0));

  constructor(
    private modalService: NgbModal,
    private router: Router,
    private instanceState: FederationInstanceStateService,
    private instancesService: FederationInstancesService
  ) {}

  openCreateModal(): void {
    const ref = this.modalService.open(InstanceFormModalComponent, {
      centered: true,
      backdrop: 'static',
      keyboard: false
    });
    ref.componentInstance.saved.subscribe((created: FederationInstance) => {
      ref.close();
      this.instancesService.list().subscribe(instances => {
        const list = instances || [];
        this.instanceState.setInstances(list);
        const next = list.find(i => i.id === created.id) || list[0];
        if (next) {
          this.instanceState.setActive(next, false);
        }
        this.router.navigate(['/dashboard/overview']);
      });
    });
  }
}
