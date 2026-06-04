import {ChangeDetectionStrategy, Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';
import {
  AlertFieldService
} from '../../../../../../data-management/alert-management/shared/services/alert-field.service';
import {IndexFieldController} from '../../../../../../log-analyzer/shared/behaviors/index-field-controller.behavior';
import {UtmFieldType} from '../../../../../types/table/utm-field.type';

@Component({
  selector: 'app-utm-dtable-columns',
  templateUrl: './utm-dtable-columns.component.html',
  styleUrls: ['./utm-dtable-columns.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class UtmDtableColumnsComponent implements OnInit, OnChanges {
  /**
   * UtmTableFieldType Array
   */
  @Input() fields: UtmFieldType[];
  activeFields: any[];
  inactiveFields: any[];
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
  order: 'asc' | 'desc' = 'asc';
  searchTerm = '';

  constructor(private indexFieldController: IndexFieldController,
              private alertFieldService: AlertFieldService) {
  }

  ngOnInit() {
    this.innerFields = this.fields.slice();
  }

  ngOnChanges(changes: SimpleChanges): void {
    this.fields = this.fields.reduce((acc: UtmFieldType[], field: UtmFieldType) => {
      if (field.fields && field.fields.length > 0) {
        return acc.concat(field.fields);
      } else {
        acc.push(field);
        return acc;
      }
    }, []);
    this.loadInactiveFields();
    this.loadActiveFields();
  }

/*  public removeItem(item: UtmFieldType): void {
    if (this.showInactive) {
      this.fields[this.fields.indexOf(item)].visible = false;
    } else {
      this.fields.splice(this.fields.indexOf(item), 1);
      this.indexFieldController.$field.next(item.field);
    }
  }*/

  removeItem(item: UtmFieldType): void {
    const findAndUpdateItem = (fields: UtmFieldType[], itemToRemove: UtmFieldType): boolean => {
      for (let i = 0; i < fields.length; i++) {
        const field = fields[i];
        if (field === itemToRemove) {
            if (this.showInactive) {
              field.visible = false;
              this.alertFieldService.update(field);
            } else {
              fields.splice(i, 1);
              this.indexFieldController.$field.next(field.field);
            }
            return true;
        }
        if (field.fields) {
          const found = findAndUpdateItem(field.fields, itemToRemove);
          if (found) {
            return true;
          }
        }
      }
      return false;
    };
    if (findAndUpdateItem(this.fields, item)) {
      this.loadActiveFields();
      this.loadInactiveFields();
    }
  }

  setVisibleItem(item: UtmFieldType) {
    this.fields.forEach((field: UtmFieldType) => {
      if (field.field === item.field) {
        field.visible = true;
      }
    });
    this.loadActiveFields();
    this.loadInactiveFields();
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

  loadInactiveFields() {
    this.inactiveFields = this.fields.filter(f => !f.visible)
      .map(f => this.mapFieldsToItem(f));

    return this.inactiveFields;
  }

  loadActiveFields(){
    this.activeFields = this.fields.filter(f => f.visible)
     .map(f => this.mapFieldsToItem(f));
    return this.activeFields;
  }

  findField($event: { id: string; }) {
    return UtmFieldType.findFieldByNameInArray($event.id, this.fields);
  }

  isVisible(item: any) {
    const itemVisible = UtmFieldType.findFieldByNameInArray(item.id, this.fields);
    if (itemVisible.visible) {
      return true;
    }
  }

  setOrder(order: 'asc' | 'desc') {
    this.order = order;

    this.activeFields = [...this.activeFields].sort((a, b) => {
      if (order === 'asc') {
        return a.name.localeCompare(b.name);
      } else {
        return b.name.localeCompare(a.name);
      }
    });
  }


  onChange(searchTerm: string) {
    if (searchTerm.trim().length > 0) {
      this.activeFields = this.activeFields.filter(f => f.name.toLowerCase().includes(searchTerm));
    } else {
      this.loadActiveFields();
    }
  }

  mapFieldsToItem(field: UtmFieldType) {
    if (!field.fields) {
      return {
        id: field.field,
        name: field.label ? field.label : field.field,
      };
    } else {
      return {
        id: field.field,
        name: field.label ? field.label : field.field,
        items: field.fields.
          map(childField => this.mapFieldsToItem(childField))
      };
    }
  }
}
