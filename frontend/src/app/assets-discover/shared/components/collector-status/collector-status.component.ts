import {ChangeDetectionStrategy, Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';
import {UtmModuleCollectorType} from '../../../../app-module/shared/type/utm-module-collector.type';

@Component({
  selector: 'app-collector-status',
  templateUrl: './collector-status.component.html',
  styleUrls: ['./collector-status.component.scss']
})
export class CollectorStatusComponent implements OnInit {
  @Input() collector: UtmModuleCollectorType;
  statusClass: string;
  statusLabel: string;

  constructor() {
  }

  ngOnInit() {
    this.statusClass = this.getClass();
    this.statusLabel = this.getLabel();
  }

  getClass(): string {
    if (this.collector.status === 'OFFLINE') {
      return 'text-danger';
    } else   {
      return 'text-success';
    }
  }

  getLabel(): string {
    if (this.collector.status === 'OFFLINE') {
      return 'Disconnected';
    } else  {
      return 'Connected';
    }
  }

}
