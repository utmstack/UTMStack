import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';

@Injectable({providedIn: 'root'})
export class ModuleChangeStatusBehavior {
  private moduleChangeStatusBehavior = new BehaviorSubject<{status: boolean, hasPreAction: boolean}>(null);
  public moduleStatus$: Observable<{status: boolean, hasPreAction: boolean}> = this.moduleChangeStatusBehavior.asObservable();

  getLastStatus() {
    return this.moduleChangeStatusBehavior.value && this.moduleChangeStatusBehavior.value.status ?
        this.moduleChangeStatusBehavior.value.status : false;
  }
  setStatus(status: boolean, hasPreAction: boolean) {
    this.moduleChangeStatusBehavior.next({ status, hasPreAction});
  }
}
