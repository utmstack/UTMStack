import {Component, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmHealthService} from '../health-checks.service';

@Component({
  selector: 'app-health-detail',
  templateUrl: './health-detail.component.html',
  styleUrls: ['./health-detail.component.scss']
})
export class HealthDetailComponent implements OnInit {
  currentHealth: any;

  constructor(private healthService: UtmHealthService, public activeModal: NgbActiveModal) {
  }

  baseName(name) {
    return this.healthService.getBaseName(name);
  }

  subSystemName(name) {
    return this.healthService.getSubSystemName(name);
  }


  getProgressValue(): number {
    return (100 * this.getDiskSpaceByKey('free')) / this.getDiskSpaceByKey('total');
  }

  getDiskSpaceByKey(key: string): number {
    if (this.currentHealth) {
      return (this.currentHealth.details.details[key] / 1073741824);
    } else {
      return 0;
    }
  }


  readableValue(value: number) {
    if (this.currentHealth.name === 'diskSpace') {
      // Should display storage space in an human readable unit
      const val = value / 1073741824;
      if (val > 1) {
        // Value
        return val.toFixed(2) + ' GB';
      } else {
        return (value / 1048576).toFixed(2) + ' MB';
      }
    }

    if (typeof value === 'object') {
      return JSON.stringify(value);
    } else {
      return value.toString();
    }
  }

  ngOnInit(): void {
  }

  extractHealth(): { name: string, value: string | number }[] {
    const health: { name: string, value: string | number }[] = [];
    Object.keys(this.currentHealth.details.details).forEach(key => {
      health.push({
        name: key === 'hello' ? 'ping' : key,
        value: this.currentHealth.details.details[key]
      });
    });
    return health;
  }
}
