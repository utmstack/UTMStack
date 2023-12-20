import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {AssetFieldEnum} from '../enums/asset-field.enum';
import {NetScanType} from '../types/net-scan.type';

export function onRowClicked(td: UtmFieldType, asset: NetScanType) {
  switch (td.field) {
    case AssetFieldEnum.ASSET_SEVERITY:
      break;
    default:
      return asset;
  }
}
