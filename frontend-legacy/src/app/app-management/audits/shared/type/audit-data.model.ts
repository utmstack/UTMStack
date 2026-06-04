export class UtmAuditData {
  message: any;
  constructor(public remoteAddress: string, public sessionId: string, message: string) {
  }
}
