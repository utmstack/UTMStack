import {Component, EventEmitter, HostListener, Input, OnInit, Output} from '@angular/core';
import {ElasticDataTypesEnum} from '../../../../shared/enums/elastic-data-types.enum';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';

@Component({
  selector: 'app-log-analyzer-field-card',
  templateUrl: './log-analyzer-field-card.component.html',
  styleUrls: ['./log-analyzer-field-card.component.scss']
})
export class LogAnalyzerFieldCardComponent implements OnInit {
  @Input() field: ElasticSearchFieldInfoType;
  @Input() pattern: string;
  @Input() filters: ElasticFilterType[];
  @Output() addFieldToColumn = new EventEmitter<ElasticSearchFieldInfoType>();
  @Input() fieldSelected: ElasticSearchFieldInfoType[];
  public isCollapsed = false;
  viewAddButton = false;
  fieldWidth: string;

  constructor() {
  }

  ngOnInit() {
  }

  @HostListener('width')
  widthChange($event) {
  }

  resolveIcon(): string {
    switch (this.field.type) {
      case ElasticDataTypesEnum.TEXT:
        return 'icon-text-color icon-field';
      case ElasticDataTypesEnum.LONG:
        return 'icon-hash icon-field';
      case ElasticDataTypesEnum.FLOAT:
        return 'icon-hash icon-field';
      case ElasticDataTypesEnum.DATE:
        return 'icon-calendar52 icon-field';
      case ElasticDataTypesEnum.OBJECT:
        return 'icon-circle-css icon-field';
      case ElasticDataTypesEnum.BOOLEAN:
        return 'icon-split icon-field';
      default:
        return 'icon-question3 icon-field';
    }
  }

  addToColumns(field: ElasticSearchFieldInfoType) {
    this.addFieldToColumn.emit(field);
  }

  isInColumns(field: ElasticSearchFieldInfoType): boolean {
    return this.fieldSelected.findIndex(value => value.name === field.name) !== -1;
  }

  processField(name: string) {
    return name.replace('.keyword', '');
  }
}
