import {Type} from '@angular/core';
import {LogAnalyzerQueryType} from './log-analyzer-query.type';
import {
  ElasticFilterDefaultTime
} from "../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component";

export class TabType {
  public id?: number;
  public title: string;
  public tabData: LogAnalyzerQueryType;
  public active: boolean;
  public component: Type<any>;
  public uuid?: string;
  public defaultTime?: ElasticFilterDefaultTime;

  constructor(component: Type<any>, title: string, tabData: any, active: boolean, id?: number,
              uuid?: string, defaultTime?: ElasticFilterDefaultTime) {
    this.id = id;
    this.tabData = tabData;
    this.component = component;
    this.title = title;
    this.active = active;
    this.uuid = uuid;
    this.defaultTime = defaultTime;
  }
}
