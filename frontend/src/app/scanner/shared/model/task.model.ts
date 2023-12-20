import {TaskStatusEnum} from '../../scanner-config/task/shared/enums/task-status.enum';
import {TaskTrendEnum} from '../../scanner-config/task/shared/enums/task-trend.enum';

export interface ITask {
  id?: string;
  uuid?: string;
  name?: string;
  comment?: string;
  creationTime?: string;
  modificationTime?: string;
  writable?: string;
  inUse?: string;
  status?: string;
  progress?: {
    value?: number
  };
  alterable?: string;
  config?: {
    uuid?: string,
    name?: string,
    trash?: string
  };
  target?: {
    uuid?: string,
    name?: string,
    trash?: string
  };
  scanner?: {
    uuid?: string,
    name?: string,
    trash?: string,
    type?: string
  };
  schedule?: {
    uuid?: string,
    name?: string,
    nextTime?: string,
    trash?: string,
    firstTime?: string,
    period?: string,
    periodMonths?: string,
    duration?: string
  };
  reportCount?: {
    value?: string,
    finished?: string
  };
  trend?: string;
  lastReport?: {
    report?: {
      uuid?: string,
      timestamp?: string,
      scanStart?: string,
      scanEnd?: string,
      resultCount?: {
        debug?: string,
        hole?: string,
        info?: string,
        log?: string,
        warning?: string,
        falsePositive?: string
      },
      severity?: string
    }
  };
  firstReport?: {
    report?: {
      uuid?: string,
      timestamp?: string,
      scanStart?: string,
      scanEnd?: string,
      resultCount?: {
        debug?: string,
        hole?: string,
        info?: string,
        log?: string,
        warning?: string,
        falsePositive?: string
      },
      severity?: string
    }
  };
  averageDuration?: string;
  reports?: string;
  resultCount?: string;
  preferences?: {
    name?: string, scannerName?: string, value?: string
  }[];

}

export class TaskModel implements ITask {
  constructor(
    public id?: string,
    public uuid?: string,
    public name?: string,
    public comment?: string,
    public creationTime?: string,
    public modificationTime?: string,
    public writable?: string,
    public inUse?: string,
    public status?: TaskStatusEnum,
    public progress?: {
      value?: number
    },
    public alterable?: any,
    public config?: {
      uuid?: string,
      name?: string,
      trash?: string
    },
    public target?: {
      uuid?: string,
      name?: string,
      trash?: string
    },
    public  scanner?: {
      uuid?: string,
      name?: string,
      trash?: string,
      type?: string
    },
    public  schedule?: {
      uuid?: string,
      name?: string,
      nextTime?: string,
      trash?: string,
      firstTime?: string,
      period?: string,
      periodMonths?: string,
      duration?: string
    },
    public reportCount?: {
      value?: string,
      finished?: string
    },
    public trend?: TaskTrendEnum,
    public  lastReport?: {
      report?: {
        uuid?: string,
        timestamp?: string,
        scanStart?: string,
        scanEnd?: string,
        resultCount?: {
          debug?: string,
          hole?: string,
          info?: string,
          log?: string,
          warning?: string,
          falsePositive?: string
        },
        severity?: string
      }
    },
    public firstReport?: {
      report?: {
        uuid?: string,
        timestamp?: string,
        scanStart?: string,
        scanEnd?: string,
        resultCount?: {
          debug?: string,
          hole?: string,
          info?: string,
          log?: string,
          warning?: string,
          falsePositive?: string
        },
        severity?: string
      }
    },
    public averageDuration?: string,
    public reports?: string,
    public resultCount?: string,
    public preferences?: {
      name?: string, scannerName?: string, value?: string
    }[]
  ) {
    this.id = id ? id : null;
    this.name = name ? name : null;
    this.comment = comment ? comment : null;
    this.creationTime = creationTime ? creationTime : null;
    this.modificationTime = modificationTime ? modificationTime : null;
    this.writable = writable ? writable : null;
    this.inUse = inUse ? inUse : null;
    this.status = status ? status : null;
    this.progress = progress ? progress : null;
    this.alterable = alterable ? alterable : null;
    this.config = config ? config : null;
    this.target = target ? target : null;
    this.scanner = scanner ? scanner : null;
    this.schedule = schedule ? schedule : null;
    this.reportCount = reportCount ? reportCount : null;
    this.trend = trend ? trend : null;
    this.lastReport = lastReport ? lastReport : null;
    this.firstReport = firstReport ? firstReport : null;
    this.averageDuration = averageDuration ? averageDuration : null;
    this.reports = reports ? reports : null;
    this.resultCount = resultCount ? resultCount : null;
    this.preferences = preferences ? preferences : null;
  }

}
