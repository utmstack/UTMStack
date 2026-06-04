import {Component, Input, OnInit} from '@angular/core';
import {ElasticOperatorsEnum} from '../../../../../../shared/enums/elastic-operators.enum';
import {ElasticFilterType} from '../../../../../../shared/types/filter/elastic-filter.type';
import {FileFieldEnum} from '../../../enum/file-field.enum';
import {resolveFileFieldNameByFilter} from '../../../util/file-util-functions';

@Component({
  selector: 'app-file-active-filters',
  templateUrl: './file-active-filters.component.html',
  styleUrls: ['./file-active-filters.component.scss']
})
export class FileActiveFiltersComponent implements OnInit {
  @Input() filters: ElasticFilterType[] = [];
  operatorsEnum = ElasticOperatorsEnum;
  FILE_EVENT_ID_FIELD = FileFieldEnum.FILE_EVENT_ID_FIELD;
  FILE_OBJECT_TYPE_FIELD = FileFieldEnum.FILE_OBJECT_TYPE_FIELD;

  constructor() {
  }

  ngOnInit() {
  }

  resolveFieldName(filter: ElasticFilterType) {
    return resolveFileFieldNameByFilter(filter);
  }

}
