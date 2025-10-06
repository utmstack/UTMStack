import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
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
import {AlertEchoesTimelineService, TimelineGroup} from './alert-echoes-timeline.service';


@Component({
  selector: 'app-alert-echoes-timeline',
  templateUrl: './alert-echoes-timeline.component.html',
  styleUrls: ['./alert-echoes-timeline.component.scss']
})
export class AlertEchoesTimelineComponent implements OnInit {

  @Input() alert: UtmAlertType;
  @Input() page = 0;
  @Input() pageSize = 100;
  @Input() total = 0;
  @Input() title = '';
  @Output() itemClick = new EventEmitter<TimelineItem>();

  chartInstance: any;

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

  constructor(private timelineService: AlertEchoesTimelineService,
              private elasticDataService: ElasticDataService,
              private utmToastService: UtmToastService, ) {
  }

  onChartInit(ec: any) {
    this.chartInstance = ec;
  }

  refreshChart() {
    if (this.chartInstance && this.chartOption) {
      this.chartInstance.clear(); // limpia todo el canvas
      this.chartInstance.setOption(this.chartOption, true); // redibuja
    }
  }

  buildChart() {

    const items = this.timelineService.buildTimelineFromAlerts(this.alerts);
    this.groups = this.timelineService.generateTimelineGroups(this.alerts, this.intervalMs);

    const seriesData = [];
    const cardHeight = 60;
    const spacing = 10;


    this.groups.forEach((group, index) => {
      const timestamps = group.items.map(i => new Date(i.startDate).getTime());
      group.startTimestamp = Math.floor(timestamps.reduce((sum, t) => sum + t, 0) / timestamps.length);

      const rep = group.items[0] || ({} as any); // representative item
      console.log('group', group, rep);
      seriesData.push({
        value: [
          group.startTimestamp,                         // 0: timestamp (start of minute)
          0,                                            // 1: y coordinate (not used)
          rep.name || `Echoes`,                         // 2: representative name/title
          new Date(group.startTimestamp).toISOString(), // 3: formatted minute
          rep.iconUrl || 'assets/images/default-echo.png', // 4: icon url
          group.items.length,                             // 5: count of echoes
          index,
          group.yOffset || 0,
          rep.metadata
        ],
        groupData: group.items,                         // full list for drill-down
      });
    });


    const allTimestamps = items.map(i => new Date(i.startDate).getTime());
    const minTimestamp = Math.min(...allTimestamps);
    const maxTimestamp = Math.max(...allTimestamps);
    const padding = (maxTimestamp - minTimestamp) * 0.1;

    const expand = (allTimestamps.length === 1) ? 30 * 60 * 1000 : (maxTimestamp - minTimestamp) * 0.1;

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
        min: minTimestamp - expand,
        max: maxTimestamp + expand,
        axisLabel: {
          formatter: (val: number) => {
            const d = new Date(val);
            const year = d.getUTCFullYear();
            const month = (d.getUTCMonth() + 1).toString().padStart(2, '0');
            const day = d.getUTCDate().toString().padStart(2, '0');
            const hours = d.getUTCHours().toString().padStart(2, '0');
            const minutes = d.getUTCMinutes().toString().padStart(2, '0');
            const seconds = d.getUTCSeconds().toString().padStart(2, '0');
            return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
          },
        },
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
          renderItem: (params: any, api: any) => this.timelineService.renderItem(params, api),
          encode: {x: 0, y: 1}
        }
      ]
    };
  }

  onChartClick(event: any) {
    if (event.data && event.data.value) {
      this.itemClick.emit(event.data.value[8] || {} as UtmAlertType);
    }
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
          this.refreshChart();
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

}
