import {Component, Input, OnInit} from '@angular/core';
import {ElasticFilterDefaultTime} from '../../../../shared/components/utm/filters/elastic-filter-time/elastic-filter-time.component';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {UtmFieldType} from '../../../../shared/types/table/utm-field.type';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';
import {AssetGroupType} from '../../../asset-groups/shared/type/asset-group.type';
import {ASSETS_FIELDS} from '../../const/asset-field.const';
import {AssetFieldEnum} from '../../enums/asset-field.enum';
import {UtmNetScanService} from '../../services/utm-net-scan.service';
import {AssetFilterType} from '../../types/asset-filter.type';
import {NetScanType} from '../../types/net-scan.type';

@Component({
  selector: 'app-asset-by-group',
  templateUrl: './asset-by-group.component.html',
  styleUrls: ['./asset-by-group.component.scss']
})
export class AssetByGroupComponent implements OnInit {
  @Input() group: AssetGroupType;
  defaultTime: ElasticFilterDefaultTime = new ElasticFilterDefaultTime('now-1y', 'now');
  page = 1;
  loading = true;
  itemsPerPage = ITEMS_PER_PAGE;
  requestParam: AssetFilterType = {
    discoveredEndDate: null,
    discoveredInitDate: null,
    groups: null,
    page: 0,
    size: ITEMS_PER_PAGE,
    sort: null
  };
  totalItems: number;
  assets: NetScanType[] = [];
  assetsFields: UtmFieldType[] = ASSETS_FIELDS.filter(value =>
    value.field === AssetFieldEnum.ASSET_IP ||
    value.field === AssetFieldEnum.ASSET_NAME ||
    value.field === AssetFieldEnum.ASSET_SEVERITY ||
    value.field === AssetFieldEnum.ASSET_METRICS);
  assetFieldEnum = AssetFieldEnum;

  constructor(private utmNetScanService: UtmNetScanService) {
  }

  ngOnInit() {
  }

  loadPage($event: number) {
    this.requestParam.page = $event - 1;
    this.getAssets();
  }

  getAssets() {
    this.utmNetScanService.query(this.requestParam).subscribe(response => {
      this.totalItems = Number(response.headers.get('X-Total-Count'));
      this.assets = response.body;
      this.loading = false;
    });
  }

  onSortBy($event: SortEvent) {
    this.requestParam.sort = $event.column + ',' + $event.direction;
    this.getAssets();
  }

  onSearch($event: string) {
    this.requestParam.assetIpMacName = $event;
    this.requestParam.page = 0;
    this.getAssets();
  }

  onItemsPerPageChange($event: number) {
    this.itemsPerPage = $event;
    this.requestParam.size = $event;
    this.getAssets();
  }

  onTimeFilterChange($event: TimeFilterType) {
    this.requestParam.groups = [this.group.groupName];
    this.requestParam.discoveredInitDate = $event.timeFrom;
    this.requestParam.discoveredEndDate = $event.timeTo;
    this.getAssets();
  }

}
