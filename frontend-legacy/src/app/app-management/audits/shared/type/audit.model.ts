import {UtmAuditData} from './audit-data.model';

export class UtmAudit {
  constructor(public data: UtmAuditData, public principal: string, public timestamp: string, public type: string) {
  }
}
