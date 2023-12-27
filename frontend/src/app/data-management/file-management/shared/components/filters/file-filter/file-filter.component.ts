import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {ElasticOperatorsEnum} from '../../../../../../shared/enums/elastic-operators.enum';
import {ElasticFilterType} from '../../../../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../../../../shared/types/table/utm-field.type';
import {ACCESS_MASK_CODES} from '../../../const/file-acces-mask.constant';
import {FILE_FILTER_FIELDS} from '../../../const/file-field.constant';
import {FileFieldEnum} from '../../../enum/file-field.enum';
import {FileAccessMaskCodeType} from '../../../types/file-access-mask-code.type';

@Component({
  selector: 'app-file-filter',
  templateUrl: './file-filter.component.html',
  styleUrls: ['./file-filter.component.scss']
})
export class FileFilterComponent implements OnInit {
  fieldFilters: UtmFieldType[] = FILE_FILTER_FIELDS;
  accesses: FileAccessMaskCodeType[] = ACCESS_MASK_CODES;
  @Output() filterChange = new EventEmitter<ElasticFilterType>();
  @Output() filterReset = new EventEmitter<boolean>();
  access: any;
  fileFieldEnum = FileFieldEnum;

  constructor() {
  }

  ngOnInit() {
  }

  resetAllFilters() {
    this.filterReset.emit(true);
  }

  onFileSearch($event: string) {
    const filter: ElasticFilterType = {
      field: 'logx.wineventlog.*',
      operator: ElasticOperatorsEnum.IS_IN_FIELD,
      value: $event
    };
    this.filterChange.emit(filter);
  }

  onFilterGenericChange($event: ElasticFilterType) {
    this.filterChange.emit($event);
  }

  onAccessChange($event: any[]) {
    const hexs = $event.map(value => {
      return value.hex;
    });
    const filterAcction: ElasticFilterType = {
      field: this.fileFieldEnum.FILE_ACCESS_MASK_FIELD,
      operator: ElasticOperatorsEnum.IS_ONE_OF,
      value: $event ? hexs : null
    };
    this.filterChange.emit(filterAcction);
  }
}
