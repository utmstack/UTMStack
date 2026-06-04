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
import {Asset, itemsPerPage} from '../../../models/rule.model';
import {Actions} from '../../models/config.type';
import {AssetManagerService} from '../../services/asset-manager.service';
import {ConfigService} from '../../services/config.service';
import {AddAssetComponent} from './components/components/add-asset/add-asset.component';
import {IncidentSeverityEnum} from '../../../../shared/enums/incident/incident-severity.enum';

@Component({
  selector: 'app-assets',
  templateUrl: './assets.component.html',
  styleUrls: ['./assets.component.scss']
})
export class AssetsComponent implements OnInit, OnDestroy {

  assets: Asset[];
  assets$: Observable<Asset[]>;

  sortEvent: SortEvent;
  sortBy = ALERT_TIMESTAMP_FIELD + ',desc';
  fields = RULE_FIELDS;
  checkbox: any;
  typesSelected: Asset[] = [];

  page = 0;
  totalItems: number;
  itemsPerPage = itemsPerPage;

  dataType: any;
  loading: any;
  viewAssetDetail: any;
  ruleDetail: Asset;
  isInitialized = false;
  request: any;
  destroy$: Subject<void> = new Subject<void>();
  viewDetail = false;
  asset: Asset;

  constructor(private route: ActivatedRoute,
              private assetManagerService: AssetManagerService,
              private configService: ConfigService,
              public activeModal: NgbActiveModal,
              private utmToast: UtmToastService,
              private modalService: NgbModal) { }

  ngOnInit() {
    this.request = {
      page: this.page,
      size: this.itemsPerPage
    };

    this.assets$ = this.route.data
      .pipe(
        tap((data: { response: HttpResponse<Asset[]> }) => {
          this.handleDataResponse(data.response);
        }),
        map((data: { response: HttpResponse<Asset[]> }) =>  data.response.body)
      );

    this.configService.action$
      .pipe(
        takeUntil(this.destroy$),
        filter(action => action === Actions.CREATE_ASSET)
      )
      .subscribe(() => this.addAsset());

  }

  loadAssets() {
    this.loading = true;
    this.assets$ = this.assetManagerService.getAll(this.request)
      .pipe(
        tap(( response: HttpResponse<Asset[]> ) => {
          this.handleDataResponse(response);
        }),
        map((response: HttpResponse<Asset[]> ) =>  response.body));
  }

  addAsset() {
    const modal = this.modalService.open(AddAssetComponent, {centered: true});
    this.handleResponse(modal);
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
    this.loadAssets();
  }

  deleteAsset(event: Event, asset: Asset) {
    event.stopPropagation();
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModalRef.componentInstance.header = 'Delete asset';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete this asset?';
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-display';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.result.then(() => {
      this.delete(asset);
    });
  }

  delete(asset: Asset) {
    this.assetManagerService.delete(asset.id)
      .subscribe(() => {
        this.loadAssets();
        this.utmToast.showSuccessBottom('Asset deleted successfully');
      }, () => {
        this.utmToast.showError('Error', 'Error deleting regex asset');
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
    this.loadAssets();
  }
  onSortBy($event: SortEvent) {
    const sort =  $event.column + ',' + $event.direction;
    this.request = {
      ...this.request,
      sort
    };
    this.loadAssets();
  }
  toggleCheck() {
    this.checkbox = !this.checkbox;
    if (!this.checkbox) {
      this.typesSelected = [];
    } else {
      for (const rule of this.assets) {
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

  isSelected(asset: Asset): boolean {
    return this.typesSelected.findIndex(value => value.id === asset.id) !== -1;
  }

  viewDetailAsset(rule: Asset) {
    this.ruleDetail = rule;
    this.viewAssetDetail = true;
  }

  trackByFn(index: number, asset: Asset): any {
    return asset.id;
  }

  onSearch($event: string | number) {
    this.request = {
      search: $event,
      page: 0
    };

    this.loadAssets();
  }

  editAsset(asset: Asset) {
    const modal = this.modalService.open(AddAssetComponent, {centered: true});
    modal.componentInstance.asset = asset;

    this.handleResponse(modal);
  }

  handleResponse(modal: NgbModalRef) {
    modal.result.then((result: boolean) => {
      if (result) {
        this.loadAssets();
      }
    });
  }

  private handleDataResponse(response: HttpResponse<Asset[]>) {
    this.loading = false;
    this.assets = response.body;
    this.totalItems = parseInt(response.headers.get('X-Total-Count') || '0', 10);
    this.isInitialized = true;
  }

  getSeverity(value: any): IncidentSeverityEnum {
    switch (value) {
      case 1: return IncidentSeverityEnum.LOW;
      case 2: return IncidentSeverityEnum.MEDIUM;
      case 3: return IncidentSeverityEnum.HIGH;
    }
  }

  onViewDetail($event: Asset){
    this.viewDetail = true;
    this.asset = $event;
  }

  ngOnDestroy(): void {
    this.configService.onAction(null);
    this.destroy$.next();
    this.destroy$.complete();
  }

}
