import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {ComplianceStatusExtendedEnum, getComplianceStatusLabel} from '../../enums/compliance-status.enum';
import {ComplianceControlEvaluationsType} from '../../type/compliance-control-evaluations.type';

@Component({
  selector: 'app-compliance-timeline',
  templateUrl: './compliance-timeline.component.html',
  styleUrls: ['./compliance-timeline.component.scss']
})
export class ComplianceTimelineComponent implements OnInit {
  @Input() evaluations: ComplianceControlEvaluationsType[];
  @Output() evaluationSelected = new EventEmitter<ComplianceControlEvaluationsType>();
  private originalData: ComplianceControlEvaluationsType[] = [];

  options: any;
  complianceValue = [
    'COMPLIANT',
    'NON_COMPLIANT',
  ];
  bgColors = {
    COMPLIANT: '#28A745',
    NON_COMPLIANT: '#DC3545',
  };
  textColors = {
    COMPLIANT: '#ffffff',
    NON_COMPLIANT: '#ffffff',
  };
  lightBgColors = {
    COMPLIANT: '#E8F5E9',
    NON_COMPLIANT: '#FAEBED',
  };

  private rawData: any[];
  private monthDays: string[] = [];

  constructor() {
  }

  ngOnInit(): void {
    this.options = null;

    if (this.evaluations && this.evaluations.length > 0) {
      this.transformResponse(this.evaluations);
      if (this.rawData.length > 0) {
        this.initChartOptions();
      }
    }
  }

  private transformResponse(data: ComplianceControlEvaluationsType[]): void {
    this.rawData = [];
    this.monthDays = [];
    this.originalData = this.evaluations;

    data.forEach((entry, index) => {
      const date = new Date(entry.timestamp);
      const dayLabel = date.toLocaleDateString('en-US', {
        month: 'short',
        day: '2-digit',
        year: 'numeric',
      });
      this.monthDays.push(dayLabel);

      this.rawData.push([index, 0, entry.status === ComplianceStatusExtendedEnum.COMPLIANT ? 1 : 0]);
      this.rawData.push([index, 1, entry.status === ComplianceStatusExtendedEnum.NON_COMPLIANT ? 1 : 0]);
    });
  }

  private initChartOptions(): void {
    this.options = {
      tooltip: {
        position: 'top',
        formatter: (params: any) => {
          const point = Array.isArray(params) ? params[0] : params;
          if (
            !point ||
            typeof point.value !== 'object' ||
            point.value === null ||
            !Array.isArray(point.value)
          ) {
            return '';
          }

          const [dayIdx, sevIdx, count] = point.value as [
            number,
            number,
            number
          ];

          if (
            dayIdx >= this.monthDays.length ||
            sevIdx >= this.complianceValue.length
          ) {
            return 'Invalid data';
          }

          const day = this.monthDays[dayIdx];
          const value = this.complianceValue[sevIdx];
          const markerColor = count === 0 ? this.lightBgColors[value] : this.bgColors[value];
          // tslint:disable-next-line:max-line-length
          const marker = `<span style="display:inline-block;margin-right:5px;border-radius:3px;width:10px;height:10px;background-color:${markerColor};"></span>`;
          return count !== 0 ?
            `${marker} ${day} (${getComplianceStatusLabel(value)})`
            : '';
        },
      },
      grid: {
        height: '50%',
        top: '15%',
        left: '10%',
        right: '3%',
        bottom: '5%',
        width: this.evaluations.length * 30,
        containLabel: false
      },
      xAxis: {
        type: 'category',
        data: this.monthDays,
        axisLabel: { rotate: 45, fontSize: 10, interval: 0 },
        axisLine: { show: false },
        axisTick: { alignWithLabel: true },
      },
      yAxis: {
        type: 'category',
        data: this.complianceValue.map((value) =>
          value
        ),
        axisLine: { show: false },
        axisTick: { show: false },
        axisLabel: { fontSize: 10 }
      },
      series: [
        {
          name: 'Compliance Timeline',
          type: 'custom',
          renderItem: this.renderItem.bind(this),
          data: this.rawData,
          emphasis: {
            disabled: true
          },
          encode: {
            x: 0,
            y: 1,
            tooltip: [0, 1, 2],
          },
          silent: false,
        },
      ],
    };
  }

  private renderItem(
    params: any,
    api: any): any {
    const dayIndex = api.value(0) as number;
    const valueIndex = api.value(1) as number;
    const count = api.value(2) as number;
    const cellCenterCoord = api.coord([dayIndex, valueIndex]);
    const cellWidth = 25;
    const cellHeight = 45;
    const x = cellCenterCoord[0] - cellWidth / 2;
    const y = cellCenterCoord[1] - cellHeight / 2;
    const value = this.complianceValue[valueIndex];
    const baseColor = this.bgColors[value];
    const fillColor = count === 0 ? this.lightBgColors[value] : baseColor;
    const textColor = count > 0 ? this.textColors[value] : undefined;
    const rectShape = {
      x,
      y,
      width: cellWidth,
      height: cellHeight,
      r: 4,
    };
    const rectStyle = {
      fill: fillColor,
      lineWidth: 1,
    };
    const rectElement = {
      type: 'rect',
      shape: rectShape,
      style: {
        ...rectStyle,
        cursor: 'pointer',
      },
    };

    const childrenElements: any[] = [rectElement];

    if (count > 0 && textColor) {
      childrenElements.push({
        type: 'text',
        style: {
          fill: textColor,
          x: cellCenterCoord[0],
          y: cellCenterCoord[1],
          textAlign: 'center',
          textVerticalAlign: 'middle',
          fontSize: 11,
          fontWeight: 'bold',
        },
      });
    }
    return {
      type: 'group',
      children: childrenElements,
      silent: count === 0,
      info: {
        value: [dayIndex, valueIndex, count],
      },
    };
  }

  onChartClick(event: any): void {
    const value = event ? event.data : null;
    if (Array.isArray(value)) {
      const [dayIndex] = value;
      this.evaluationSelected.emit(this.originalData[dayIndex]);
    }
  }
}
