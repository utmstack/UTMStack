import {Component, EventEmitter, Input, OnChanges, Output, SimpleChanges} from '@angular/core';
import {
  AlertSeveritySerie,
  CountAlertsBySeverityEntry
} from '../../domain/count-alerts-by-severity.model';
import {AlertStatusSerie, CountAlertsByStatusEntry} from '../../domain/count-alerts-by-status.model';
import {FEDERATION_NAV_ACTIONS, NavAction} from '../../domain/federation-nav-actions';
import {SYSTEM_MENU_ICONS_PATH} from '../../../shared/constants/menu_icons.constants';

interface MeterDatum {
  label: string;
  value: number;
  color: string;
}

const STATUS_COLORS: Record<AlertStatusSerie, string> = {
  'Open': '#28a745',
  'In review': '#ffc107',
  'Completed': '#17a2b8'
};

const SEVERITY_COLORS: Record<AlertSeveritySerie, string> = {
  'High': '#dc3545',
  'Medium': '#ffc107',
  'Low': '#0a6ebd'
};

const SEVERITY_ORDER: ReadonlyArray<AlertSeveritySerie> = ['High', 'Medium', 'Low'];
const STATUS_ORDER: ReadonlyArray<AlertStatusSerie> = ['Open', 'In review', 'Completed'];

@Component({
  selector: 'app-instance-overview-card',
  templateUrl: './instance-overview-card.component.html',
  styleUrls: ['./instance-overview-card.component.scss']
})
export class InstanceOverviewCardComponent implements OnChanges {
  @Input() entry!: CountAlertsByStatusEntry;
  @Input() severityEntry: CountAlertsBySeverityEntry | null = null;
  @Output() select = new EventEmitter<string>();
  iconPath=SYSTEM_MENU_ICONS_PATH

  readonly actions: ReadonlyArray<NavAction> = FEDERATION_NAV_ACTIONS;
  severityMeters: MeterDatum[] = [];
  statusMeters: MeterDatum[] = [];
  severityMax = 0;
  statusMax = 0;
  totalAlerts = 0;
  totalSeverity = 0;

  ngOnChanges(changes: SimpleChanges): void {
    if (changes.entry && this.entry) {
      const buckets = this.entry.data || [];
      this.totalAlerts = buckets.reduce((sum, bucket) => sum + bucket.value, 0);
      this.statusMeters = STATUS_ORDER.map(serie => {
        const bucket = buckets.find(b => b.serie === serie);
        return {
          label: serie,
          value: bucket ? bucket.value : 0,
          color: STATUS_COLORS[serie]
        };
      });
      this.statusMax = this.statusMeters.reduce((m, d) => m+d.value, 0);
    }
    if (changes.severityEntry) {
      const buckets = this.severityEntry ? this.severityEntry.data.value : [];
      this.totalSeverity = buckets.reduce((sum, bucket) => sum + bucket.value, 0);
      this.severityMeters = SEVERITY_ORDER.map(serie => {
        const bucket = buckets.find(b => b.name === serie);
        return {
          label: serie,
          value: bucket ? bucket.value : 0,
          color: SEVERITY_COLORS[serie]
        };
      });
      this.severityMax = this.severityMeters.reduce((m, d) => Math.max(m, d.value), 0);
    }
  }

  get hasData(): boolean {
    return this.totalAlerts > 0 || this.totalSeverity > 0;
  }

  onSelect(route: string): void {
    this.select.emit(route);
  }
}
