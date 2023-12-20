import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {TimeFilterType} from '../../../../../../shared/types/time-filter.type';
import {SEVERITY_VALUE, STATUS_TASK} from '../../../../../shared/const/task-status.const';
import {ScannerSeverityEnum} from '../../../../../shared/enums/scanner-severity.enum';

@Component({
    selector: 'app-task-filter',
    templateUrl: './task-filter.component.html',
    styleUrls: ['./task-filter.component.scss']
})
export class TaskFilterComponent implements OnInit {
    @Output() taskFilterChange = new EventEmitter<any>();
    filter = {};
    status = STATUS_TASK;
    severities = SEVERITY_VALUE;
    severity = 'All';

    constructor() {
    }

    ngOnInit() {
        this.filter['status.equals'] = ScannerSeverityEnum.ALL;
    }

    searchByText() {
        setTimeout(() => this.taskFilterChange.emit(this.filter), 2000);
    }

    searchByCreation($event: TimeFilterType) {
        this.filter['created.greaterThan'] = $event.timeFrom.substring(0, $event.timeFrom.lastIndexOf(':'));
        this.filter['created.lessThan'] = $event.timeTo.substring(0, $event.timeFrom.lastIndexOf(':'));
        this.taskFilterChange.emit(this.filter);
    }

    searchByStatus(stat: Event) {
        if (this.filter['status.equals'] !== ScannerSeverityEnum.ALL) {
            this.taskFilterChange.emit(this.filter);
        } else {
            this.filter['status.equals'] = undefined;
            this.taskFilterChange.emit(this.filter);
        }
    }

    searchBySeverity($event: Event) {
        if (this.severity !== ScannerSeverityEnum.ALL) {
            switch (this.severity) {
                case ScannerSeverityEnum.LOG:
                    this.filter['severity.greaterThan'] = 0;
                    this.filter['severity.lessThan'] = 0;
                    this.taskFilterChange.emit(this.filter);
                    break;
                case ScannerSeverityEnum.LOW:
                    this.filter['severity.greaterThan'] = 0.1;
                    this.filter['severity.lessThan'] = 3.9;
                    this.taskFilterChange.emit(this.filter);
                    break;
                case ScannerSeverityEnum.MEDIUM:
                    this.filter['severity.greaterThan'] = 4;
                    this.filter['severity.lessThan'] = 6.9;
                    this.taskFilterChange.emit(this.filter);
                    break;
                case ScannerSeverityEnum.HIGH:
                    this.filter['severity.greaterThan'] = 7;
                    this.filter['severity.lessThan'] = 999999;
                    this.taskFilterChange.emit(this.filter);
                    break;
                case ScannerSeverityEnum.UNKNOWN:
                    this.filter['severity.equals'] = '""';
                    this.taskFilterChange.emit(this.filter);
                    break;
            }
        } else {
            this.filter['severity.greaterThan'] = undefined;
            this.filter['severity.lessThan'] = undefined;
            this.taskFilterChange.emit(this.filter);
        }
    }
}
