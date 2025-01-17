import {Type} from '@angular/core';
import {LogAnalyzerQueryType} from './log-analyzer-query.type';

export class TabType {
  public id?: number;
  public title: string;
  public tabData: LogAnalyzerQueryType;
  public active: boolean;
  public component: Type<any>;
  public uuid?: string;

  constructor(component: Type<any>, title: string, tabData: any, active: boolean, id?: number,
              uuid?: string) {
    this.id = id;
    this.tabData = tabData;
    this.component = component;
    this.title = title;
    this.active = active;
    this.uuid = uuid;
  }
}
