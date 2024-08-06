import {HttpResponse} from '@angular/common/http';
import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal, NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import {Observable, Subject} from 'rxjs';
import {filter, map, takeUntil, tap} from 'rxjs/operators';
import { SortEvent } from 'src/app/shared/directives/sortable/type/sort-event';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ALERT_TIMESTAMP_FIELD} from '../../../../shared/constants/alert/alert-field.constant';
import {RULE_FIELDS} from '../../../models/rule.constant';
import {Asset, Rule} from '../../../models/rule.model';
import {FilterService} from '../../../services/filter.service';
import {RuleService} from '../../../services/rule.service';
import {itemsPerPage} from '../../../services/rules.resolver.service';
import {Actions} from '../../../app-correlation-management/models/config.type';
import {ConfigService} from '../../../app-correlation-management/services/config.service';
import {AddRuleComponent} from '../add-rule/add-rule.component';


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
              private filterService: FilterService,
              private ruleService: RuleService,
              private modalService: NgbModal,
              private utmToast: UtmToastService,
              private configService: ConfigService) { }

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

    this.configService.action$
      .pipe(
        takeUntil(this.destroy$),
        filter(action => action === Actions.CREATE_RULE)
      )
      .subscribe(() => this.addRule());
  }

  addRule(){
    const modalRef = this.modalService.open(AddRuleComponent, {size: 'lg', centered: true});
    this.handleResponse(modalRef);
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

  activeRule(rule: Rule){
    const params = {
      id: rule.id,
      active: !rule.ruleActive
    };

    this.ruleService.activeRule(params)
      .subscribe(() => this.loadRules(),
        () => {
          this.utmToast.showError('Error', 'Error changing rule status');
        });
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

  deleteRule(event: Event, rule: Rule) {
    event.stopPropagation();
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModalRef.componentInstance.header = 'Delete rule';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete this rule?';
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-display';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.result.then(() => {
      this.delete(rule);
    });
  }

  delete(rule: Rule) {
    this.ruleService.delete(rule.id)
      .subscribe(() => {
        this.loadRules();
        this.utmToast.showSuccessBottom('Rule deleted successfully');
      }, () => {
        this.utmToast.showError('Error', 'Error deleting rule');
      });
  }

  editRule(rule: Rule) {
    const modal = this.modalService.open(AddRuleComponent, {size: 'lg', centered: true});
    modal.componentInstance.rule = rule;

    this.handleResponse(modal);
  }

  handleResponse(modal: NgbModalRef) {
    modal.result.then((result: boolean) => {
      if (result) {
        this.loadRules();
      }
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
