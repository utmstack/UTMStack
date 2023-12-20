import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {UUID} from 'angular2-uuid';
import {LogAnalyzerQueryService} from '../../shared/services/log-analyzer-query.service';
import {TabService} from '../../shared/services/tab.service';
import {LogAnalyzerQueryType} from '../../shared/type/log-analyzer-query.type';
import {TabType} from '../../shared/type/tab.type';
import {LogAnalyzerViewComponent} from '../log-analyzer-view/log-analyzer-view.component';

@Component({
  selector: 'app-log-analyzer-tabs',
  templateUrl: './log-analyzer-tabs.component.html',
  styleUrls: ['./log-analyzer-tabs.component.scss']
})
export class LogAnalyzerTabsComponent implements OnInit, OnDestroy {
  tabs = new Array<TabType>();
  selectedTab: number;
  tabSelected: TabType;
  tabNumber = 0;
  queryId: string;
  query: LogAnalyzerQueryType;

  constructor(public tabService: TabService,
              private logAnalyzerQueryService: LogAnalyzerQueryService,
              private activatedRoute: ActivatedRoute) {
    activatedRoute.queryParams.subscribe(params => {
      this.queryId = params.queryId;
    });
  }

  ngOnInit() {
    if (this.queryId) {
      this.logAnalyzerQueryService.find(this.queryId).subscribe(vis => {
        this.query = vis.body;
        this.addNewTab(this.query.name, this.query);
      });
    } else {
      this.addNewTab();
    }
    this.tabService.tabSub.subscribe(tabs => {
      this.tabs = tabs;
      this.selectedTab = tabs.findIndex(tab => tab.active);
      this.tabSelected = this.tabs[this.selectedTab];
    });
  }

  ngOnDestroy(): void {
    this.tabService.tabs = [];
  }

  tabChanged(tab: TabType) {
    this.tabSelected = tab;
    this.tabService.setActiveTab(tab.id);
  }

  addNewTab(tabName?: string, query?: LogAnalyzerQueryType) {
    this.tabNumber += 1;
    this.tabService.addTab(
      new TabType(LogAnalyzerViewComponent, (tabName ? tabName : 'New query ' + this.tabNumber),
        query ? query : null, true, null, UUID.UUID())
    );
  }

  removeTab(index: number): void {
    this.tabService.removeTab(index);
  }

}
