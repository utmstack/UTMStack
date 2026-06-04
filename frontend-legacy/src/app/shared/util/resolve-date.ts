import * as moment from 'moment';

/**
 * Return date range based on time defined(year,month,week,all,day) if is not defined
 * get value day
 * @param time Range to return date
 */
export function resolveRangeByTime(time: string): { timeFrom: string, timeTo: string, range: string } {
  let dateTo = moment.parseZone().utc().format();
  let dateFrom;
  if (time === 'year') {
    // dateTo = new Date(moment().format('YYYY-MM-DD HH:MM:ss'));
    dateFrom = moment.parseZone(moment().subtract(365, 'd')).utc().format();
  } else if (time === 'month') {
    dateFrom = moment.parseZone(moment().subtract(30, 'd')).utc().format();
  } else if (time === 'week') {
    dateFrom = moment.parseZone(moment().subtract(7, 'd')).utc().format();
  } else if (time === 'all') {
    dateTo = null;
    dateFrom = null;
  } else {
    dateFrom = moment.parseZone(moment().subtract(1, 'd')).utc().format();
    time = 'day';
  }

  return {
    timeFrom: dateFrom ? dateFrom : null,
    timeTo: dateTo ? dateTo : null,
    range: time
  };
}

/**
 * Return label based on time if is not defined
 * get value day
 * @param time String of time range
 */
export function resolveFilterName(time): string {
  let label = '';
  switch (time) {
    case 'day':
      label = 'alert.filter.time.day';
      break;
    case 'week':
      label = 'alert.filter.time.week';
      break;
    case 'month':
      label = 'alert.filter.time.month';
      break;
    case 'year':
      label = 'alert.filter.time.year';
      break;
    case 'all':
      label = 'alert.filter.time.all';
      break;
    default :
      label = time;
      break;
  }
  return label;
}
