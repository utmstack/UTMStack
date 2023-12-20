export const TIME_INTERVAL: BucketDateHistogramInterval[] = [
  // {interval: '1ms', label: 'Millisecond'},
  // {interval: '1s', label: 'Second'},
  {interval: 'Minute', label: 'Minute'},
  {interval: 'Hour', label: 'Hourly'},
  {interval: 'Day', label: 'Daily'},
  {interval: 'Week', label: 'Weekly'},
  {interval: 'Month', label: 'Monthly'},
  {interval: 'Quarter', label: 'Quarterly'},
  {interval: 'Year', label: 'Yearly'}
];


// Second("second", "1s"),
//
//   Minute("minute", "1m"),
//
//   Hour("hour", "1h"),
//
//   Day("day", "1d"),
//
//   Week("week", "1w"),
//
//   Month("month", "1M"),
//
//   Quarter("quarter", "1q"),
//
//   Year("year", "1Y"),

export class BucketDateHistogramInterval {
  interval: string;
  label: string;
}
