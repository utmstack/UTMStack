// app/date-pipe-options.factory.ts


import {DatePipeDefaultOptions} from '../types/date-pipe-default-options';
import {TimezoneFormatService} from './utm-timezone.service';


export function datePipeOptionsFactory(timezoneFormatService: TimezoneFormatService): DatePipeDefaultOptions {
  let options: DatePipeDefaultOptions;

  timezoneFormatService.getDateFormatSubject().subscribe(format => {
    options = format;
  });

  return options;
}
