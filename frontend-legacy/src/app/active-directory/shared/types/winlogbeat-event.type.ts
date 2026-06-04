export class WinlogbeatEventType {
  // tslint?:disable-next-line?:variable-name
  id: string;
  timestamp: string;
  logx?: {
    wineventlog?: {
      message?: string
      eventId?: number,
      eventName?: string,
      computerName?: string,
      eventData?: {
        subjectUserSid?: string,
        subjectUserName?: string,
        targetUserSid?: string,
        targetUserName?: string,
        targetDomainName?: string
      }
    }
  };
}

