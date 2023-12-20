import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ChangeFilterValueService {
  private selectedValueSubject = new BehaviorSubject<{ field: any, value: any }>(null);

  selectedValue$: Observable<any> = this.selectedValueSubject.asObservable();

  changeSelectedValue(newValues: { field: any, value: any }): void {
    this.selectedValueSubject.next(newValues);
  }
}
