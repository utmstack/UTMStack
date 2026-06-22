import {Component, EventEmitter, Input, OnChanges, Output, SimpleChanges} from '@angular/core';
import {
  AlertSeveritySerie,
  CountAlertsBySeverityBucket,
  CountAlertsBySeverityEntry
} from '../../domain/count-alerts-by-severity.model';
import {AlertStatusSerie, CountAlertsByStatusEntry} from '../../domain/count-alerts-by-status.model';
import {FEDERATION_NAV_ACTIONS, NavAction} from '../../domain/federation-nav-actions';

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

const SEVERITY_COLORS: Record<AlertSeveritySerie, string> = {
  'High': '#dc3545',
  'Medium': '#ffc107',
  'Low': '#0a6ebd'
};

const SEVERITY_ORDER: ReadonlyArray<AlertSeveritySerie> = ['High', 'Medium', 'Low'];

@Component({
  selector: 'app-instance-overview-card',
  templateUrl: './instance-overview-card.component.html',
  styleUrls: ['./instance-overview-card.component.scss']
})
export class InstanceOverviewCardComponent implements OnChanges {
  @Input() entry!: CountAlertsByStatusEntry;
  @Input() severityEntry: CountAlertsBySeverityEntry | null = null;
  @Output() select = new EventEmitter<string>();

  readonly actions: ReadonlyArray<NavAction> = FEDERATION_NAV_ACTIONS;
  statusOption: PieChartOption | null = null;
  severityOption: PieChartOption | null = null;
  totalAlerts = 0;
  totalSeverity = 0;

  ngOnChanges(changes: SimpleChanges): void {
    if (changes.entry && this.entry) {
      console.log(this.entry)
      this.totalAlerts = this.entry.data.reduce((sum, bucket) => sum + bucket.value, 0);
      this.statusOption = this.totalAlerts > 0 ? this.buildStatusOption(this.entry) : null;
    }
    if (changes.severityEntry) {
      const buckets = this.severityEntry ? this.severityEntry.data : [];
      this.totalSeverity = buckets.reduce((sum, bucket) => sum + bucket.value, 0);
      this.severityOption = this.totalSeverity > 0
        ? this.buildSeverityOption(buckets)
        : null;
    }
  }

  get hasData(): boolean {
    return this.totalAlerts > 0 || this.totalSeverity > 0;
  }

  onSelect(route: string): void {
    this.select.emit(route);
  }

  private buildStatusOption(entry: CountAlertsByStatusEntry): PieChartOption {
    const data: PieDatum[] = entry.data.map(bucket => ({
      name: bucket.serie,
      value: bucket.value,
      itemStyle: {color: STATUS_COLORS[bucket.serie] || '#6c757d'}
    }));
    return this.buildPieOption('Status', data, entry.data.map(bucket => bucket.serie));
  }

  private buildSeverityOption(buckets: CountAlertsBySeverityBucket[]): PieChartOption {
    const ordered = SEVERITY_ORDER
      .map(serie => buckets.find(bucket => bucket.serie === serie))
      .filter((bucket): bucket is CountAlertsBySeverityBucket => !!bucket && bucket.value > 0);
    const data: PieDatum[] = ordered.map(bucket => ({
      name: bucket.serie,
      value: bucket.value,
      itemStyle: {color: SEVERITY_COLORS[bucket.serie] || '#6c757d'}
    }));
    return this.buildPieOption('Severity', data, ordered.map(bucket => bucket.serie));
  }

  private buildPieOption(title: string, data: PieDatum[], legend: string[]): PieChartOption {
    return {
      tooltip: {trigger: 'item', formatter: '{b}: {c} ({d}%)'},
      legend: {
        orient: 'horizontal',
        bottom: 0,
        data: legend,
        textStyle: {fontSize: 11}
      },
      title: {
        text: title,
        left: 'center',
        top: 4,
        textStyle: {fontSize: 12, fontWeight: '600'}
      },
      series: [{
        name: title,
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['50%', '55%'],
        data,
        label: {show: false},
        labelLine: {show: false}
      }]
    };
  }
}
