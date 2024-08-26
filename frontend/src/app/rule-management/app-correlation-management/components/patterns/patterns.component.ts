import {HttpResponse} from '@angular/common/http';
import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbActiveModal, NgbModal, NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import {Observable, Subject} from 'rxjs';
import {filter, map, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ALERT_TIMESTAMP_FIELD} from '../../../../shared/constants/alert/alert-field.constant';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {RULE_FIELDS} from '../../../models/rule.constant';
import {RegexPattern} from '../../../models/rule.model';
import { itemsPerPage } from '../../../models/rule.model';
import {Actions} from '../../models/config.type';
import {ConfigService} from '../../services/config.service';
import {PatternManagerService} from '../../services/pattern-manager.service';
import {AddPatternComponent} from './components/add-pattern.component';

@Component({
  selector: 'app-pattern-manager',
  templateUrl: './patterns.component.html',
  styleUrls: ['./patterns.component.scss']
})
export class PatternsComponent implements OnInit, OnDestroy {

  patterns: RegexPattern[];
  patterns$: Observable<RegexPattern[]>;

  sortEvent: SortEvent;
  sortBy = ALERT_TIMESTAMP_FIELD + ',desc';
  fields = RULE_FIELDS;
  checkbox: any;
  typesSelected: RegexPattern[] = [];

  page = 0;
  totalItems: number;
  itemsPerPage = itemsPerPage;

  dataType: any;
  loading: any;
  viewRegexPatternDetail: any;
  ruleDetail: RegexPattern;
  isInitialized = false;
  request: any;
  destroy$: Subject<void> = new Subject<void>();

  constructor(private route: ActivatedRoute,
              private patternManagerService: PatternManagerService,
              private configService: ConfigService,
              public activeModal: NgbActiveModal,
              private utmToast: UtmToastService,
              private modalService: NgbModal) { }

  ngOnInit() {
    this.request = {
      page: this.page,
      size: this.itemsPerPage
    };

    this.patterns$ = this.route.data
        .pipe(
            tap((data: { response: HttpResponse<RegexPattern[]> }) => {
              this.handleDataResponse(data.response);
            }),
            map((data: { response: HttpResponse<RegexPattern[]> }) =>  data.response.body)
        );

    this.configService.action$
        .pipe(
            takeUntil(this.destroy$),
            filter(action => action === Actions.CREATE_PATTERN)
        )
        .subscribe(() => this.addRegexPattern());

  }

  loadRegexPatterns() {
    this.loading = true;
    this.patterns$ = this.patternManagerService.getAll(this.request)
        .pipe(
            tap(( response: HttpResponse<RegexPattern[]> ) => {
              this.handleDataResponse(response);
            }),
            map((response: HttpResponse<RegexPattern[]> ) =>  response.body));
  }

  addRegexPattern() {
    const modal = this.modalService.open(AddPatternComponent, {centered: true});
    this.handleResponse(modal);
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
    this.loadRegexPatterns();
  }

  deleteType(event: Event, type: RegexPattern) {
    event.stopPropagation();
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModalRef.componentInstance.header = 'Delete pattern';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete this pattern?';
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-display';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.result.then(() => {
      this.delete(type);
    });
  }

  delete(pattern: RegexPattern) {
    this.patternManagerService.delete(pattern.id)
        .subscribe(() => {
          this.loadRegexPatterns();
          this.utmToast.showSuccessBottom('Regex pattern deleted successfully');
        }, () => {
          this.utmToast.showError('Error', 'Error deleting regex pattern');
        });
  }

  onItemsPerPageChange($event: number) {
    if (this.isInitialized) {
      this.isInitialized = false;
      return;
    }
    this.itemsPerPage = $event;
    this.request = {
      ...this.request,
      size: this.itemsPerPage
    };
    this.loadRegexPatterns();
  }
  onSortBy($event: SortEvent) {
    const sort =  $event.column + ',' + $event.direction;
    this.request = {
      ...this.request,
      sort
    };
    this.loadRegexPatterns();
  }
  toggleCheck() {
    this.checkbox = !this.checkbox;
    if (!this.checkbox) {
      this.typesSelected = [];
    } else {
      for (const rule of this.patterns) {
        const index = this.typesSelected.indexOf(rule);
        if (index === -1) {
          this.typesSelected.push(rule);
        }
      }
    }
  }

  addToSelected(alert: any) {
    const index = this.typesSelected.indexOf(alert);
    if (index === -1) {
      this.typesSelected.push(alert);
    } else {
      this.typesSelected.splice(index, 1);
    }
  }

  isSelected(pattern: RegexPattern): boolean {
    return this.typesSelected.findIndex(value => value.id === pattern.id) !== -1;
  }

  viewDetailRegexPattern(rule: RegexPattern) {
    this.ruleDetail = rule;
    this.viewRegexPatternDetail = true;
  }

  trackByFn(index: number, pattern: RegexPattern): any {
    return pattern.id;
  }

  onSearch($event: string | number) {
      this.request = {
        search: $event,
        page: 0
      };

      this.loadRegexPatterns();
  }

  editRegexPattern(pattern: RegexPattern) {
    const modal = this.modalService.open(AddPatternComponent, {centered: true});
    modal.componentInstance.pattern = pattern;

    this.handleResponse(modal);
  }

  handleResponse(modal: NgbModalRef) {
    modal.result.then((result: boolean) => {
      if (result) {
        this.loadRegexPatterns();
      }
    });
  }

  private handleDataResponse(response: HttpResponse<RegexPattern[]>) {
    this.loading = false;
    this.patterns = response.body;
    this.totalItems = parseInt(response.headers.get('X-Total-Count') || '0', 10);
    this.isInitialized = true;
  }

  ngOnDestroy(): void {
    this.configService.onAction(null);
    this.destroy$.next();
    this.destroy$.complete();
  }

}
