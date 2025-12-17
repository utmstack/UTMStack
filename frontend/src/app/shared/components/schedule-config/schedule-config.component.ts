import {Component, EventEmitter, forwardRef, Input, OnInit, Output} from '@angular/core';
import {NG_VALUE_ACCESSOR} from '@angular/forms';
import {ScheduleConfig} from "./schedule-config.validator";

@Component({
  selector: 'app-utm-schedule-config',
  templateUrl: './schedule-config.component.html',
  styleUrls: ['./schedule-config.component.scss'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => ScheduleConfigComponent),
      multi: true
    }
  ]
})
export class ScheduleConfigComponent implements OnInit {
  @Input() initialConfig: ScheduleConfig = {
    days: [],
    startTime: { hour: 0, minute: 0 },
    endTime: { hour: 23, minute: 59 }
  };
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

  private onChange: (value: ScheduleConfig) => void = () => {};
  private onTouched: () => void = () => {};

  ngOnInit() {
    this.config = { ...this.initialConfig };
  }

  writeValue(value: ScheduleConfig): void {
    if (value) {
      this.config = {
        days: [...value.days],
        startTime: this.parseTimeString(value.startTime),
        endTime: this.parseTimeString(value.endTime)
      };
    }
  }

  registerOnChange(fn: (value: ScheduleConfig) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState?(isDisabled: boolean): void {

  }

  private emitChange() {
    const newConfig = { ...this.config };
    this.scheduleChange.emit(newConfig);
    this.onChange({
      days: newConfig.days,
      startTime: this.formatTime(newConfig.startTime),
      endTime: this.formatTime(newConfig.endTime)
    });
    this.onTouched();
  }

  getDayNames(): string {
    const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
    return this.config.days
      .map(num => dayNames[num])
      .join(', ');
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

  formatTime(time: any): string {
    if (!time) { return ''; }
    const h = time.hour.toString().padStart(2, '0');
    const m = time.minute.toString().padStart(2, '0');
    return `${h}:${m}`;
  }

  private parseTimeString(timeStr: string): {hour: number, minute: number} {
    const [h, m] = timeStr.split(':').map(Number);
    return { hour: h, minute: m };
  }


}
