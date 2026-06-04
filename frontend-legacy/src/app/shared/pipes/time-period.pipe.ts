import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'timePeriod'
})
export class TimePeriodPipe implements PipeTransform {

  transform(value: string): string {
    // Match different time formats
    const minuteMatch = value.match(/^now-(\d+)m$/);
    const hourMatch = value.match(/^now-(\d+)h$/);
    const dayMatch = value.match(/^now-(\d+)d$/);
    const yearMatch = value.match(/^now-(\d+)y$/);

    if (minuteMatch) {
      const minutes = parseInt(minuteMatch[1], 10);
      if (minutes === 1) {
        return 'Last minute';
      }
      return `Last ${minutes} minutes`;
    }

    if (hourMatch) {
      const hours = parseInt(hourMatch[1], 10);
      if (hours === 1) {
        return 'Last hour';
      }
      return `Last ${hours} hours`;
    }

    if (dayMatch) {
      const days = parseInt(dayMatch[1], 10);
      if (days === 7) {
        return 'Last 7 days';
      } else if (days === 30) {
        return 'Last 30 days';
      } else if (days === 90) {
        return 'Last 90 days';
      }
      return `Last ${days} days`;
    }

    if (yearMatch) {
      const years = parseInt(yearMatch[1], 10);
      return `Last ${years} year${years > 1 ? 's' : ''}`;
    }

    return value;  // If no match, return the original value
  }
}
