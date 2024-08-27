import {HttpResponse} from '@angular/common/http';
import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
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
import {DataType} from '../../../models/rule.model';
import {DataTypeService} from '../../../services/data-type.service';
import {Actions} from '../../models/config.type';
import {ConfigService} from '../../services/config.service';
import {itemsPerPage} from '../../services/types.resolver.service';
import {AddTypeComponent} from './components/add-type.component';

@Component({
  selector: 'app-assets',
  templateUrl: './types.component.html',
  styleUrls: ['./types.component.scss']
})
export class TypesComponent implements OnInit, OnDestroy {

  types: DataType[];
  types$: Observable<DataType[]>;

  sortEvent: SortEvent;
  sortBy = ALERT_TIMESTAMP_FIELD + ',desc';
  fields = RULE_FIELDS;
  checkbox: any;
  typesSelected: DataType[] = [];

  page = 0;
  totalItems: number;
  itemsPerPage = itemsPerPage;

  dataType: any;
  loading: any;
  viewDataTypeDetail: any;
  ruleDetail: DataType;
  isInitialized = false;
  request: any;
  destroy$: Subject<void> = new Subject<void>();

  constructor(private route: ActivatedRoute,
              private router: Router,
              private dataTypeService: DataTypeService,
              private configService: ConfigService,
              public activeModal: NgbActiveModal,
              private utmToast: UtmToastService,
              private modalService: NgbModal) { }

  ngOnInit() {
    this.request = {
      page: this.page,
      size: this.itemsPerPage
    };

    this.types$ = this.route.data
        .pipe(
            tap((data: { response: HttpResponse<DataType[]> }) => {
              this.types = data.response.body;
              this.totalItems = parseInt(data.response.headers.get('X-Total-Count') || '0', 10);
              console.log('TOTAL', this.totalItems);
              this.isInitialized = true;
            }),
            map((data: { response: HttpResponse<DataType[]> }) =>  data.response.body)
        );

    this.configService.action$
        .pipe(
            takeUntil(this.destroy$),
            filter(action => action === Actions.CREATE_TYPE)
        )
        .subscribe(() => this.addDataType());

  }

  loadDataTypes() {
    this.loading = true;
    this.types$ = this.dataTypeService.getAll(this.request)
        .pipe(
            tap(( response: HttpResponse<DataType[]> ) => {
              this.types = response.body;
              this.totalItems = parseInt(response.headers.get('X-Total-Count') || '0', 10);
              this.loading = false;
            }),
            map((response: HttpResponse<DataType[]> ) =>  response.body));
  }

  addDataType() {
    const modal = this.modalService.open(AddTypeComponent, {centered: true});

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
    this.loadDataTypes();
  }

  deleteType(event: Event, type: DataType) {
    event.stopPropagation();
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModalRef.componentInstance.header = 'Delete data type';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete this data type?';
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-display';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.result.then(() => {
      this.delete(type);
    });
  }

  delete(type: DataType) {
    this.dataTypeService.delete(type.id)
        .subscribe(() => {
          this.loadDataTypes();
          this.utmToast.showSuccessBottom('Data type deleted successfully');
    }, () => {
      this.utmToast.showError('Error', 'Error deleting data type');
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
    this.loadDataTypes();
  }
  onSortBy($event: SortEvent) {
    const sort =  $event.column + ',' + $event.direction;
    this.request = {
      ...this.request,
      sort
    };
    this.loadDataTypes();
  }
  toggleCheck() {
    this.checkbox = !this.checkbox;
    if (!this.checkbox) {
      this.typesSelected = [];
    } else {
      for (const rule of this.types) {
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

  isSelected(alert: any): boolean {
    return this.typesSelected.findIndex(value => value.id === alert.id) !== -1;
  }

  viewDetailDataType(rule: DataType) {
    this.ruleDetail = rule;
    this.viewDataTypeDetail = true;
  }

  trackByFn(index: number, rule: DataType): any {
    return rule.id;
  }

  onSearch($event: string | number) {
      this.request = {
        page: 0,
        search: $event
      };
      this.loadDataTypes();
  }

  editDataType(type: DataType) {
    const modal = this.modalService.open(AddTypeComponent, {centered: true});
    modal.componentInstance.type = type;

    this.handleResponse(modal);
  }

  handleResponse(modal: NgbModalRef) {
    modal.result.then((result: boolean) => {
      if (result) {
        this.loadDataTypes();
      }
    });
  }

  ngOnDestroy(): void {
    this.configService.onAction(null);
    this.destroy$.next();
    this.destroy$.complete();
  }

}
