import { Pipe, PipeTransform } from '@angular/core';
import * as moment from 'moment-timezone';

@Pipe({
  name: 'timezoneOffset'
})
export class TimezoneOffsetPipe implements PipeTransform {

  transform(timezone: string): string {
    console.log(timezone);
    const offset = moment.tz(timezone).format('Z');  // Obtener offset como +02:00
    return `GMT${offset}`;  // Concatenar con 'GMT'
  }

}
