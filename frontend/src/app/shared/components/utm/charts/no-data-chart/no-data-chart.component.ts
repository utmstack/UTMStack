import {Component, Input, OnInit} from '@angular/core';
import {UTM_CHART_ICONS} from '../../../../constants/icons-chart.const';
import {ChartTypeEnum} from '../../../../enums/chart-type.enum';

@Component({
  selector: 'app-no-data-chart',
  templateUrl: './no-data-chart.component.html',
  styleUrls: ['./no-data-chart.component.scss']
})
export class NoDataChartComponent implements OnInit {
  @Input() typeChart: ChartTypeEnum;
  @Input() error = false;

  constructor() {
  }

  ngOnInit() {
  }

  resolveIcon() {
    if (this.typeChart !== undefined) {
      return UTM_CHART_ICONS[this.typeChart];
    }
  }
}
