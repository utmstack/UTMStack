import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {ElasticOperatorsEnum} from '../../../../../../shared/enums/elastic-operators.enum';
import {ElasticFilterType} from '../../../../../../shared/types/filter/elastic-filter.type';
import {FileFiltersBehavior} from '../../../behavior/file-filters.behavior';
import {FILE_FILTER_FIELDS} from '../../../const/file-field.constant';
import {FileFieldEnum} from '../../../enum/file-field.enum';
import {resolveFileFieldNameByFilter} from '../../../util/file-util-functions';

@Component({
  selector: 'app-file-filter-applied',
  templateUrl: './file-filter-applied.component.html',
  styleUrls: ['./file-filter-applied.component.scss']
})
export class FileFilterAppliedComponent implements OnInit {
  filtersActive: ElasticFilterType[] = [];
  operatorEnum = ElasticOperatorsEnum;
  filters: ElasticFilterType[] = [];
  @Output() filterAppliedChange = new EventEmitter<{ filter: ElasticFilterType, valueDelete: string }>();

  constructor(private fileFiltersBehavior: FileFiltersBehavior) {
  }

  ngOnInit() {
    this.fileFiltersBehavior.$fileFilters.subscribe(filters => {
      if (filters) {
        this.filtersActive = filters.filter(value =>
          value.field !== FileFieldEnum.FILE_TIMESTAMP_FIELD
          && value.field !== FileFieldEnum.FILE_ACCESS_MASK_FIELD
          && value.field !== FileFieldEnum.FILE_EVENT_ID_FIELD
          && value.field !== FileFieldEnum.FILE_OBJECT_TYPE_FIELD);
        /**
         * If field is not visible then set visible filter
         */
        for (const filter of this.filtersActive) {
          const indexFilter = FILE_FILTER_FIELDS.findIndex(value => value.field === filter.field);
          if (indexFilter !== -1) {
            if (!FILE_FILTER_FIELDS[indexFilter].visible) {
              FILE_FILTER_FIELDS[indexFilter].visible = true;
            }
          }
        }
      }
    });
  }

  resolveFilterName(filter: ElasticFilterType): string {
    return resolveFileFieldNameByFilter(filter);
  }

  onDelete(filter: ElasticFilterType, value?: string): Promise<ElasticFilterType> {
    return new Promise<ElasticFilterType>(resolve => {
      if (value) {
        const indexVal = filter.value.findIndex(val => val === value);
        filter.value.splice(indexVal, 1);
        if (filter.value.length === 0) {
          const activeFilterIndex = this.filtersActive.findIndex(fil => fil.field === filter.field);
          if (activeFilterIndex !== -1) {
            this.filtersActive.splice(activeFilterIndex, 1);
          }
        }
      }
      resolve(filter);
    });
  }

  deleteFilter(filter: ElasticFilterType, value?: string) {
    this.onDelete(filter, value).then(fil => {
      this.filterAppliedChange.emit({filter: fil, valueDelete: value});
    });
  }
}
