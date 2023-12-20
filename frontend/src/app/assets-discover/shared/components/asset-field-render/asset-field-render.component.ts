import {Component, Input, OnInit} from '@angular/core';
import {UtmDateFormatEnum} from '../../../../shared/enums/utm-date-format.enum';
import {UtmFieldType} from '../../../../shared/types/table/utm-field.type';
import {AssetFieldEnum} from '../../enums/asset-field.enum';
import {NetScanType} from '../../types/net-scan.type';

@Component({
  selector: 'app-asset-field-render',
  templateUrl: './asset-field-render.component.html',
  styleUrls: ['./asset-field-render.component.scss']
})
export class AssetFieldRenderComponent implements OnInit {
  @Input() data: NetScanType;
  @Input() field: UtmFieldType;
  utmFormatDate = UtmDateFormatEnum.UTM_SHORT_UTC;
  assetFieldEnum = AssetFieldEnum;

  constructor() {
  }

  ngOnInit() {
  }

}
