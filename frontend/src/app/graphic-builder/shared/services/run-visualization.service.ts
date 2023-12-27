import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {VisualizationService} from '../../visualization/shared/services/visualization.service';

@Injectable({
  providedIn: 'root'
})
export class RunVisualizationService {

  constructor(private visualizationService: VisualizationService) {
  }

  /**
   * Method return observable of visualization run response
   * @param visualization Visualization to run
   */
  run(visualization: VisualizationType): Observable<any> {
    return new Observable<any>(subscriber => {
      if (typeof visualization.chartConfig !== 'string') {
        visualization.chartConfig = JSON.stringify(visualization.chartConfig);
      }
      if (typeof visualization.chartAction !== 'string') {
        visualization.chartAction = JSON.stringify(visualization.chartAction);
      }
      this.visualizationService.run(visualization).subscribe(resp => {
        if (typeof visualization.chartConfig === 'string') {
          visualization.chartConfig = JSON.parse(visualization.chartConfig);
        }
        if (typeof visualization.chartAction === 'string') {
          visualization.chartAction = JSON.parse(visualization.chartAction);
        }
        subscriber.next(resp.body);
      }, error => {
        subscriber.error(error);
      });
    });
  }

  parseError(error: any): string {
    // const errorObject = JSON.parse(error.headers.get('X-utmvaultApp-error'));
    //
    // return errorObject.root_cause.reason;
    return '';
  }
}


