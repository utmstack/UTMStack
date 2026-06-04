import {Component, Input, OnInit} from '@angular/core';
import {VsSeverityResolverService} from '../../../../shared/services/scan/vs-severity-resolver.service';

@Component({
  selector: 'app-asset-severity-chart',
  templateUrl: './asset-severity-chart.component.html',
  styleUrls: ['./asset-severity-chart.component.scss']
})
export class AssetSeverityChartComponent implements OnInit {
  @Input() severity: string;
  @Input() type: 'card' | 'row';
  pieOption: any;
  firstLoad = true;

  constructor(public severityResolverService: VsSeverityResolverService) {
  }

  ngOnInit() {
    const dataStyle = {
      normal: {
        borderWidth: 1,
        borderColor: this.severityResolverService.resolveColor(this.severity),
        label: {show: false},
        labelLine: {show: false}
      }
    };
    const placeHolderStyle = {
      cursor: 'pointer',
      normal: {
        color: '#f3f3f3f3',
        borderWidth: 0
      }
    };
    this.pieOption = {
      // Colors
      animation: this.firstLoad,
      color: [this.severityResolverService.resolveColor(this.severity)],

      // Global text styles
      textStyle: {
        fontFamily: 'Roboto, Arial, Verdana, sans-serif',
        fontSize: this.type === 'card' ? 12 : 11
      },
      //
      // Add title
      title: {
        cursor: 'pointer',
        text: this.severityResolverService.resolveSeverityLabel(this.severity),
        left: this.type === 'card' ? 'center' : 'right',
        top: this.type === 'card' ? '33%' : '30%',
        textStyle: {
          cursor: 'pointer',
          color: this.severityResolverService.resolveColor(this.severity),
          fontSize: this.type === 'card' ? 13 : 12,
          fontWeight: 300
        },
        subtextStyle: {
          fontSize: this.type === 'card' ? 13 : 12
        }
      },

      series: [
        {
          type: 'pie',
          cursor: 'pointer',
          clockWise: false,
          radius: ['89%', '90%'],
          hoverOffset: 0,
          itemStyle: dataStyle,
          data: [
            {
              value: Number.parseFloat(this.severity) <= 0
                ? 10 : Number.parseFloat(this.severity),
              itemStyle: {
                cursor: 'pointer',
              }
            },
            {
              value: 10 - (Number.parseFloat(this.severity) === 0
                ? 10 : Number.parseFloat(this.severity)),
              itemStyle: placeHolderStyle
            }
          ]
        },
      ]
    };
    this.firstLoad = false;
  }

  isGreaterThanCero() {
    return Number.parseFloat(this.severity) < 0;
  }
}
