import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {EventDataTypeEnum} from '../enums/event-data-type.enum';

@Injectable({
  providedIn: 'root'
})
export class AlertDataTypeBehavior {
  $alertDataType = new BehaviorSubject<EventDataTypeEnum>(null);
}
