import {Type} from '@angular/core';

export class TabType {
  public id?: number;
  public title: string;
  public tabData: any;
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
