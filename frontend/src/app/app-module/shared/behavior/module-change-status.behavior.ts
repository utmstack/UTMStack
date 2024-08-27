import {Inject, Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';

@Injectable({providedIn: 'root'})
export class ModuleChangeStatusBehavior {
  private moduleChangeStatusBehavior = new BehaviorSubject<{status: boolean, hasPreAction: boolean}>(null);
  public moduleStatus$: Observable<{status: boolean, hasPreAction: boolean}> = this.moduleChangeStatusBehavior.asObservable();

  getLastStatus() {
    return this.moduleChangeStatusBehavior.value.status;
  }
  setStatus(status: boolean, hasPreAction: boolean) {
    this.moduleChangeStatusBehavior.next({ status, hasPreAction});
  }
}
