import {DatePipe} from '@angular/common';
import {Inject, LOCALE_ID, Pipe, PipeTransform} from '@angular/core';
import * as momentTimezone from 'moment-timezone';
import {TimezoneFormatService} from '../services/utm-timezone.service';
import {DatePipeDefaultOptions} from '../types/date-pipe-default-options';

@Pipe({
  name: 'date'
})
export class UtmDatePipe extends DatePipe implements PipeTransform {
  options: DatePipeDefaultOptions;

  constructor(@Inject(LOCALE_ID) locale: string, timezoneFormatService: TimezoneFormatService) {
    super(locale);
    timezoneFormatService.getDateFormatSubject().subscribe(format => {
      this.options = format;
    });
  }

  transform(value: any, format = 'mediumDate', timezone = 'UTC', locale?: string): string {
    format = this.options.dateFormat || format;
    timezone = this.options.timezone || timezone;
    const timezoneOffset = momentTimezone(value).tz(timezone).format('Z');
    return super.transform(value, format, timezoneOffset, locale);
  }

}
