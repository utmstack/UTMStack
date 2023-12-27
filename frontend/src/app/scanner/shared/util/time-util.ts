// export today to disable datepiker
export function datepickerToday(): { year: number, month: number, day: number } {
  const today = new Date();
  return {year: today.getFullYear(), month: today.getMonth() - 1, day: today.getDay()};
}
