import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';

@Component({
  selector: 'app-utm-viewer',
  templateUrl: './utm-viewer.component.html',
  styleUrls: ['./utm-viewer.component.scss']
})
export class UtmViewerComponent implements OnInit {
  @Input() chartId: number;
  @Input() building: boolean;
  @Input() chart: ChartTypeEnum;
  @Input() visualization: VisualizationType;
  @Input() width: string;
  @Input() height: string;
  @Input() showTime: boolean;
  @Input() timeByDefault: any;
  @Output() runned = new EventEmitter<string>();
  @Input() exportFormat: boolean;
  chartTypeEnum = ChartTypeEnum;

  constructor() {
  }

  ngOnInit() {
  }

  onRunVisualization($event: string) {
    this.runned.emit($event);
  }
}
