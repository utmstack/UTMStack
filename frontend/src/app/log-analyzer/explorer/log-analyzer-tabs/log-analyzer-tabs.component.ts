import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute, NavigationEnd, NavigationStart, Params, Router} from '@angular/router';
import {UUID} from 'angular2-uuid';
import {Observable, Subject} from 'rxjs';
import {filter, takeUntil, tap} from 'rxjs/operators';
import {UtmIndexPattern} from '../../../shared/types/index-pattern/utm-index-pattern';
import {IndexPatternBehavior} from '../../shared/behaviors/index-pattern.behavior';
import {LogAnalyzerQueryService} from '../../shared/services/log-analyzer-query.service';
import {TabService} from '../../shared/services/tab.service';
import {LogAnalyzerQueryType} from '../../shared/type/log-analyzer-query.type';
import {TabType} from '../../shared/type/tab.type';
import {LogAnalyzerViewComponent} from '../log-analyzer-view/log-analyzer-view.component';
import {
  ElasticFilterDefaultTime
} from "../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component";

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

  ngOnInit(): void {
    this.activatedRoute.queryParams
      .pipe(takeUntil(this.destroy$))
      .subscribe(params => {
        this.queryId = params.queryId;
        const isRefresh = params.refreshRoute || null;
        if (this.queryId) {
          this.logAnalyzerQueryService.find(this.queryId).subscribe(vis => {
            this.query = vis.body;
            this.addNewTab(this.query.name, this.query, params);
          });
        } else {
          if (isRefresh) {
            this.tabService.deleteActiveTab();
          }
          this.addNewTab(null, null, params);
        }
      });

    this.tabs$ = this.tabService.tabs$.pipe(
      tap((tabs) => {
        this.tabSelected = tabs.find(t => t.active);
      })
    );

    this.tabService.onSaveTabSubject
      .pipe(takeUntil(this.destroy$),
            filter(query => !!query))
      .subscribe(query => this.tabService.updateActiveTab(query));

    this.router.events.pipe(
      filter(event => event instanceof NavigationStart),
      tap((event: NavigationStart) => {
        if (event.url !== '/discover/log-analyzer-queries' && !event.url.includes('/discover/log-analyzer')) {
          this.tabService.closeAllTabs();
        }
      }),
      takeUntil(this.destroy$)
    ).subscribe();
  }

  tabChanged(tab: TabType) {
    this.tabSelected = tab;
    this.tabService.setActiveTab(tab.id);
    this.indexPatternBehavior.changePattern({pattern: tab.tabData.pattern, tabUUID: tab.uuid});
  }

  addNewTab(tabName?: string, query?: LogAnalyzerQueryType, params?: Params) {
    const uuid = UUID.UUID();
    this.tabNumber = this.tabService.getTabCount() + 1;
    const pattern = params && params.patternId ? new UtmIndexPattern(params.patternId, params.indexPattern, true) :
      new UtmIndexPattern(1, 'log-*', true);

    this.tabService.addTab(
      new TabType(
        LogAnalyzerViewComponent,
        (tabName ? tabName : 'New query ' + this.tabNumber),
        query ? query : {pattern},
        true,
        null,
        uuid,
        this.getDefaultTime(query))
    );

    if (tabName && pattern) {
        this.indexPatternBehavior.changePattern({pattern, tabUUID: uuid});
    }
  }

  getDefaultTime(query: LogAnalyzerQueryType): ElasticFilterDefaultTime {
    if (query) {
      const timestampFilter = query.filtersType.find(f => f.field === '@timestamp');
      return timestampFilter ? new ElasticFilterDefaultTime(timestampFilter.value[0], timestampFilter.value[1]) :
        new ElasticFilterDefaultTime('now-24h', 'now');
    } else {
      return new ElasticFilterDefaultTime('now-24h', 'now');
    }
  }

  removeTab(index: number): void {
    this.tabService.removeTab(index);
  }

  trackByFn(index: number, tab: TabType): any {
    return tab.uuid;
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
