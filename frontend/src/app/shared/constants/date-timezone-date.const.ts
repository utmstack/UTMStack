export const DATE_SECTION_ID = 5;

export const DEFAULT_DATE_SETTING_TIMEZONE = 'UTC';
export const DEFAULT_DATE_SETTING_DATE = 'medium';

export const DATE_SETTING_TIMEZONE_SHORT = 'utmstack.time.zone';
export const DATE_SETTING_FORMAT_SHORT = 'utmstack.time.dateformat';

/*export const TIMEZONES: Array<{ label: string; timezone: string, zone: string }> = [
  {label: 'UTC', timezone: 'UTC', zone: 'UTC'},
  {label: 'Eastern Standard Time (New York)', timezone: 'America/New_York', zone: 'America'},
  {label: 'Pacific Standard Time (Los Angeles)', timezone: 'America/Los_Angeles', zone: 'America'},
  {label: 'Central Standard Time (Chicago)', timezone: 'America/Chicago', zone: 'America'},
  {label: 'Mountain Standard Time (Denver)', timezone: 'America/Denver', zone: 'America'},
  {label: 'Atlantic Standard Time (Halifax)', timezone: 'America/Halifax', zone: 'America'},
  {label: 'Alaska Standard Time (Anchorage)', timezone: 'America/Anchorage', zone: 'America'},
  {label: 'Hawaii-Aleutian Standard Time (Honolulu)', timezone: 'Pacific/Honolulu', zone: 'Pacific'},
  {label: 'London (GMT)', timezone: 'Europe/London', zone: 'Europe'},
  {label: 'Paris (CET)', timezone: 'Europe/Paris', zone: 'Europe'},
  {label: 'Berlin (CET)', timezone: 'Europe/Berlin', zone: 'Europe'},
  {label: 'Madrid (CET)', timezone: 'Europe/Madrid', zone: 'Europe'},
  {label: 'Rome (CET)', timezone: 'Europe/Rome', zone: 'Europe'},
  {label: 'Moscow (MSK)', timezone: 'Europe/Moscow', zone: 'Europe'},
  {label: 'Istanbul (TRT)', timezone: 'Europe/Istanbul', zone: 'Europe'},
  {label: 'Sydney (AEST)', timezone: 'Australia/Sydney', zone: 'Australia'},
  {label: 'Melbourne (AEST)', timezone: 'Australia/Melbourne', zone: 'Australia'},
  {label: 'Perth (AWST)', timezone: 'Australia/Perth', zone: 'Australia'},
  {label: 'Beijing (CST)', timezone: 'Asia/Shanghai', zone: 'Asia'},
  {label: 'Tokyo (JST)', timezone: 'Asia/Tokyo', zone: 'Asia'},
  {label: 'Seoul (KST)', timezone: 'Asia/Seoul', zone: 'Asia'},
  {label: 'Singapore (SGT)', timezone: 'Asia/Singapore', zone: 'Asia'},
  {label: 'Hong Kong (HKT)', timezone: 'Asia/Hong_Kong', zone: 'Asia'},
  {label: 'New Delhi (IST)', timezone: 'Asia/Kolkata', zone: 'Asia'},
  {label: 'Dubai (GST)', timezone: 'Asia/Dubai', zone: 'Asia'},
  {label: 'Jerusalem (IST)', timezone: 'Asia/Jerusalem', zone: 'Asia'},
  {label: 'Buenos Aires (ART)', timezone: 'America/Argentina/Buenos_Aires', zone: 'America'},
  {label: 'São Paulo (BRT)', timezone: 'America/Sao_Paulo', zone: 'America'},
];*/

export const TIMEZONES: Array<{ label: string; timezone: string, zone: string }> = [
  {label: getOffset('UTC'), timezone: 'UTC', zone: 'UTC'},
  {label: getOffset('America/New_York'), timezone: 'America/New_York', zone: 'America'},
  {label: getOffset('America/Los_Angeles'), timezone: 'America/Los_Angeles', zone: 'America'},
  {label: getOffset('America/Chicago'), timezone: 'America/Chicago', zone: 'America'},
  {label: getOffset('America/Denver'), timezone: 'America/Denver', zone: 'America'},
  {label: getOffset('America/Halifax'), timezone: 'America/Halifax', zone: 'America'},
  {label: getOffset('America/Anchorage'), timezone: 'America/Anchorage', zone: 'America'},
  {label: getOffset('Pacific/Honolulu'), timezone: 'Pacific/Honolulu', zone: 'Pacific'},
  {label: getOffset('Europe/London'), timezone: 'Europe/London', zone: 'Europe'},
  {label: getOffset('Europe/Paris'), timezone: 'Europe/Paris', zone: 'Europe'},
  {label: getOffset('Europe/Berlin'), timezone: 'Europe/Berlin', zone: 'Europe'},
  {label: getOffset('Europe/Madrid'), timezone: 'Europe/Madrid', zone: 'Europe'},
  {label: getOffset('Europe/Rome'), timezone: 'Europe/Rome', zone: 'Europe'},
  {label: getOffset('Europe/Moscow'), timezone: 'Europe/Moscow', zone: 'Europe'},
  {label: getOffset('Europe/Istanbul'), timezone: 'Europe/Istanbul', zone: 'Europe'},
  {label: getOffset('Australia/Sydney'), timezone: 'Australia/Sydney', zone: 'Australia'},
  {label: getOffset('Australia/Melbourne'), timezone: 'Australia/Melbourne', zone: 'Australia'},
  {label: getOffset('Australia/Perth'), timezone: 'Australia/Perth', zone: 'Australia'},
  {label: getOffset('Asia/Shanghai'), timezone: 'Asia/Shanghai', zone: 'Asia'},
  {label: getOffset('Asia/Tokyo'), timezone: 'Asia/Tokyo', zone: 'Asia'},
  {label: getOffset('Asia/Seoul'), timezone: 'Asia/Seoul', zone: 'Asia'},
  {label: getOffset('Asia/Singapore'), timezone: 'Asia/Singapore', zone: 'Asia'},
  {label: getOffset('Asia/Hong_Kong'), timezone: 'Asia/Hong_Kong', zone: 'Asia'},
  {label: getOffset('Asia/Kolkata'), timezone: 'Asia/Kolkata', zone: 'Asia'},
  {label: getOffset('Asia/Dubai'), timezone: 'Asia/Dubai', zone: 'Asia'},
  {label: getOffset('Asia/Jerusalem'), timezone: 'Asia/Jerusalem', zone: 'Asia'},
  {label: getOffset('America/Argentina/Buenos_Aires'), timezone: 'America/Argentina/Buenos_Aires', zone: 'America'},
  {label: getOffset('America/Sao_Paulo'), timezone: 'America/Sao_Paulo', zone: 'America'},
];


export const DATE_FORMATS: Array<{ label: string; format: string; equivalentTo: string }> = [
  {label: 'Short', format: 'short', equivalentTo: 'M/d/yy, h:mm a'},
  {label: 'Medium', format: 'medium', equivalentTo: 'MMM d, y, h:mm:ss a'},
  {label: 'Long', format: 'long', equivalentTo: 'MMMM d, y, h:mm:ss a z'},
  {label: 'Full', format: 'full', equivalentTo: 'EEEE, MMMM d, y, h:mm:ss a zzzz'},
  {label: 'Short Date', format: 'shortDate', equivalentTo: 'M/d/yy'},
  {label: 'Medium Date', format: 'mediumDate', equivalentTo: 'MMM d, y'},
  {label: 'Long Date', format: 'longDate', equivalentTo: 'MMMM d, y'},
  {label: 'Full Date', format: 'fullDate', equivalentTo: 'EEEE, MMMM d, y'},
  {label: 'Short Time', format: 'shortTime', equivalentTo: 'h:mm a'},
  {label: 'Medium Time', format: 'mediumTime', equivalentTo: 'h:mm:ss a'},
  {label: 'Long Time', format: 'longTime', equivalentTo: 'h:mm:ss a z'},
  {label: 'Full Time', format: 'fullTime', equivalentTo: 'h:mm:ss a zzzz'},
  // Custom formats
  {label: 'Year-Month-Day', format: 'yyyy-MM-dd', equivalentTo: '2021-01-31'},
  {label: 'Month/Day/Year', format: 'MM/dd/yyyy', equivalentTo: '01/31/2021'},
  {label: 'Day-Month-Year', format: 'dd-MM-yyyy', equivalentTo: '31-01-2021'},
  {label: 'Hour:Minute AM/PM', format: 'hh:mm a', equivalentTo: '05:30 PM'},
  {label: 'Hour:Minute:Second', format: 'HH:mm:ss', equivalentTo: '17:30:00'},
];

export function getDateInfo(
  searchValue: string
): { label: string; format: string; equivalentTo: string } | null {
  const formatObj = DATE_FORMATS.find(
    (format) => format.format === searchValue
  );

  return formatObj ? formatObj : null;
}

// Helper function to get the UTC offset
function getOffset(timezone: string): string {
  const now = new Date();
  const options = {
    timeZone: timezone,
    timeZoneName: 'shortOffset'  // Esto devuelve el offset en formato GMT±X
  };
  const formatter = new Intl.DateTimeFormat('en-US', options);
  const parts = formatter.formatToParts(now);

  // Buscamos la parte que contiene el GMT±X
  const gmtPart = parts.find(part => part.type === 'timeZoneName');
  return gmtPart ? gmtPart.value : '';
}
