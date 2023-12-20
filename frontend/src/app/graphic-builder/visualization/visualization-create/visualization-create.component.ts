import {Component, Input, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {ChartType} from '../../../shared/chart/types/chart.type';
import {UTM_CHART_ICONS} from '../../../shared/constants/icons-chart.const';
import {ChartTypeEnum} from '../../../shared/enums/chart-type.enum';
import {RouteCallbackEnum} from '../../../shared/enums/route-callback.enum';
import {UtmIndexPattern} from '../../../shared/types/index-pattern/utm-index-pattern';
import {CHART_TYPES} from '../../shared/const/chart-type.const';
import {VisualizationQueryParamsEnum} from '../../shared/enums/visualization-query-params.enum';

@Component({
  selector: 'app-visualization-create',
  templateUrl: './visualization-create.component.html',
  styleUrls: ['./visualization-create.component.scss']
})
export class VisualizationCreateComponent implements OnInit {
  @Input() callback: RouteCallbackEnum;
  charts = CHART_TYPES;
  chartSelected: ChartType;
  chartSearch: string;
  creating = false;
  pattern: UtmIndexPattern;

  constructor(public activeModal: NgbActiveModal,
              private router: Router) {
  }

  ngOnInit() {
  }

  chartIconResolver(chartType: string) {
    if (chartType !== undefined) {
      return UTM_CHART_ICONS[chartType];
    }
  }

  chartBlur(chart: ChartType) {
    this.chartSelected = chart;
  }

  searchChart($event: Event) {
    this.charts = this.charts.filter((value) => {
      return value.name.toLowerCase().includes(this.chartSearch.toLowerCase());
    });
  }

  createVisualization() {
    this.creating = true;
    const queryParams = {};
    queryParams[VisualizationQueryParamsEnum.PATTERN_NAME] = this.pattern.pattern;
    queryParams[VisualizationQueryParamsEnum.PATTERN_ID] = this.pattern.id;
    queryParams[VisualizationQueryParamsEnum.CHART] = this.chartSelected.type;
    queryParams[VisualizationQueryParamsEnum.CALLBACK] = this.callback;
    const route = this.chartSelected.type === ChartTypeEnum.TEXT_CHART ?
      '/creator/builder/text-builder' : '/creator/builder/chart-builder';
    this.router.navigate([route],
      {
        queryParams
      }).then(() => {
      this.creating = false;
      this.activeModal.close();
    });
  }
}
