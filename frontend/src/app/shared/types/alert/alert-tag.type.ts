export interface IAlertTags {
  id?: number;
  tagName?: string;
  tagColor?: string;
  systemOwner?: boolean;
}

export class AlertTags implements IAlertTags {
  constructor(public id?: number,
              public tagName?: string,
              public tagColor?: string,
              public systemOwner?: boolean) {
    this.id = id ? id : null;
    this.tagName = tagName ? tagName : null;
    this.tagColor = tagColor ? tagColor : null;
    this.systemOwner = systemOwner ? systemOwner : null;
  }
}
