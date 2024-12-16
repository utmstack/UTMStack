import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {UtmFieldType} from '../../../../shared/types/table/utm-field.type';

@Injectable({
  providedIn: 'root'
})
export class AlertFieldService {
  private fieldUpdateBehavior: BehaviorSubject<UtmFieldType> = new BehaviorSubject(null);
  public field$ = this.fieldUpdateBehavior.asObservable();

  update(field: UtmFieldType) {
    this.fieldUpdateBehavior.next(field);
  }
}
