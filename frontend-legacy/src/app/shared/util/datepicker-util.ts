/**
 * Return valid dates to pass to datepicker
 */
export function setMaxDateToday(): { year: number, month: number, day: number } {
  const dateObj = new Date();
  return {
    year: dateObj.getUTCFullYear(),
    month: dateObj.getUTCMonth() + 1,
    day: dateObj.getUTCDate()
  };
}

export function compareDate(): boolean {
  if (this.fromDate !== undefined && this.toDate !== undefined) {
    if (this.fromDate !== null && this.toDate !== null) {
      return this.fromDate.getTime() - this.toDate.getTime() > 0;
    } else {
      return false;
    }
  }
}
