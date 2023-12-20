import {NgbDate, NgbTimeStruct} from '@ng-bootstrap/ng-bootstrap';

export function ngbDateToDate(date: NgbDate, time?: NgbTimeStruct): string {
  return date.year + '-' + validateDateNumber(date.month) + '-' +
    validateDateNumber(date.day) +
    'T' + validateDateNumber(time.hour) + ':' +
    validateDateNumber(time.minute) + ':' +
    validateDateNumber(time.second) + '.999Z';
}


/**
 * Determine if need add 0 at start of number
 * @param unit Number time
 */
export function validateDateNumber(unit) {
  return (unit < 10 ? ('0' + unit) : unit);
}

export function getCurrentDateTimeString() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  const day = String(now.getDate()).padStart(2, '0');
  const hours = String(now.getHours()).padStart(2, '0');
  const minutes = String(now.getMinutes()).padStart(2, '0');

  return `${year}${month}${day}${hours}${minutes}`;
}
