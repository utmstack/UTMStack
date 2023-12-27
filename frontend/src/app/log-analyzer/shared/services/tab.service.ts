import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {TabType} from '../type/tab.type';

@Injectable()
export class TabService {
  public tabs: TabType[] = [];
  public tabSub = new BehaviorSubject<TabType[]>(this.tabs);

  public removeTab(index: number) {
    this.tabs.splice(index, 1);
    if (this.tabs.length > 0) {
      this.tabs[this.tabs.length - 1].active = true;
    }
    this.tabSub.next(this.tabs);
  }

  public addTab(tab: TabType) {
    // tslint:disable-next-line:prefer-for-of
    for (let i = 0; i < this.tabs.length; i++) {
      if (this.tabs[i].active === true) {
        this.tabs[i].active = false;
      }
    }
    tab.id = this.tabs.length + 1;
    tab.active = true;
    this.tabs.push(tab);
    this.tabSub.next(this.tabs);
  }

  public setActiveTab(tabId: number) {
    this.tabs.forEach(value => {
      if (value.id !== tabId && value.active) {
        value.active = false;
      }
    });
    const index = this.tabs.findIndex(value => value.id === tabId);
    this.tabs[index].active = true;
    this.tabSub.next(this.tabs);
  }
}
