import {Injectable, Query} from '@angular/core';
import {BehaviorSubject, Subject} from 'rxjs';
import {LogAnalyzerQueryType} from '../type/log-analyzer-query.type';
import {TabType} from '../type/tab.type';

@Injectable()
export class TabService {
  private tabs: TabType[] = [];
  private tabSub = new BehaviorSubject<TabType[]>(this.tabs);
  public tabs$ = this.tabSub.asObservable();
  public onSaveTabSubject = new Subject<LogAnalyzerQueryType>();

  public removeTab(index: number) {
    this.tabs.splice(index, 1);
    if (this.tabs.length > 0) {
      this.tabs[this.tabs.length - 1].active = true;
    }
    this.tabSub.next(this.tabs);
  }

  public addTab(tab: TabType) {
    this.tabs.forEach(t => t.active = false);
    tab = { ...tab, id: this.tabs.length + 1, active: true };
    this.tabs = [
      ...this.tabs,
      tab
    ];
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

  updateActiveTab(query: LogAnalyzerQueryType){
    const activeTab = this.tabs.find(t => t.active);
    const tabs = this.tabs.filter(tab => !tab.active);

    this.tabs = [
      ...tabs,
      {
        ...activeTab,
        title: query.name,
      }
    ];
    this.tabSub.next(this.tabs);
  }

  closeAllTabs() {
    this.tabs = [];
    this.tabSub.next(this.tabs);
  }
}
