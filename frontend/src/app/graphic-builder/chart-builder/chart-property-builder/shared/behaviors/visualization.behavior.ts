import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';

@Injectable({
  providedIn: 'root'
})
export class VisualizationBehavior {
  $visualization = new BehaviorSubject<VisualizationType>(null);
}
