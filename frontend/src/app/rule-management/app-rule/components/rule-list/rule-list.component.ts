import {HttpResponse} from '@angular/common/http';
import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal, NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {merge, Observable, of, Subject} from 'rxjs';
import {catchError, filter, map, switchMap, takeUntil, tap} from 'rxjs/operators';
import { SortEvent } from 'src/app/shared/directives/sortable/type/sort-event';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ALERT_TIMESTAMP_FIELD} from '../../../../shared/constants/alert/alert-field.constant';
import {Actions} from '../../../app-correlation-management/models/config.type';
import {ConfigService} from '../../../app-correlation-management/services/config.service';
import {RULE_FIELDS} from '../../../models/rule.constant';
import { Rule, RULE_REQUEST} from '../../../models/rule.model';
import {FilterService} from '../../../services/filter.service';
import {RuleService} from '../../../services/rule.service';
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

  dataType: any;
  loading: any;
  viewRuleDetail: any;
  ruleDetail: Rule;
  isInitialized = false;
  request = RULE_REQUEST;
  destroy$: Subject<void> = new Subject<void>();

  constructor(private route: ActivatedRoute,
              private filterService: FilterService,
              private ruleService: RuleService,
              private modalService: NgbModal,
              private utmToast: UtmToastService,
              private configService: ConfigService) { }

  ngOnInit() {

    this.rules$ = merge(
      this.ruleService.onRefresh$.pipe(
        filter(refresh => !!refresh),
        switchMap(() => this.ruleService.fetchData(this.request))
      ),
      this.route.data
        .pipe(
          tap((data: { response: HttpResponse<Rule[]> }) => {
            this.totalItems = parseInt(data.response.headers.get('X-Total-Count') || '0', 10);
            this.isInitialized = true;
          }),
          map((data: { response: HttpResponse<Rule[]> }) => data.response))
    ).pipe(
      map(( response: HttpResponse<Rule[]>) => response.body)
    );

    this.filterService.filterChange
        .pipe(
            takeUntil(this.destroy$),
            filter( (request: { prop: string, values: any }) => !!request),
            tap((request) => {
              if (request && Object.keys(request).length === 0) {
                this.request = RULE_REQUEST;
              } else {
                this.request = {
                  ...RULE_REQUEST,
                  [request.prop]: request.values.length > 0 ? request.values : null
                };
              }
              this.ruleService.notifyRefresh(true);
            })
        ).subscribe();

    this.configService.action$
      .pipe(
        takeUntil(this.destroy$),
        filter(action => action === Actions.CREATE_RULE)
      )
      .subscribe(() => this.addRule());
  }

  addRule() {
    const modalRef = this.modalService.open(AddRuleComponent, {size: 'lg', centered: true});
    this.handleResponse(modalRef);
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
    this.ruleService.notifyRefresh(true);
  }

  onItemsPerPageChange(size: number) {
    this.request = {
      ...this.request,
      size
    };
    this.ruleService.notifyRefresh(true);
  }

  onSortBy($event: SortEvent) {
    const sort =  $event.column + ',' + $event.direction;
    this.request = {
      ...this.request,
      sort
    };
    this.ruleService.notifyRefresh(true);
  }

  viewDetailRule(rule: Rule) {
    this.ruleDetail = rule;
    this.viewRuleDetail = true;
  }

  activeRule(rule: Rule) {
    const params = {
      id: rule.id,
      active: !rule.ruleActive
    };

    this.ruleService.activeRule(params)
      .subscribe(() =>  this.ruleService.notifyRefresh(true),
        () => {
          this.utmToast.showError('Error', 'Error changing rule status');
        });
  }

  trackByFn(index: number, rule: Rule): any {
    return rule.id;
  }

  onSearch($event: string | number) {
    this.request = {
      ...RULE_REQUEST,
      search: $event
    };
    this.ruleService.notifyRefresh(true);
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
        this.ruleService.notifyRefresh(true);
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
        this.ruleService.notifyRefresh(true);
        this.filterService.notifyRefresh('ALL');
      }
    });
  }

  ngOnDestroy(): void {
    this.configService.onAction(null);
    this.destroy$.next();
    this.destroy$.complete();
  }
}
