import {Component, Input, OnInit} from '@angular/core';
import {IndexFieldController} from '../../../../../../log-analyzer/shared/behaviors/index-field-controller.behavior';
import {UtmFieldType} from '../../../../../types/table/utm-field.type';

@Component({
  selector: 'app-utm-dtable-columns',
  templateUrl: './utm-dtable-columns.component.html',
  styleUrls: ['./utm-dtable-columns.component.scss']
})
export class UtmDtableColumnsComponent implements OnInit {
  /**
   * UtmTableFieldType Array
   */
  @Input() fields: UtmFieldType[];
  /**
   * String icon, use iconmoon
   */
  @Input() icon: string;
  /**
   * If true show panel Inactive columns and operator to change visibility.
   * If false when hide element will delete from UtmTableFieldType Array
   */
  @Input() showInactive: boolean;
  /**
   * Label to show in header
   */
  @Input() label: string;
  inactiveSelected: UtmFieldType;
  activeSelected: UtmFieldType;

  innerFields: UtmFieldType[];
  constructor(private indexFieldController: IndexFieldController) {
  }

  ngOnInit() {
    this.innerFields = this.fields.slice();
  }

  public removeItem(item: UtmFieldType): void {
    if (this.showInactive) {
      this.fields[this.fields.indexOf(item)].visible = false;
    } else {
      this.fields.splice(this.fields.indexOf(item), 1);
      this.indexFieldController.$field.next(item.field);
    }
  }

  setVisibleItem(item: UtmFieldType) {
    this.fields[this.fields.indexOf(item)].visible = true;
  }

  selectInactive(item: UtmFieldType) {
    this.inactiveSelected = item;
  }

  isInactiveSelected(item: UtmFieldType) {
    return this.inactiveSelected && this.inactiveSelected.field === item.field;
  }

  getFieldsVisible(): string[] {
    return this.fields.filter(value => value.visible === true).map(value => value.label);
  }
}
