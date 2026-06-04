export interface ISchedule {
  id?: string;
  uuid?: string;
  name?: string;
  comment?: string;
  creationTime?: string;
  modificationTime?: string;
  writable?: string;
  inUse?: string;
  firstTime?: string;
  nextTime?: string;
  timezone?: string;
  timezoneAbbrev?: string;
  period?: string;
  periodMonths?: string;
  simplePeriod?: {
    value?: string,
    unit?: string
  };
  duration?: string;
  simpleDuration?: {
    value?: string,
    unit?: string
  };
  tasks?: {
    id: string,
    name: string
  }[];

}

export class ScheduleModel implements ISchedule {
  constructor(
    public uuid?: string,
    public id?: string,
    public name?: string,
    public comment?: string,
    public creationTime?: string,
    public modificationTime?: string,
    public writable?: string,
    public  inUse?: string,
    public  firstTime?: string,
    public  nextTime?: string,
    public  timezone?: string,
    public  timezoneAbbrev?: string,
    public  period?: string,
    public  periodMonths?: string,
    public  simplePeriod?: {
      value?: string,
      unit?: string
    },
    public duration?: string,
    public simpleDuration?: {
      value?: string,
      unit?: string
    },
    public tasks?: {
      id: string,
      name: string
    }[]
  ) {
    this.comment = comment ? comment : null;
    this.creationTime = creationTime ? creationTime : null;
    this.duration = duration ? duration : null;
    this.firstTime = firstTime ? firstTime : null;
    this.writable = writable ? writable : null;
    this.modificationTime = modificationTime ? modificationTime : null;
    this.name = name ? name : null;
    this.inUse = inUse ? inUse : null;
    this.period = period ? period : null;
    this.periodMonths = periodMonths ? periodMonths : null;
    this.timezone = timezone ? timezone : null;
    this.uuid = uuid ? uuid : null;
    this.id = id ? id : null;
    this.nextTime = nextTime ? nextTime : null;
    this.timezoneAbbrev = timezoneAbbrev ? timezoneAbbrev : null;
    this.simplePeriod = simplePeriod ? simplePeriod : null;
    this.simpleDuration = simpleDuration ? simpleDuration : null;
    this.tasks = tasks ? tasks : null;
  }
}
