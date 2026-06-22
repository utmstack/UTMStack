import {Component, EventEmitter, Input, OnChanges, Output, SimpleChanges} from '@angular/core';
import {AlertStatusSerie, CountAlertsByStatusEntry} from '../../domain/count-alerts-by-status.model';

interface NavAction {
  readonly label: string;
  readonly iconClass: string;
  readonly route: string;
}

interface PieDatum {
  name: string;
  value: number;
  itemStyle: { color: string };
}

interface PieChartOption {
  tooltip: { trigger: string; formatter: string };
  legend: {
    orient: 'horizontal';
    bottom: number;
    data: string[];
    textStyle: { fontSize: number };
  };
  title: {
    text: string;
    left: string;
    top: number;
    textStyle: { fontSize: number; fontWeight: string };
  };
  series: Array<{
    name: string;
    type: 'pie';
    radius: [string, string];
    center: [string, string];
    data: PieDatum[];
    label: { show: boolean };
    labelLine: { show: boolean };
  }>;
}

const STATUS_COLORS: Record<AlertStatusSerie, string> = {
  'Open': '#dc3545',
  'In review': '#ffc107',
  'Completed': '#28a745'
};

const ACTIONS: ReadonlyArray<NavAction> = [
  {label: 'Log Explorer', iconClass: 'icon-search4', route: '/discover/log-analyzer'},
  {label: 'Alert Management', iconClass: 'icon-bell2', route: '/data/alert/view'},
  {label: 'SOAR Flows', iconClass: 'icon-stack-play', route: '/soar/flows'},
  {label: 'Data Sources', iconClass: 'icon-database', route: '/data-sources'}
];

@Component({
  selector: 'app-instance-overview-card',
  templateUrl: './instance-overview-card.component.html',
  styleUrls: ['./instance-overview-card.component.scss']
})
export class InstanceOverviewCardComponent implements OnChanges {
  @Input() entry!: CountAlertsByStatusEntry;
  @Output() select = new EventEmitter<string>();

  readonly actions: ReadonlyArray<NavAction> = ACTIONS;
  pieOption: PieChartOption | null = null;
  totalAlerts = 0;

  ngOnChanges(changes: SimpleChanges): void {
    if (changes.entry && this.entry) {
      this.totalAlerts = this.entry.data.reduce((sum, bucket) => sum + bucket.value, 0);
      this.pieOption = this.totalAlerts > 0 ? this.buildOption(this.entry) : null;
    }
  }

  get hasData(): boolean {
    return this.totalAlerts > 0;
  }

  onSelect(route: string): void {
    this.select.emit(route);
  }

  private buildOption(entry: CountAlertsByStatusEntry): PieChartOption {
    const data: PieDatum[] = entry.data.map(bucket => ({
      name: bucket.serie,
      value: bucket.value,
      itemStyle: {color: STATUS_COLORS[bucket.serie] || '#6c757d'}
    }));
    return {
      tooltip: {trigger: 'item', formatter: '{b}: {c} ({d}%)'},
      legend: {
        orient: 'horizontal',
        bottom: 0,
        data: entry.data.map(bucket => bucket.serie),
        textStyle: {fontSize: 11}
      },
      title: {
        text: entry.instanceName,
        left: 'center',
        top: 8,
        textStyle: {fontSize: 14, fontWeight: '600'}
      },
      series: [{
        name: 'Alerts',
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '52%'],
        data,
        label: {show: false},
        labelLine: {show: false}
      }]
    };
  }
}
