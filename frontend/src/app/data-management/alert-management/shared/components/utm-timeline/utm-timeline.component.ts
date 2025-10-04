import {group} from '@angular/animations';
import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges} from '@angular/core';
import {text} from '@fortawesome/fontawesome-svg-core';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {
  ALERT_PARENT_ID,
  ALERT_STATUS_FIELD_AUTO,
  ALERT_TAGS_FIELD, ALERT_TIMESTAMP_FIELD, FALSE_POSITIVE_OBJECT
} from '../../../../../shared/constants/alert/alert-field.constant';
import {AUTOMATIC_REVIEW} from '../../../../../shared/constants/alert/alert-status.constant';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {DataNatureTypeEnum} from '../../../../../shared/enums/nature-data.enum';
import {ElasticDataService} from '../../../../../shared/services/elasticsearch/elastic-data.service';
import {UtmAlertType} from '../../../../../shared/types/alert/utm-alert.type';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {TimelineItem} from '../../../../../shared/types/utm-timeline-item';
import {sanitizeFilters} from '../../../../../shared/util/elastic-filter.util';
import {TimelineGroup, TimelineService} from './timeline.service';


@Component({
  selector: 'utm-timeline',
  templateUrl: './utm-timeline.component.html',
  styleUrls: ['./utm-timeline.component.scss']
})
export class UtmTimeLineComponent implements OnInit {

  @Input() alert: UtmAlertType;
  @Input() page = 0;
  @Input() pageSize = 100;
  @Input() total = 0;
  @Input() title = '';
  @Output() itemClick = new EventEmitter<TimelineItem>();


  sortBy = ALERT_TIMESTAMP_FIELD + ',desc';
  alerts: UtmAlertType[] = [];
  filters: ElasticFilterType[] = [
    {field: ALERT_STATUS_FIELD_AUTO, operator: ElasticOperatorsEnum.IS_NOT, value: AUTOMATIC_REVIEW},
    {field: ALERT_TAGS_FIELD, operator: ElasticOperatorsEnum.IS_NOT, value: FALSE_POSITIVE_OBJECT.tagName},
  ];
  loading = false;
  chartOption: any = {};
  intervalMs = 60 * 1000;
  groups: TimelineGroup[] = [];
  readonly Math = Math;

  ngOnInit(): void {
    this.filters.push({
      field: ALERT_PARENT_ID,
      operator: ElasticOperatorsEnum.IS,
      value: this.alert.id
    });
    this.loadData();
  }

  constructor(private timelineService: TimelineService,
              private elasticDataService: ElasticDataService,
              private utmToastService: UtmToastService,) {
  }

  buildChart() {
    const items = this.buildTimelineFromAlerts();
    const groups = this.timelineService.groupByInterval(items, this.intervalMs);
    this.groups = this.assignYOffsetToGroups(groups);
    /*console.log(this.groups.map(g => ({ ts: g.startTimestamp, yOffset: g.yOffset })));*/
    const seriesData = [];

    /*const cardWidth = 250;
    const cardHeight = 60;
    const spacing = 10;*/

    const cardWidth = 250;
    const cardHeight = 60;
    const spacing = 10;
    const baseOffset = 80;

    this.groups.forEach((group, index) => {
      const timestamps = group.items.map(i => new Date(i.startDate).getTime());
      group.startTimestamp = Math.floor(timestamps.reduce((sum, t) => sum + t, 0) / timestamps.length);

      const rep = group.items[0] || ({} as any); // representative item
      seriesData.push({
        value: [
          group.startTimestamp,                         // 0: timestamp (start of minute)
          0,                                            // 1: y coordinate (not used)
          rep.name || `Echoes`,                         // 2: representative name/title
          new Date(group.startTimestamp).toISOString(), // 3: formatted minute
          rep.iconUrl || 'assets/images/default-echo.png', // 4: icon url
          group.items.length,                             // 5: count of echoes
          index,
          group.yOffset || 0
        ],
        groupData: group.items,                         // full list for drill-down
      });
    });


    const allTimestamps = items.map(i => new Date(i.startDate).getTime());
    const minTimestamp = Math.min(...allTimestamps);
    const maxTimestamp = Math.max(...allTimestamps);
    const padding = (maxTimestamp - minTimestamp) * 0.1;

    this.chartOption = {
      title: {text: this.title, left: 'center', textStyle: {fontSize: 16, fontWeight: 'bold'}},
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
        min: minTimestamp - padding,
        max: maxTimestamp + padding,
        axisLabel: {formatter: (val: number) => new Date(val).toLocaleTimeString()},
        splitLine: {
          show: true,
          lineStyle: {
            type: 'dashed',
            color: '#ccc',
            width: 1
          }
        }
      },
      yAxis: {
        type: 'value',
        min: 0,
        max: (cardHeight + spacing) * this.groups.length + 100,
        show: false
      },
      dataZoom: [
        {type: 'slider', xAxisIndex: 0, start: 0, end: 100},
        {type: 'inside', xAxisIndex: 0, zoomLock: false}
      ],
      series: [
        {
          type: 'custom',
          data: seriesData,
          renderItem: (params: any, api: any) => {
            /*const ts = api.value(0);
            const coord = api.coord([ts, 0]);
            const chartWidth = params.coordSys.width;
            const offsetIndex = api.value(6) || 0;
            const totalOffset = (cardHeight + spacing) * offsetIndex + 80;
            const level = api.value(7) || 0;
            const cardOffset = level * (cardHeight + spacing) + 80;

            let x = coord[0] - cardWidth / 2;
            x = Math.max(0, Math.min(x, chartWidth - cardWidth));*/

            const ts = api.value(0);
            const coord = api.coord([ts, 0]);
            const chartWidth = params.coordSys.width;

            // center x for card, clamped inside chart
            let x = coord[0] - cardWidth / 2;
            x = Math.max(0, Math.min(x, chartWidth - cardWidth));

            const level = api.value(7) || 0; // stackLevel
            const levelPx = level * (cardHeight + spacing);
            const totalOffset = baseOffset + levelPx;

            return {
              type: 'group',
              children: this.buildChildren(api, coord, x, totalOffset, cardWidth, cardHeight)
            };
          },
          encode: {x: 0, y: 1}
        }
      ]
    };
  }

  onChartClick(event: any) {
    if (event.data && event.data.itemData) {
      this.itemClick.emit(event.data.itemData);
    }
  }

  truncateText(text: string, maxWidth: number) {
    const avgCharWidth = 7;
    const maxChars = Math.floor(maxWidth / avgCharWidth);
    return text.length > maxChars ? text.substring(0, maxChars - 3) + '...' : text;
  }

  loadData() {
    this.loading = true;
    this.elasticDataService.search(this.page, this.pageSize,
      100000000, DataNatureTypeEnum.ALERT,
      sanitizeFilters(this.filters), this.sortBy, true)
      .subscribe(
        (res: HttpResponse<any>) => {
          this.total = Number(res.headers.get('X-Total-Count'));
          this.alerts = res.body;
          this.loading = false;
          this.buildChart();
        },
        (res: HttpResponse<any>) => {
          this.utmToastService.showError('Error', 'An error occurred while listing the alerts. Please try again later.');
          this.loading = false;
        }
      );
  }

  prevPage() {
    this.page = this.page - 1;
    this.loadData();
  }

  nextPage() {
    this.page = this.page + 1;
    this.loadData();
  }

  buildTimelineFromAlerts(): TimelineItem[] {
    return this.alerts.map(cha => ({
      startDate: cha['@timestamp'],
      name: cha.name,
      metadata: cha,
      iconUrl: 'assets/icons/echoes/echoes_default.png'
    }));
  }

  private buildChildren(api: any, coord: number[], x: number, totalOffset: number, cardWidth: number, cardHeight: number): any[] {
    const children: any[] = [
      // Background card
      {
        type: 'rect',
        shape: {
          x,
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
          x: x + 5,
          y: coord[1] - totalOffset - cardHeight + 5,
          width: cardHeight - 10,
          height: cardHeight - 10
        }
      },

      // Title
      {
        type: 'text',
        style: {
          x: x + (cardHeight - 2.5) + 15,
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
          x: x + (cardHeight - 2.5) + 15,
          y: coord[1] - totalOffset - cardHeight + 25,
          text: api.value(3),
          textAlign: 'left',
          fill: '#666',
          fontSize: 12
        }
      }
    ];

    // Total echoes (solo si > 1)
    if (api.value(5) > 1) {
      children.push({
        type: 'text',
        style: {
          x: x + (cardHeight - 2.5) + 15,
          y: coord[1] - totalOffset - cardHeight + 42,
          text: `Total: ${api.value(5)} echoes`,
          textAlign: 'left',
          fill: '#444',
          fontSize: 12,
          fontWeight: 'bold'
        }
      });
    }

    // Connector line: conecta desde la base (coord[1]) hasta la tarjeta real (coord[1] - totalOffset)
    children.push({
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
    });

    return children;
  }

  private assignYOffsetToGroups(groups: TimelineGroup[]): TimelineGroup[] {
    const FIVE_MINUTE = 5 * 60 * 1000;
    const sorted = [...groups].sort((a, b) => a.startTimestamp - b.startTimestamp);

    for (let i = 0; i < sorted.length; i++) {
      const group = sorted[i];
      group.yOffset = 0; // default base

      for (let j = 0; j < i; j++) {
        const prev = sorted[j];
        const dx = group.startTimestamp - prev.startTimestamp;
        if (dx < FIVE_MINUTE) {
          group.yOffset = Math.max(group.yOffset, (prev.yOffset || 0) + 1);
        }
      }
    }

    return sorted;
  }
}
