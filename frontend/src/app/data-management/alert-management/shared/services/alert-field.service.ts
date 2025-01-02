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


  findField(fields: UtmFieldType[], name: string): UtmFieldType | undefined {
    for (const field of fields) {
      if (field.field === name) {
        return field;
      }
      if (field.fields) {
        const nestedField = this.findField(field.fields, name);
        if (nestedField) {
          return nestedField;
        }
      }
    }
    return undefined;
  }
}
