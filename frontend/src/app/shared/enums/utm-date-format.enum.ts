export enum UtmDateFormatEnum {
  SHORT = 'short',
  MEDIUM = 'medium',
  LONG = 'long',
  SHORT_DATE = 'shortDate',
  MEDIUM_DATE = 'mediumDate',
  LONG_DATE = 'longDate',
  FULL_DATE = 'fullDate',
  SHORT_TIME = 'shortTime',
  MEDIUM_TIME = 'mediumTime',
  LONG_TIME = 'longTime',
  FULL_TIME = 'fullTime',
  UTM_SHORT = 'M/d/yyyy, h:mm:ss a',
  UTM_SHORT_UTC = 'short'
}

// 'short': equivalent to 'M/d/yy, h:mm a' (6/15/15, 9:03 AM).
// 'medium': equivalent to 'MMM d, y, h:mm:ss a' (Jun 15, 2015, 9:03:01 AM).
// 'long': equivalent to 'MMMM d, y, h:mm:ss a z' (June 15, 2015 at 9:03:01 AM GMT+1).
// 'full': equivalent to 'EEEE, MMMM d, y, h:mm:ss a zzzz' (Monday, June 15, 2015 at 9:03:01 AM GMT+01:00).
// 'shortDate': equivalent to 'M/d/yy' (6/15/15).
// 'mediumDate': equivalent to 'MMM d, y' (Jun 15, 2015).
// 'longDate': equivalent to 'MMMM d, y' (June 15, 2015).
// 'fullDate': equivalent to 'EEEE, MMMM d, y' (Monday, June 15, 2015).
// 'shortTime': equivalent to 'h:mm a' (9:03 AM).
// 'mediumTime': equivalent to 'h:mm:ss a' (9:03:01 AM).
// 'longTime': equivalent to 'h:mm:ss a z' (9:03:01 AM GMT+1).
// 'fullTime': equivalent to 'h:mm:ss a zzzz' (9:03:01 AM GMT+01:00).
