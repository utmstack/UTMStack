import {Injectable, Type} from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, RouterStateSnapshot} from '@angular/router';
import {GridsterItem} from 'angular-gridster2';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable} from 'rxjs';
import {map, tap} from 'rxjs/operators';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {RenderLayoutService} from './render-layout.service';
import {UtmRenderVisualization} from './utm-render-visualization.service';

@Injectable()
export class DashboardResolverService implements Resolve<{ grid: GridsterItem, visualization: VisualizationType } []> {

  constructor(private spinner: NgxSpinnerService,
              private utmRenderVisualization: UtmRenderVisualization,
              private layoutService: RenderLayoutService) {
  }

  resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot):
    Observable<{ grid: GridsterItem, visualization: VisualizationType } []> {
    const request = {
      page: 0,
      size: 10000,
      'idDashboard.equals': route.params.id,
      sort: 'order,asc'
    };
    this.spinner.show('loadingSpinner');
    return this.utmRenderVisualization.query(request)
      .pipe(
        map((response) => {
            this.layoutService.clearLayout();
            const visualizations = response.body;
            this.layoutService.dashboard = visualizations.length > 0 ? visualizations[0].dashboard : null;

            visualizations.forEach((vis) => {
              const visualization: VisualizationType = {...vis.visualization, showTime: vis.showTimeFilter};
              this.layoutService.addItem(visualization, JSON.parse(vis.gridInfo));
            });

            return this.layoutService.layout;
          }
        ),
        tap(() => {
          this.spinner.hide('loadingSpinner');
        }));
  }
}
