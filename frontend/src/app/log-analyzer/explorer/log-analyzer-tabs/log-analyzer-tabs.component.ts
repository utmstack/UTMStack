import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute, NavigationEnd, NavigationStart, Params, Router} from '@angular/router';
import {UUID} from 'angular2-uuid';
import {Observable, Subject} from 'rxjs';
import {filter, takeUntil, tap} from 'rxjs/operators';
import {LogAnalyzerQueryService} from '../../shared/services/log-analyzer-query.service';
import {TabService} from '../../shared/services/tab.service';
import {LogAnalyzerQueryType} from '../../shared/type/log-analyzer-query.type';
import {TabType} from '../../shared/type/tab.type';
import {LogAnalyzerViewComponent} from '../log-analyzer-view/log-analyzer-view.component';
import {UtmIndexPattern} from "../../../shared/types/index-pattern/utm-index-pattern";
import {IndexPatternBehavior} from "../../shared/behaviors/index-pattern.behavior";

@Component({
  selector: 'app-log-analyzer-tabs',
  templateUrl: './log-analyzer-tabs.component.html',
  styleUrls: ['./log-analyzer-tabs.component.scss']
})
export class LogAnalyzerTabsComponent implements OnInit, OnDestroy {
  tabs$: Observable<TabType[]>;
  // tabs = new Array<TabType>();
  tabSelected: TabType;
  tabNumber = 0;
  queryId: string;
  query: LogAnalyzerQueryType;
  destroy$: Subject<void> = new Subject<void>();

  constructor(public tabService: TabService,
              private logAnalyzerQueryService: LogAnalyzerQueryService,
              private activatedRoute: ActivatedRoute,
              private router: Router,
              private indexPatternBehavior: IndexPatternBehavior
            ) {}

  ngOnInit() {
    this.activatedRoute.queryParams.subscribe(params => {
      this.queryId = params.queryId;
      if (this.queryId) {
        this.logAnalyzerQueryService.find(this.queryId).subscribe(vis => {
          this.query = vis.body;
          this.addNewTab(this.query.name, this.query, params);
        });
      } else {
        this.addNewTab();
      }
    });

    this.tabs$ = this.tabService.tabs$.pipe(
      tap((tabs) => {
        this.tabSelected = tabs.find(t => t.active);
        this.tabNumber = tabs.length;
      })
    );

    this.tabService.onSaveTabSubject
      .pipe(takeUntil(this.destroy$),
            filter(query => !!query))
      .subscribe(query => this.tabService.updateActiveTab(query));

    this.router.events.pipe(
      filter(event => event instanceof NavigationStart),
      tap((event: NavigationStart) => {
        if (event.url !== '/discover/log-analyzer-queries') {
          this.tabService.closeAllTabs();
        }
      }),
      takeUntil(this.destroy$)
    ).subscribe();
  }

  tabChanged(tab: TabType) {
    this.tabSelected = tab;
    this.tabService.setActiveTab(tab.id);
  }

  addNewTab(tabName?: string, query?: LogAnalyzerQueryType, params?: Params) {
    const uuid = UUID.UUID();
    this.tabNumber += 1;
    this.tabService.addTab(
      new TabType(LogAnalyzerViewComponent, (tabName ? tabName : 'New query ' + this.tabNumber),
        query ? query : null, true, null, uuid)
    );

    if (tabName) {
      if (params.patternId) {
        const pattern = new UtmIndexPattern(params.patternId, params.indexPattern, true);
        this.indexPatternBehavior.$pattern.next({pattern, tabUUID: uuid});
      }
    }
  }

  removeTab(index: number): void {
    this.tabService.removeTab(index);
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }


  format(){
    const numbers = [1, 2, 3, 4, 5, 6];

    numbers.map(num => {
      if (num % 2 === 0) {
        return num * 2;
      }
      return num;
    });
  }

   countOccurrences(arr: number[]) {
     const countMap = {};
     for (const num of arr) {
       countMap[num] = (countMap[num] || 0) + 1;
     }

     return countMap;
  }

}
