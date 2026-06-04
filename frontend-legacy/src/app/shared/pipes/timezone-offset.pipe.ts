import { Pipe, PipeTransform } from '@angular/core';
import * as moment from 'moment-timezone';

@Pipe({
  name: 'timezoneOffset'
})
export class TimezoneOffsetPipe implements PipeTransform {

  transform(timezone: string): string {
    const offset = moment.tz(timezone).format('Z');
    return `GMT${offset}`;
  }
}
