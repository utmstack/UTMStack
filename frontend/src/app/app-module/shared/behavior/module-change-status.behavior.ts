import {Inject, Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';

@Injectable({providedIn: 'root'})
export class ModuleChangeStatusBehavior {
  private $moduleChangeStatusBehavior = new BehaviorSubject<boolean>(null);
  public moduleStatus$: Observable<boolean> = this.$moduleChangeStatusBehavior.asObservable();

  setStatus(status: boolean) {
    this.$moduleChangeStatusBehavior.next(status);
  }
}
