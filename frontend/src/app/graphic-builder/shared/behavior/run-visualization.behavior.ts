import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class RunVisualizationBehavior {
  $run = new BehaviorSubject<number>(null);
}
