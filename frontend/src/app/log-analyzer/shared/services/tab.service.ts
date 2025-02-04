import { Injectable } from '@angular/core';
import { BehaviorSubject, Subject } from 'rxjs';
import { LogAnalyzerQueryType } from '../type/log-analyzer-query.type';
import { TabType } from '../type/tab.type';

@Injectable()
export class TabService {
  private tabs: TabType[] = [];
  private tabSubject = new BehaviorSubject<TabType[]>(this.tabs);
  public tabs$ = this.tabSubject.asObservable();
  public onSaveTabSubject = new Subject<LogAnalyzerQueryType>();

  /**
   * Add a new tab and make it active.
   * @param tab The tab to add.
   */
  public addTab(tab: TabType): void {
    this.deactivateAllTabs();
    const newTab: TabType = {
      ...tab,
      id: this.tabs.length + 1,
      active: true,
    };
    this.tabs.push(newTab);
    this.emitTabs();
  }

  /**
   * Remove a tab by index. Activates the last tab if any remain.
   * @param index The index of the tab to remove.
   */
  public removeTab(index: number): void {
    this.tabs.splice(index, 1);
    if (this.tabs.length > 0) {
      this.tabs[this.tabs.length - 1].active = true;
    }
    this.emitTabs();
  }

  /**
   * Set a specific tab as active by its ID.
   * @param tabId The ID of the tab to activate.
   */
  public setActiveTab(tabId: number): void {
    this.deactivateAllTabs();
    const tab = this.tabs.find(t => t.id === tabId);
    if (tab) {
      tab.active = true;
    }
    this.emitTabs();
  }

  /**
   * Update the active tab with new data.
   * @param query The query containing updated data.
   */
  public updateActiveTab(query: LogAnalyzerQueryType): void {
    const activeTab = this.getActiveTab();
    if (activeTab) {
      activeTab.title = query.name;
      this.emitTabs();
    }
  }

  /**
   * Close all tabs.
   */
  public closeAllTabs(): void {
    this.tabs = [];
    this.emitTabs();
  }

  /**
   * Get the currently active tab.
   * @returns The active tab, or undefined if no tab is active.
   */
  public getActiveTab(): TabType | undefined {
    return this.tabs.find(t => t.active);
  }

  /**
   * Delete the currently active tab.
   */
  public deleteActiveTab(): void {
    const activeIndex = this.tabs.findIndex(t => t.active);
    if (activeIndex !== -1) {
      this.tabs.splice(activeIndex, 1);
      if (this.tabs.length > 0) {
        this.tabs[this.tabs.length - 1].active = true;
      }
      this.emitTabs();
    }
  }

  /**
   * Get the total number of tabs.
   * @returns The number of tabs.
   */
  public getTabCount(): number {
    return this.tabs.length;
  }

  /**
   * Deactivate all tabs.
   */
  private deactivateAllTabs(): void {
    this.tabs.forEach(t => (t.active = false));
  }

  /**
   * Emit the latest state of the tabs.
   */
  private emitTabs(): void {
    this.tabSubject.next(this.tabs);
  }
}
