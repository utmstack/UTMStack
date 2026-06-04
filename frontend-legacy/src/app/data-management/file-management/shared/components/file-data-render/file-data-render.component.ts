import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmDateFormatEnum} from '../../../../../shared/enums/utm-date-format.enum';
import {UtmFieldType} from '../../../../../shared/types/table/utm-field.type';
import {extractValueFromObjectByPath} from '../../../../../shared/util/get-value-object-from-property-path.util';
import {FileFieldEnum} from '../../enum/file-field.enum';

@Component({
  selector: 'app-file-data-render',
  templateUrl: './file-data-render.component.html',
  styleUrls: ['./file-data-render.component.scss']
})
export class FileDataRenderComponent implements OnInit {
  @Input() file: any;
  @Input() field: UtmFieldType;
  @Output() refreshData = new EventEmitter<boolean>();
  utmFormatDate = UtmDateFormatEnum.UTM_SHORT_UTC;
  fileFieldEnum = FileFieldEnum;

  constructor() {
  }

  ngOnInit() {
  }

  resolveValue(file: any, td: UtmFieldType) {
    return extractValueFromObjectByPath(file, td);
  }

}
