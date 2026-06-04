import {Component, forwardRef, OnInit} from '@angular/core';
import {ControlValueAccessor, FormBuilder, FormGroup, NG_VALUE_ACCESSOR} from '@angular/forms';
import {NgbTimeStruct} from '@ng-bootstrap/ng-bootstrap';
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
export class UtmCpCronEditorComponent implements ControlValueAccessor, OnInit {
  form: FormGroup;
  disabled = false;

  TimeFrequency = TimeFrequency;
  DaysOfWeek = DaysOfWeek;
  months: { name: string, value: string }[] = Object.keys(MonthsOfYear).map(key => ({
    name: key,
    value: MonthsOfYear[key as keyof typeof MonthsOfYear]
  }));

  cmdCron = '';
  onChange: (value: string) => void = () => {};
  onTouched = () => {};

  constructor(private fb: FormBuilder) {}

  ngOnInit() {
    this.form = this.fb.group({
      timeFrequency: [TimeFrequency.Daily],
      dailyFrequency: [1],
      monthlyFrequency: [1],
      yearlyFrequency: [this.months[0].value],
      time: this.fb.control({ hour: 0, minute: 0, second: 0 } as NgbTimeStruct),
      days: [[]]
    });

    this.form.valueChanges.subscribe(() => {
      this.emitChange();
    });
  }

  get frequenciesByType() {
    return this.form.get('timeFrequency').value === TimeFrequency.Daily ? dailyFrequencies : monthlyFrequencies;
  }

  get monthlyValue() {
    const freq = this.form.get('timeFrequency').value;
    if (freq === TimeFrequency.Yearly) {
      return this.form.get('yearlyFrequency').value;
    } else if (freq === TimeFrequency.Weekly || freq === TimeFrequency.Daily) {
      return '*';
    }
    return `*/${this.form.get('monthlyFrequency').value}`;
  }

  getTimeEnumValues(obj: any): string[] {
    return Object.values(obj) as string[];
  }

  getTime(position: number): string {
    const time = this.form.get('time').value;
    return (time.hour === 0 && time.minute === 0) ? '0' : (position === 1 ? time.minute : time.hour).toString();
  }

  getDay(): string {
    const tf = this.form.get('timeFrequency').value;
    const df = this.form.get('dailyFrequency').value;
    return tf === TimeFrequency.Daily && df !== 0 ? `*/${df}` : '*';
  }

  getDays(): string {
    const days = this.form.get('days').value;
    return days.length > 0 ? days.join(',') : '*';
  }

  isSelected(day: string): boolean {
    return this.form.get('days').value.includes(this.getIndexDay(day));
  }

  setDays(day: string) {
    const daysControl = this.form.get('days');
    const currentDays = daysControl.value || [];
    const index = currentDays.indexOf(this.getIndexDay(day));
    if (index >= 0) {
      currentDays.splice(index, 1);
    } else {
      currentDays.push(this.getIndexDay(day));
    }
    daysControl.setValue([...currentDays]);
  }

  onChangeFrequency() {
    const tf = this.form.get('timeFrequency').value;
    if (tf === TimeFrequency.Weekly) {
      this.form.get('days').setValue(['0']);
    } else {
      this.form.get('days').setValue([]);
    }
    if (tf === TimeFrequency.Weekly || tf === TimeFrequency.Yearly) {
      this.form.get('dailyFrequency').setValue(0);
    }
  }

  getIndexDay(day: string): string {
    return Object.values(DaysOfWeek).findIndex(value => value === day).toString();
  }

  emitChange() {
    console.log('emitChange');
    const cron = `0 ${this.getTime(1)} ${this.getTime(0)} ${this.getDay()} ${this.monthlyValue} ${this.getDays()}`;
    this.cmdCron = cron;
    this.onChange(cron);
  }

  writeValue(cron: string): void {
    if (!cron) { return; }
    const parts = cron.split(' ');
    const [_, minute, hour, day, month, weekDays] = parts;

    const time = { hour: Number(hour), minute: Number(minute), second: 0 };
    let timeFrequency = TimeFrequency.Daily;
    let dailyFrequency = 1;
    let monthlyFrequency = 1;
    let yearlyFrequency = this.months[0].value;
    let days: string[] = [];

    if (day !== '*') {
      timeFrequency = TimeFrequency.Daily;
      dailyFrequency = Number(day.split('*/')[1]);
    }
    if (month !== '*') {
      if (!isNaN(Number(month))) {
        timeFrequency = TimeFrequency.Yearly;
        yearlyFrequency = this.months[Number(month) - 1].value;
      } else {
        timeFrequency = TimeFrequency.Monthly;
        monthlyFrequency = Number(month.split('*/')[1]);
      }
    }
    if (weekDays !== '*') {
      timeFrequency = TimeFrequency.Weekly;
      days = weekDays.split(',');
    }

    this.form.patchValue({
      timeFrequency,
      dailyFrequency,
      monthlyFrequency,
      yearlyFrequency,
      time,
      days,
    }, { emitEvent: false });

    this.cmdCron = cron;
  }

  registerOnChange(fn: any): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled = isDisabled;
    isDisabled ? this.form.disable() : this.form.enable();
  }

  get timeFrequency() {
    return this.form.get('timeFrequency');
  }

}
