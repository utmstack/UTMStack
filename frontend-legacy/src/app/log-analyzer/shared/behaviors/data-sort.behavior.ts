import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({providedIn: 'root'})
export class DataSortBehavior {
  $dataSort = new BehaviorSubject<string>(null);
}



