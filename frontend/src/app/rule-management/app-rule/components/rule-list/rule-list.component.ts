import {HttpResponse} from '@angular/common/http';
import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {ResizeEvent} from 'angular-resizable-element';
import {Observable, Subject} from 'rxjs';
import {filter, map, takeUntil, tap} from 'rxjs/operators';
import { SortEvent } from 'src/app/shared/directives/sortable/type/sort-event';
import {ALERT_TIMESTAMP_FIELD} from '../../../../shared/constants/alert/alert-field.constant';
import {UtmFieldType} from '../../../../shared/types/table/utm-field.type';
import {RULE_FIELDS} from '../../../models/rule.constant';
import {Rule} from '../../../models/rule.model';
import {FilterService} from '../../../services/filter.service';
import {RuleService} from '../../../services/rule.service';
import {itemsPerPage} from '../../../services/rules.resolver.service';


@Component({
  selector: 'app-rule-list',
  templateUrl: './rule-list.component.html',
  styleUrls: ['./rule-list.component.css']
})
export class RuleListComponent implements OnInit, OnDestroy {

  @Input()
  // rulesResponse$: Observable<HttpResponse<Rule[]>>;
  rules$: Observable<Rule[]>;
  rules: Rule[];

  sortEvent: SortEvent;
  sortBy = ALERT_TIMESTAMP_FIELD + ',desc';
  fields = RULE_FIELDS;
  checkbox: any;
  rulesSelected: Rule[] = [];

  page = 0;
  totalItems: number;
  itemsPerPage = itemsPerPage;

  dataType: any;
  loading: any;
  viewRuleDetail: any;
  ruleDetail: Rule;
  isInitialized = false;
  request: any;
  destroy$: Subject<void> = new Subject<void>();

  constructor(private route: ActivatedRoute,
              private router: Router,
              private filterService: FilterService,
              private ruleService: RuleService) { }

  ngOnInit() {
    this.request = {
      page: this.page,
      size: this.itemsPerPage
    };

    this.rules$ = this.route.data
        .pipe(
          tap((data: { response: HttpResponse<Rule[]> }) => {
            this.rules = data.response.body;
            this.totalItems = parseInt(data.response.headers.get('X-Total-Count') || '0', 10);
            this.isInitialized = true;
          }),
          map((data: { response: HttpResponse<Rule[]> }) =>  data.response.body)
    );

    this.filterService.filterChange
        .pipe(
            takeUntil(this.destroy$),
            filter( (request: { prop: string, values: any }) => !!request),
            tap((request) => {
              if (request && Object.keys(request).length === 0) {
                this.request = {
                  page: 0,
                  size: this.itemsPerPage
                };
              } else {
                this.request = {
                  ...this.request,
                  [request.prop]: request.values.length > 0 ? request.values : null
                };
              }

              this.loadRules();
            })
        ).subscribe();
  }

  loadRules() {
    this.loading = true;
    this.rules$ = this.ruleService.getRules(this.request)
        .pipe(
            tap(( response: HttpResponse<Rule[]> ) => {
              this.rules = response.body;
              this.totalItems = parseInt(response.headers.get('X-Total-Count') || '0', 10);
              this.loading = false;
            }),
            map((response: HttpResponse<Rule[]> ) =>  response.body));
  }

  addRule() {

  }

  onResize($event: ResizeEvent) {

  }

  loadPage($event: number) {
    if (this.isInitialized) {
      this.isInitialized = false;
      return;
    }
    const page = $event - 1;
    this.request = {
      ...this.request,
      page
    };
    this.loadRules();
  }

  onItemsPerPageChange($event: number) {

  }
  getRowToFiltersData(alert: any) {
    /*const modalRef = this.modalService.open(RowToFiltersComponent, {centered: true});
    modalRef.componentInstance.alert = alert;
    modalRef.componentInstance.fields = this.fields;
    modalRef.componentInstance.addRowToFilter.subscribe(filterRow => {
      mergeParams(filterRow, this.filters).then(value => {
        this.filters = value;
        this.page = 1;
        this.getAlert('on add row to filter');
        this.alertFiltersBehavior.$filters.next(this.filters);
      });
    });*/
  }

  onRefreshData($event: boolean) {

  }

  onSortBy($event: SortEvent) {
    const sort =  $event.column + ',' + $event.direction;
    this.request = {
      ...this.request,
      sort
    };
    this.loadRules();
  }
  toggleCheck() {
    this.checkbox = !this.checkbox;
    if (!this.checkbox) {
      this.rulesSelected = [];
    } else {
      for (const rule of this.rules) {
        const index = this.rulesSelected.indexOf(rule);
        if (index === -1) {
          this.rulesSelected.push(rule);
        }
      }
    }
  }

  addToSelected(alert: any) {
    const index = this.rulesSelected.indexOf(alert);
    if (index === -1) {
      this.rulesSelected.push(alert);
    } else {
      this.rulesSelected.splice(index, 1);
    }
  }

  isSelected(alert: any): boolean {
    return this.rulesSelected.findIndex(value => value.id === alert.id) !== -1;
  }

  viewDetailRule(rule: Rule) {
    this.ruleDetail = rule;
    this.viewRuleDetail = true;
  }

  trackByFn(index: number, rule: Rule): any {
    return rule.id;
  }

  onSearch($event: string | number) {
    this.request = {
      page: 0,
      search: $event
    };
    this.loadRules();
  }

  deleteRule(rule: Rule) {

  }

  editRule(rule: Rule) {
    const {id} = rule;
    this.router.navigate(['alerting-rules/rule', id]);
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
