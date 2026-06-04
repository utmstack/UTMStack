import {Injectable} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {ClickFactory} from '../../../shared/chart/factories/echart-click-factory/click-factory';
import {EchartClickAction} from '../../../shared/chart/types/action/echart-click-action';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {resolveChartNavigationUrl} from '../../../shared/util/resolve-chart-navigation-link';

@Injectable({
  providedIn: 'root'
})
export class UtmChartClickActionService {
  chartClickFactory = new ClickFactory();

  constructor(private spinner: NgxSpinnerService,
              private router: Router) {
  }

  /**
   * @param visualization Visualization
   * @param chartClickAction EchartClickAction
   * Manage click route navigation on chart click
   * @param forceSingle Force to single on charts with multiple buckets, used by single bar chart navigation
   */
  onClickNavigate(visualization: VisualizationType, chartClickAction: EchartClickAction, forceSingle?: boolean) {
    if (visualization.chartAction.active) {
      const queryParams = this.chartClickFactory.createParams(visualization, chartClickAction, forceSingle);
      this.spinner.show('loadingSpinner');
      this.router.navigate([resolveChartNavigationUrl(visualization)], {queryParams})
        .then(() => {
          this.spinner.hide('loadingSpinner');
        });
    } else {
      console.log('NAVIGATION IS DISABLED FOR THIS CHART');
    }
  }
}
