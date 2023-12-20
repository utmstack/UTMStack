import {Injectable} from '@angular/core';
import {TimeFrequency} from '../components/utm-cp-cron-editor/models/time-frequency';

@Injectable({
  providedIn: 'root'
})

export class CronDescriptionGeneratorService {

  private static dayOfWeekMap: Record<string, string> = {
    '0': 'Sunday',
    '1': 'Monday',
    '2': 'Tuesday',
    '3': 'Wednesday',
    '4': 'Thursday',
    '5': 'Friday',
    '6': 'Saturday'
  };

  private static monthMap: Record<string, string> = {
    '1': 'January',
    '2': 'February',
    '3': 'March',
    '4': 'April',
    '5': 'May',
    '6': 'June',
    '7': 'July',
    '8': 'August',
    '9': 'September',
    '10': 'October',
    '11': 'November',
    '12': 'December'
  };

  private cronFields: any;
  private timeFrequency: TimeFrequency;
  private frequency: string;


  private mapDayOfWeek(dayOfWeek: string): string {
    return CronDescriptionGeneratorService.dayOfWeekMap[dayOfWeek] || dayOfWeek;
  }

  private mapMonth(month: string): string {
    return CronDescriptionGeneratorService.monthMap[month] || month;
  }

  private checkCronString(cronString: string) {
    const fields = cronString.split(' ');

    if (fields.length !== 6) {
      throw new Error('Invalid cron');
    }

    this.cronFields = {
      seconds: fields[0],
      minutes: fields[1],
      hours: fields[2],
      dayOfMonth: fields[3],
      month: fields[4],
      dayOfWeek: fields[5]
    };

    if (fields[3] !== '*') {
      const day = fields[3].split('*/');
      if (Number(day[1])) {
        this.timeFrequency = TimeFrequency.Daily;
        this.frequency = day[1];
      }
    }

    if (fields[4] !== '*') {
      if (Number(fields[4])) {
        this.timeFrequency = TimeFrequency.Yearly;
        this.frequency = fields[4];
      } else {
        const day = fields[4].split('*/');
        if (Number(day[1])) {
          this.timeFrequency = TimeFrequency.Monthly;
          this.frequency = day[1];
        }
      }
    }

    if (fields[5] !== '*') {
      if (Number(fields[5]) || fields[5].split(',').length > 0) {
        this.timeFrequency = TimeFrequency.Weekly;
      }
    }

  }

  public getDescription(cronString: string): string {

    this.checkCronString(cronString);
    return this.getResults();

  }

  getHour() {
    if (this.cronFields.minutes === '*') {
      return 'at every minute';
    } else if (Number(this.cronFields.minutes) && Number(this.cronFields.hours)) {
      return `at ${this.cronFields.hours}:${this.cronFields.minutes}`;
    } else {
      return '';
    }
  }

  getDaysOfWeek() {
    if (this.cronFields.dayOfWeek === '*') {
      return '';
    } else if (Number(this.cronFields.dayOfWeek)) {
      return `on ${CronDescriptionGeneratorService.dayOfWeekMap[this.cronFields.dayOfWeek]}`;
    } else {
      const days = this.cronFields.dayOfWeek.split(',');
      return `on ${days.map((day: string) => CronDescriptionGeneratorService.dayOfWeekMap[day]).join(',')}`;
    }
  }
  private getResults() {

    let freq = this.timeFrequency.toLowerCase();
    const hours = this.getHour();

    if (this.timeFrequency === TimeFrequency.Daily || this.timeFrequency === TimeFrequency.Monthly) {
      if (Number(this.frequency) > 1) {
        freq = this.timeFrequency.toLowerCase().concat('s');
      }

      return `Runs every ${this.frequency} ${freq} ${hours}`;

    } else if (this.timeFrequency === TimeFrequency.Weekly) {
      return `Runs weekly ${this.getDaysOfWeek()} ${hours}`;
    } else {
      return `Runs yearly on ${CronDescriptionGeneratorService.monthMap[this.cronFields.month]} ${hours}`;
    }
  }

}
