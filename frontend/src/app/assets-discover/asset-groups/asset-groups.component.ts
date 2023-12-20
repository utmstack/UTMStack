import {Component, OnDestroy, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ResizeEvent} from 'angular-resizable-element';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ElasticFilterDefaultTime} from '../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {ModalConfirmationComponent} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {AssetFiltersBehavior} from '../shared/behavior/asset-filters.behavior';
import {AssetFieldFilterEnum} from '../shared/enums/asset-field-filter.enum';
import {UtmAssetGroupService} from '../shared/services/utm-asset-group.service';
import {AssetGroupCreateComponent} from './asset-group-create/asset-group-create.component';
import {ASSETS_GROUP_FIELDS_FILTERS} from './shared/const/asset-group-field.const';
import {GROUP_STATIC_FILTER} from './shared/const/asset-group.const';
import {AssetGroupFilterType} from './shared/type/asset-group-filter.type';
import {AssetGroupType} from './shared/type/asset-group.type';

@Component({
  selector: 'app-asset-groups',
  templateUrl: './asset-groups.component.html',
  styleUrls: ['./asset-groups.component.scss']
})
export class AssetGroupsComponent implements OnInit, OnDestroy {
  assetGroups: AssetGroupType[];
  defaultTime: ElasticFilterDefaultTime = new ElasticFilterDefaultTime('now-1y', 'now');
  pageWidth = window.innerWidth;
  filterWidth: number;
  tableWidth: number;
  sortEvent: any;
  totalItems: any;
  page = 1;
  loading = true;
  itemsPerPage = ITEMS_PER_PAGE;
  sortBy: string;
  checkbox: any;
  fieldFilters = ASSETS_GROUP_FIELDS_FILTERS;
  requestParam: AssetGroupFilterType = {
    assetIp: null,
    assetName: null,
    os: null,
    probe: null,
    type: null,
    groupName: null,
    id: null,
    page: 0,
    size: ITEMS_PER_PAGE,
    sort: 'id,desc',
  };
  groupsSelected: number[] = [];
  interval: any;
  // Init get group on time filter component trigger
  viewGroupDetail: AssetGroupType;

  constructor(private utmAssetGroupService: UtmAssetGroupService,
              private modalService: NgbModal,
              private utmToastService: UtmToastService,
              private assetFiltersBehavior: AssetFiltersBehavior) {
  }

  ngOnInit() {
    this.setInitialWidth();
    this.getAssetsGroups();
  }

  ngOnDestroy(): void {
    clearInterval(this.interval);
    this.assetFiltersBehavior.$assetFilter.next(null);
  }

  setInitialWidth() {
    if (this.pageWidth > 1980) {
      this.filterWidth = 350;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    } else {
      this.filterWidth = 300;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    }
    if (this.pageWidth > 2500) {
      this.filterWidth = 350;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    }
    if (this.pageWidth > 4000) {
      this.filterWidth = 400;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
    }
  }

  loadPage($event: number) {
    this.requestParam.page = $event - 1;
    this.getAssetsGroups();
  }

  getAssetsGroups() {
    this.utmAssetGroupService.query(this.requestParam).subscribe(response => {
      this.totalItems = Number(response.headers.get('X-Total-Count'));
      this.assetGroups = response.body;
      this.loading = false;
    });
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.requestParam.size = $event;
    this.getAssetsGroups();
  }

  onResize($event: ResizeEvent) {
    if ($event.rectangle.width >= 250) {
      this.tableWidth = (this.pageWidth - $event.rectangle.width - 51);
      this.filterWidth = $event.rectangle.width;
    }
  }

  onSortBy($event: SortEvent) {
    this.requestParam.sort = $event.column + ',' + $event.direction;
    this.getAssetsGroups();
  }

  toggleCheck() {
    this.checkbox = !this.checkbox;
    if (!this.checkbox) {
      this.groupsSelected = [];
    } else {
      this.groupsSelected = this.assetGroups.map(value => value.id);
    }
  }

  addToSelected(group: AssetGroupType) {
    const index = this.groupsSelected.findIndex(value => value === group.id);
    if (index === -1) {
      this.groupsSelected.push(group.id);
    } else {
      this.groupsSelected.splice(index, 1);
    }
  }

  isSelected(group: AssetGroupType): boolean {
    return this.groupsSelected.findIndex(value => value === group.id) !== -1;
  }

  resetAllFilters() {
    for (const key of Object.keys(this.requestParam)) {
      if (!GROUP_STATIC_FILTER.includes(key)) {
        this.requestParam[key] = null;
      }
    }
    this.assetFiltersBehavior.$assetFilter.next(this.requestParam);
    this.getAssetsGroups();
  }

  onFilterChange($event: { prop: AssetFieldFilterEnum, values: any }) {
    switch ($event.prop) {
      case AssetFieldFilterEnum.TYPE:
        this.requestParam.type = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.OS:
        this.requestParam.os = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.PROBE:
        this.requestParam.probe = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.IP:
        this.requestParam.assetIp = $event.values.length > 0 ? $event.values : null;
        break;
      case AssetFieldFilterEnum.NAME:
        this.requestParam.assetName = $event.values.length > 0 ? $event.values : null;
        break;
    }
    this.assetFiltersBehavior.$assetFilter.next(this.requestParam);
    this.assetFiltersBehavior.$assetAppliedFilter.next(this.requestParam);
    this.getAssetsGroups();
  }

  onSearch($event: string) {
    this.requestParam.groupName = $event;
    this.requestParam.page = 0;
    this.getAssetsGroups();
  }

  openDeleteConfirmation(group: AssetGroupType) {
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModalRef.componentInstance.header = 'Delete group';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete :' + group.groupName + '?';
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-database-remove';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.result.then(() => {
      this.deleteGroup(group);
    });
  }

  deleteGroup(group: AssetGroupType) {
    this.utmAssetGroupService.delete(group.id).subscribe(() => {
      const index = this.assetGroups.findIndex(value => value.id === group.id);
      this.assetGroups.splice(index, 1);
      this.utmToastService.showSuccessBottom('Group ' + group.groupName + ' deleted successfully');
    });
  }

  addGroup() {
    const modalGroup = this.modalService.open(AssetGroupCreateComponent, {centered: true});
    modalGroup.componentInstance.addGroup.subscribe(() => {
      this.getAssetsGroups();
    });
  }

  editGroup(group: AssetGroupType) {
    const modalGroup = this.modalService.open(AssetGroupCreateComponent, {centered: true});
    modalGroup.componentInstance.group = group;
    modalGroup.componentInstance.addGroup.subscribe(() => {
      this.getAssetsGroups();
    });
  }
}
