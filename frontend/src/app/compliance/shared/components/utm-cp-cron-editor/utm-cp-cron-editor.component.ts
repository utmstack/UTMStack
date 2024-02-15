import {Component, forwardRef} from '@angular/core';
import {ControlValueAccessor, NG_VALUE_ACCESSOR} from '@angular/forms';
import {DaysOfWeek, MonthsOfYear, TimeFrequency} from './models/time-frequency';

const dailyFrequencies: number[] = Array.from({length: 29}, (_, index) => index + 1);
const monthlyFrequencies: number[] = Array.from({length: 12}, (_, index) => index + 1);

export const CUSTOM_CONTROL_VALUE_ACCESSOR: any = {
  provide: NG_VALUE_ACCESSOR,
  useExisting: forwardRef(() => UtmCpCronEditorComponent),
  multi: true,
};

@Component({
  selector: 'app-utm-cp-cron-editor',
  templateUrl: './utm-cp-cron-editor.component.html',
  styleUrls: ['./utm-cp-cron-editor.component.scss'],
  providers: [CUSTOM_CONTROL_VALUE_ACCESSOR],
})
export class UtmCpCronEditorComponent implements ControlValueAccessor {

  disabled = false;

   TimeFrequency = TimeFrequency;
   DaysOfWeek = DaysOfWeek;

  months: { name: string, value: string }[] = Object.keys(MonthsOfYear).map(key => ({
    name: key,
    value: MonthsOfYear[key as keyof typeof MonthsOfYear]
  }));

  timeFrequency: TimeFrequency = TimeFrequency.Daily;
  dailyFrequency = 1;
  monthlyFrequency = 1;
  yearlyFrequency = this.months[0].value;

  startDate = new Date();
  endDate: any;
  time = {hour: 0, minute: 0};
  days: string[] = [];
  cmdCron: string;
  onChange: (value: string) => void = () => {};
  onTouched = () => {};

  set cronSentence(cmd: string) {
    this.cmdCron = cmd;
  }

  get frequenciesByType() {
    if (this.timeFrequency === TimeFrequency.Daily) {
      return dailyFrequencies;
    }
    return monthlyFrequencies;
  }

  get monthlyValue() {
    if (this.timeFrequency === TimeFrequency.Yearly) {
      return this.yearlyFrequency;
    } else if (this.timeFrequency === TimeFrequency.Weekly || this.timeFrequency === TimeFrequency.Daily) {
      return '*';
    }
    return `*/${this.monthlyFrequency}`;
  }

  getTimeEnumValues(obj: any): string[] {
    return Object.values(obj) as string[];
  }

  getTime(position: number): string {
    /*const formatTime = this.convertTo24Format(this.time);*/
    const time = position === 1 ? this.time.minute : this.time.hour;
    return this.time.hour === 0 && this.time.minute === 0 ? '0' : time.toString();
  }

  getDay() {
    return this.timeFrequency === TimeFrequency.Daily && this.dailyFrequency !== 0 ? `*/${this.dailyFrequency}` : '*';
  }

  getDays() {
    return this.days.length > 0 ? this.days.join(',') : '*';
  }

  isSelected(day: string): boolean {
    return this.days.includes(this.getIndexDay(day));
  }

  setDays(day: string) {
    if (this.isSelected(day)) {
      this.days.splice(this.days.indexOf(day), 1);
    } else {
      this.days.push(this.getIndexDay(day));
    }
    this.emitChange();
  }

  onChangeFrequency() {
    if (this.timeFrequency === TimeFrequency.Weekly) {
      this.days = ['0'];
    }
    if (this.timeFrequency !== TimeFrequency.Weekly) {
      this.days = [];
    }
    if (this.timeFrequency === TimeFrequency.Weekly || this.timeFrequency === TimeFrequency.Yearly) {
      this.dailyFrequency = 0;
    }
    this.emitChange();
  }
  registerOnChange(fn: any): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn;
  }

  writeValue(cron: string): void {
    if (cron != null && cron !== '' && cron !== undefined) {
      const cronParts = cron.split(' ');
      if (cronParts[1] !== '*') {
        this.time = {
          hour: Number(cronParts[2]),
          minute: Number(cronParts[1])
        };
      }

      if (cronParts[3] !== '*') {
        this.timeFrequency = TimeFrequency.Daily;
        this.dailyFrequency = Number(cronParts[3].split('*/')[1]);
      }

      if (cronParts[4] !== '*') {
        if (Number(cronParts[4])) {
          this.timeFrequency = TimeFrequency.Yearly;
          this.yearlyFrequency = this.months[Number(cronParts[4]) - 1].value;
        } else {
          this.timeFrequency = TimeFrequency.Monthly;
          this.monthlyFrequency = Number(cronParts[4].split('*/')[1]);
        }
      }

      if (cronParts[5] !== '*') {
        this.timeFrequency = TimeFrequency.Weekly;
        this.days = [];
        cronParts[5].split(',').forEach(value => {
          this.days.push(value);
        });
      }

      this.cronSentence = cron;
    }
  }
  getIndexDay(day: string) {
    return Object.values(DaysOfWeek).findIndex(value => value === day).toString();
  }
  emitChange() {
    this.cmdCron = `0 ${this.getTime(1)} ${this.getTime(0)} ${this.getDay()} ${this.monthlyValue} ${this.getDays()}`;
    this.onChange(this.cmdCron);
  }
}
