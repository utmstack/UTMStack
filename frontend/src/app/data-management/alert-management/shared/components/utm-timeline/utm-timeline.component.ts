import {Component, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges} from '@angular/core';
import {TimelineItem} from '../../../../../shared/types/utm-timeline-item';
import {TimelineGroup, TimelineService} from "./timeline.service";


@Component({
  selector: 'utm-timeline',
  templateUrl: './utm-timeline.component.html',
  styleUrls: ['./utm-timeline.component.scss']
})
export class UtmTimeLineComponent implements OnInit {

  @Input() items: TimelineItem[] = [];
  @Input() title = '';
  @Output() itemClick = new EventEmitter<TimelineItem>();


  chartOption: any = {};
  overviewData: number[] = [];
  intervalMs = 60 * 1000;
  groups: TimelineGroup[] = [];
  page = 0;
  pageSize = 10000;
  readonly Math = Math;

  ngOnInit(): void {
    if (this.items && this.items.length > 0) {
      this.groups = this.timelineService.groupByInterval(this.items, this.intervalMs);

      this.buildChart();
    }
  }

  constructor(private timelineService: TimelineService) {}

  buildChart() {
    const seriesData = [];

    this.groups.forEach((group, index) => {
      const rep = group.items[0] || ({} as any); // representative item
      seriesData.push({
        value: [
          group.startTimestamp,                         // 0: timestamp (start of minute)
          0,                                            // 1: y coordinate (not used)
          rep.name || `Echoes`,                         // 2: representative name/title
          new Date(group.startTimestamp).toLocaleString(), // 3: formatted minute
          rep.iconUrl || 'assets/images/default-echo.png', // 4: icon url
          group.items.length                             // 5: count of echoes
        ],
        groupData: group.items,                         // full list for drill-down
        indexForOffset: index                            // incremental index para posicion vertical
      });
    });

    const firstTimestamp = Math.min(...seriesData.map(d => d.value[0]));
    const lastTimestamp = Math.max(...seriesData.map(d => d.value[0]));
    const padding = 5 * 180 * 1000;

    this.chartOption = {
      title: { text: this.title, left: 'center', textStyle: { fontSize: 16, fontWeight: 'bold' } },
      tooltip: {
        trigger: 'item',
        formatter: (params: any) =>
          `<b>Echoes:</b> ${params.data.value[2]}<br/><b>Minute:</b> ${params.data.value[3]}<br/><b>Total:</b> ${params.data.value[5]}`
      },
      grid: {
        left: 0,
        right: 0,
        top: 0,
        bottom: 20,
        containLabel: true
      },
      xAxis: {
        type: 'time',
        boundaryGap: false,
        min: firstTimestamp - padding,
        max: lastTimestamp + padding,
        axisLabel: { formatter: (val: number) => new Date(val).toLocaleTimeString() }
      },
      yAxis: { show: false, type: 'value' },
      dataZoom: [
        { type: 'slider', xAxisIndex: 0, start: 0, end: 100 },
        { type: 'inside', xAxisIndex: 0, zoomLock: false }
      ],
      series: [
        {
          type: 'custom',
          data: seriesData,
          renderItem: (params: any, api: any) => {
            const ts = api.value(0);
            const coord = api.coord([ts, 0]);
            const cardWidth = 250;
            const cardHeight = 60;
            const spacing = 10;
            const offsetIndex = api.value('indexForOffset') || 0;
            const totalOffset = (cardHeight + spacing) * offsetIndex + 80;

            return {
              type: 'group',
              children: [
                // Background card
                {
                  type: 'rect',
                  shape: {
                    x: coord[0] - cardWidth / 2,
                    y: coord[1] - totalOffset - cardHeight,
                    width: cardWidth,
                    height: cardHeight,
                    r: 10
                  },
                  style: {
                    fill: '#ffffff',
                    stroke: '#0277bd',
                    lineWidth: 1,
                    shadowBlur: 8,
                    shadowColor: 'rgba(0,0,0,0.2)',
                    cursor: 'pointer'
                  }
                },

                // Icon
                {
                  type: 'image',
                  style: {
                    image: api.value(4),
                    x: coord[0] - cardWidth / 2 + 5,
                    y: coord[1] - totalOffset - cardHeight + 5,
                    width: cardHeight - 10, // icono más pequeño y centrado
                    height: cardHeight - 10
                  }
                },

                // Title
                {
                  type: 'text',
                  style: {
                    x: coord[0] - cardWidth / 2 + (cardHeight - 2.5) + 15,
                    y: coord[1] - totalOffset - cardHeight + 5,
                    text: this.truncateText(api.value(2) || '', 150),
                    textAlign: 'left',
                    fill: '#000',
                    fontSize: 14,
                    fontWeight: 600,
                    width: cardWidth - (cardHeight - 2.5) - 25,
                    overflow: 'break',
                    ellipsis: '...'
                  }
                },

                // Formatted minute/date
                {
                  type: 'text',
                  style: {
                    x: coord[0] - cardWidth / 2 + (cardHeight - 2.5) + 15,
                    y: coord[1] - totalOffset - cardHeight + 25,
                    text: api.value(3),
                    textAlign: 'left',
                    fill: '#666',
                    fontSize: 12
                  }
                },

                // Total echoes
                {
                  type: 'text',
                  style: {
                    x: coord[0] - cardWidth / 2 + (cardHeight - 2.5) + 15,
                    y: coord[1] - totalOffset - cardHeight + 42,
                    text: `Total: ${api.value(5)} echoes`,
                    textAlign: 'left',
                    fill: '#444',
                    fontSize: 12,
                    fontWeight: 'bold'
                  }
                },

                // Connector line
                {
                  type: 'line',
                  shape: {
                    x1: coord[0],
                    y1: coord[1] - totalOffset,
                    x2: coord[0],
                    y2: coord[1]
                  },
                  style: {
                    stroke: '#0277bd',
                    lineWidth: 1.5
                  }
                }
              ]
            };
          },
          encode: { x: 0, y: 1 }
        }
      ]
    };
  }

  /**
   * Emits the clicked timeline item
   */
  onChartClick(event: any) {
    if (event.data && event.data.itemData) {
      this.itemClick.emit(event.data.itemData);
    }
  }

  /**
   * Truncates long text for cards
   */
  truncateText(text: string, maxWidth: number) {
    const avgCharWidth = 7;
    const maxChars = Math.floor(maxWidth / avgCharWidth);
    return text.length > maxChars ? text.substring(0, maxChars - 3) + '...' : text;
  }
}
