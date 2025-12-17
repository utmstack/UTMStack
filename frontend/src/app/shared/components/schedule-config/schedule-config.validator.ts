export interface ScheduleConfig {
  days: number[];
  startTime: any;
  endTime: any;
}

export class ScheduleConfigValidator {
  static isValid(config: ScheduleConfig): boolean {
    if (!config) { return false; }

    if (!Array.isArray(config.days) || config.days.length === 0) { return false; }
    const validDays = config.days.every(d => Number.isInteger(d) && d >= 0 && d <= 6);
    if (!validDays) { return false; }

    if (!this.isValidTime(config.startTime)) { return false; }
    if (!this.isValidTime(config.endTime)) { return false; }

    const start = this.toMinutes(config.startTime);
    const end = this.toMinutes(config.endTime);

    if (start >= end) { return false; }

    return true;
  }

  private static isValidTime(time: any): boolean {
    if (!time) { return false; }

    if (typeof time === 'object' && 'hour' in time && 'minute' in time) {
      return (
        Number.isInteger(time.hour) &&
        Number.isInteger(time.minute) &&
        time.hour >= 0 && time.hour < 24 &&
        time.minute >= 0 && time.minute < 60
      );
    }

    if (typeof time === 'string') {
      const regex = /^([01]\d|2[0-3]):([0-5]\d)$/;
      return regex.test(time);
    }

    return false;
  }

  private static toMinutes(time: any): number {
    if (typeof time === 'object' && 'hour' in time && 'minute' in time) {
      return time.hour * 60 + time.minute;
    }
    if (typeof time === 'string') {
      const [h, m] = time.split(':').map(Number);
      return h * 60 + m;
    }
    return 0;
  }
}
