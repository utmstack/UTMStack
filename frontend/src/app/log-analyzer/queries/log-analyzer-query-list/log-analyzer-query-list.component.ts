import {HttpResponse} from '@angular/common/http';
import {Component, OnDestroy, OnInit} from '@angular/core';
import {NavigationStart, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Subject} from 'rxjs';
import {filter, takeUntil, tap} from 'rxjs/operators';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {SortByType} from '../../../shared/types/sort-by.type';
import {LogAnalyzerQueryService} from '../../shared/services/log-analyzer-query.service';
import {TabService} from '../../shared/services/tab.service';
import {LogAnalyzerQueryType} from '../../shared/type/log-analyzer-query.type';
import {LogAnalyzerQueryDeleteComponent} from '../log-analyzer-query-delete/log-analyzer-query-delete.component';
import {query} from "@angular/animations";
import {data} from "../../../active-directory/offline.data";

@Component({
  selector: 'app-log-analyzer-query-list',
  templateUrl: './log-analyzer-query-list.component.html',
  styleUrls: ['./log-analyzer-query-list.component.scss']
})
export class LogAnalyzerQueryListComponent implements OnInit, OnDestroy {
  fields: SortByType[] = [
    {
      fieldName: 'Name',
      field: 'name'
    },
    {
      fieldName: 'Last modification',
      field: 'modifiedDate'
    }
  ];
  loading = false;
  totalItems: any;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  queries: LogAnalyzerQueryType[] = [];
  searching = false;
  query: LogAnalyzerQueryType;
  private requestParams: any;
  private sortBy: SortEvent;
  destroy$ = new Subject<void>();

  constructor(private logAnalyzerQueryService: LogAnalyzerQueryService,
              private router: Router,
              private modalService: NgbModal,
              private tabService: TabService) {
  }

  ngOnInit() {
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
    };
    this.getQueryList();

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

  onSearchQuery($event: string) {
    // TODO
  }

  onSort($event: SortEvent) {
    this.requestParams.sort = $event.column + ',' + $event.direction;
    this.getQueryList();
  }

  editQuery(query: LogAnalyzerQueryType) {
    this.router.navigate(['/discover/log-analyzer'], {
      queryParams: {
        mode: 'edit',
        queryId: query.id,
        queryName: query.name.toLowerCase().replace(' ', '_'),
        patternId: query.pattern.id,
        indexPattern: query.pattern.pattern,
      }
    });
  }

  getQueryList() {
    this.logAnalyzerQueryService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  deleteQuery(query: LogAnalyzerQueryType) {
    const modal = this.modalService.open(LogAnalyzerQueryDeleteComponent, {centered: true});
    modal.componentInstance.query = query;
    modal.componentInstance.queryDeleted.subscribe(deleted => {
      this.getQueryList();
    });
  }

  loadPage(page: any) {
    this.requestParams.page = page;
    this.getQueryList();
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.queries = data;
    this.loading = false;
  }

  private onError(body: any) {
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
