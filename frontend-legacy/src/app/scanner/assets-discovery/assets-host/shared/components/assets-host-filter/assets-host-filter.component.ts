import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {TimeFilterType} from '../../../../../../shared/types/time-filter.type';
import {SEVERITY_VALUE} from '../../../../../shared/const/task-status.const';
import {ScannerSeverityEnum} from '../../../../../shared/enums/scanner-severity.enum';

@Component({
  selector: 'app-assets-host-filter',
  templateUrl: './assets-host-filter.component.html',
  styleUrls: ['./assets-host-filter.component.scss']
})
export class AssetsHostFilterComponent implements OnInit {
  filter = {};
  @Output() filterHostChange = new EventEmitter<any>();
  @Input() severityFilter: ScannerSeverityEnum;
  @Input() hostname: string;
  @Input() hostSo: string;
  severities = SEVERITY_VALUE;
  severity: ScannerSeverityEnum = ScannerSeverityEnum.ALL;

  constructor() {
  }

  ngOnInit() {
    if (this.severityFilter) {
      this.severity = this.severityFilter;
      this.searchBySeverity(null);
    }
    if (this.hostname) {
      this.filter['name.contains'] = this.hostname;
      this.searchByHostProperty();
    }

    if (this.hostSo) {
      this.filter['hostOs.contains'] = this.hostSo;
      this.searchByHostProperty();
    }
  }

  searchByHostProperty() {
    setTimeout(() => this.filterHostChange.emit(this.filter), 2000);
  }

  onChangeTime($event: TimeFilterType) {
    this.filter['created.greaterThan'] = $event.timeFrom.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.filter['created.lessThan'] = $event.timeTo.substring(0, $event.timeFrom.lastIndexOf(':'));
    this.filterHostChange.emit(this.filter);
  }

  searchBySeverity($event: any) {
    if (this.severity !== ScannerSeverityEnum.ALL) {
      switch (this.severity) {
        case ScannerSeverityEnum.LOG:
          this.filter['hostSeverity.equals'] = 0;
          this.filter['hostSeverity.greaterThan'] = undefined;
          this.filter['hostSeverity.lessThan'] = undefined;
          this.filterHostChange.emit(this.filter);
          break;
        case ScannerSeverityEnum.LOW:
          this.filter['hostSeverity.equals'] = undefined;
          this.filter['hostSeverity.greaterThan'] = 0;
          this.filter['hostSeverity.lessThan'] = 4;
          this.filterHostChange.emit(this.filter);
          break;
        case ScannerSeverityEnum.MEDIUM:
          this.filter['hostSeverity.equals'] = undefined;
          this.filter['hostSeverity.greaterThan'] = 3.9;
          this.filter['hostSeverity.lessThan'] = 7;
          this.filterHostChange.emit(this.filter);
          break;
        case ScannerSeverityEnum.HIGH:
          this.filter['hostSeverity.equals'] = undefined;
          this.filter['hostSeverity.greaterThan'] = 6.9;
          this.filter['hostSeverity.lessThan'] = 11;
          this.filterHostChange.emit(this.filter);
          break;
        default:
          this.filter['hostSeverity.equals'] = '""';
          this.filterHostChange.emit(this.filter);
      }
    } else {
      this.filter['hostSeverity.greaterThan'] = undefined;
      this.filter['hostSeverity.lessThan'] = undefined;
      this.filter['hostSeverity.equals'] = undefined;
      this.filterHostChange.emit(this.filter);
    }
  }
}
