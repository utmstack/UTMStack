import {Component, Input, OnInit} from '@angular/core';
import {IndexFieldController} from '../../../../../../../log-analyzer/shared/behaviors/index-field-controller.behavior';
import {ElasticOperatorsEnum} from '../../../../../../enums/elastic-operators.enum';
import {convertObjectToKeyValueArray} from '../../../../../../util/get-value-object-from-property-path.util';
import {UtmFilterBehavior} from '../../../../filters/utm-elastic-filter/shared/behavior/utm-filter.behavior';

@Component({
  selector: 'app-analyzer-table-view',
  templateUrl: './utm-table-detail-view.component.html',
  styleUrls: ['./utm-table-detail-view.component.scss']
})
export class UtmTableDetailViewComponent implements OnInit {
  @Input() rowDocument: any;
  @Input() showControl = true;
  tableData: { key: string, value: any }[] = [];

  constructor(private indexFieldController: IndexFieldController,
              private utmFilterBehavior: UtmFilterBehavior) {
  }

  ngOnInit() {
    this.tableData = [];
    if (this.rowDocument) {
      this.tableData = convertObjectToKeyValueArray(this.rowDocument);
    }
  }

  addFieldToColumn(row: { key: string; value: any }) {
    this.indexFieldController.$field.next(row.key);
  }

  fieldIsArray(key): boolean {
    const regex = /\.(\d+)\./g;
    return regex.test(key);
  }

  filterByValue(row: { key: string; value: any }) {
    this.utmFilterBehavior.$filterChange.next(
      {
        field: this.processKey(row.key),
        value: row.value,
        operator: ElasticOperatorsEnum.IS,
        status: 'ACTIVE'
      });
  }

  processKey(key: string): string {
    const regex = /\.(\d+)\./g;
    if (regex.test(key)) {
      return key.replace(regex, '.');
    } else if (!isNaN(Number(key.substring(key.lastIndexOf('.'), key.length)))) {
      return key.substring(0, key.lastIndexOf('.'));
    } else {
      return key;
    }
  }

  isDate(key: string, value: string): boolean {
    return (key.toLowerCase().includes('time') || key.toLowerCase().includes('date')) && !isNaN(new Date(value).getDate());
  }
}
