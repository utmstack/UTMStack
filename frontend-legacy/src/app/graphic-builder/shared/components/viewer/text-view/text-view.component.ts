import {Component, Input, OnInit} from '@angular/core';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';

@Component({
  selector: 'app-text-view',
  templateUrl: './text-view.component.html',
  styleUrls: ['./text-view.component.scss']
})
export class TextViewComponent implements OnInit {
  @Input() chartId: number;
  @Input() building: boolean;
  @Input() chart: ChartTypeEnum;
  @Input() visualization: VisualizationType;
  @Input() width: string;
  @Input() height: string;
  @Input() exportFormat: boolean;
  loadingOption: boolean;

  constructor() {
  }

  ngOnInit() {
  }

}
