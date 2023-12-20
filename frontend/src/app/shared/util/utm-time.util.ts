import * as moment from 'moment';
import {ElasticTimeEnum} from '../enums/elastic-time.enum';

/**
 * Return instant Date from elastic range
 * @param time ElasticTimeEnum
 * @param last Number for calc
 */
export function resolveInstantDate(time: ElasticTimeEnum, last: number): { timeFrom: string, timeTo: string } {
  // const dateTo = new Date(moment().format('YYYY-MM-DD HH:MM:ss')).toISOString();
  const dateTo = moment.parseZone(moment().add(1, 'd')).utc().format();
  let dateFrom;
  switch (time) {
    case ElasticTimeEnum.YEAR:
      dateFrom = moment.parseZone(moment().subtract(last, 'y')).utc().format();
      break;
    case ElasticTimeEnum.MONTH:
      dateFrom = moment.parseZone(moment().subtract(last, 'M')).utc().format();
      break;
    case ElasticTimeEnum.WEEKS:
      dateFrom = moment.parseZone(moment().subtract(last, 'w')).utc().format();
      break;
    case ElasticTimeEnum.DAY:
      dateFrom = moment.parseZone(moment().subtract(last, 'd')).utc().format();
      break;
    case ElasticTimeEnum.HOUR:
      dateFrom = moment.parseZone(moment().subtract(last, 'h')).utc().format();
      break;
    case ElasticTimeEnum.MINUTE:
      dateFrom = moment.parseZone(moment().subtract(last, 'm')).utc().format();
      break;
    case ElasticTimeEnum.SECOND:
      dateFrom = moment.parseZone(moment().subtract(last, 's')).utc().format();
      break;
  }
  return {
    timeFrom: dateFrom,
    timeTo: dateTo,
  };
}

export function resolveUTCDate(date: string): string {
  return moment(date).utc().format();
}

/**
 * Return format instant from date
 * @param time rangeTime
 */
export function buildFormatInstantFromDate(time: { from: any, to: any }): { timeFrom: any, timeTo: any } {
  const last = Number(time.from.match(/\d+/g));
  const unit = time.from.match(/[a-zA-Z]+/g)[1];
  return resolveInstantDate(unit, last);
}
