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

  // Overloads mirror Angular 11's DatePipe.transform (v11 added a
  // (value: null | undefined) => null overload; a plain `: string` override
  // is no longer assignable - TS2416). Defaults live on the implementation
  // only (TS2371 forbids initializers in overload signatures).
  transform(value: Date | string | number, format?: string, timezone?: string, locale?: string): string | null;
  transform(value: null | undefined, format?: string, timezone?: string, locale?: string): null;
  transform(value: any, format?: string, timezone?: string, locale?: string): string | null {
    if (value === null || value === undefined) {
      return null; // base DatePipe returns null for null/undefined input
    }
    format = this.options.dateFormat || format || 'mediumDate';
    timezone = this.options.timezone || timezone || 'UTC';
    const timezoneOffset = momentTimezone(value).tz(timezone).format('Z');
    return super.transform(value, format, timezoneOffset, locale);
  }

}
