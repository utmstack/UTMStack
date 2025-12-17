import { Component, OnInit, Output, EventEmitter, Input } from '@angular/core';

interface ScheduleConfig {
  days: number[];
  startTime: string;
  endTime: string;
}

@Component({
  selector: 'app-utm-schedule-config',
  templateUrl: './utm-schedule-config.component.html',
  styleUrls: ['./utm-schedule-config.component.scss']
})
export class UtmScheduleConfigComponent implements OnInit {
  @Input() initialConfig: ScheduleConfig = { days: [], startTime: '22:00', endTime: '06:00' };
  @Output() scheduleChange = new EventEmitter<ScheduleConfig>();

  days = [
    { number: 1, name: 'Monday' },
    { number: 2, name: 'Tuesday' },
    { number: 3, name: 'Wednesday' },
    { number: 4, name: 'Thursday' },
    { number: 5, name: 'Friday' },
    { number: 6, name: 'Saturday' },
    { number: 0, name: 'Sunday' }
  ];

  config: ScheduleConfig = { ...this.initialConfig };

  ngOnInit() {
    this.config = { ...this.initialConfig };
  }

  toggleDay(dayNumber: number) {
    const index = this.config.days.indexOf(dayNumber);
    if (index > -1) {
      this.config.days.splice(index, 1);
    } else {
      this.config.days.push(dayNumber);
    }
    this.config.days.sort();
    this.emitChange();
  }

  isDaySelected(dayNumber: number): boolean {
    return this.config.days.includes(dayNumber);
  }

  onTimeChange() {
    this.emitChange();
  }

  private emitChange() {
    this.scheduleChange.emit({ ...this.config });
  }
}
