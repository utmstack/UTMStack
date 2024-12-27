import {ElasticDataTypesEnum} from '../../enums/elastic-data-types.enum';

export class UtmFieldType {
  label?: string;
  field?: string;
  filterField?: string;
  type?: ElasticDataTypesEnum;
  visible?: boolean;
  customStyle?: string;
  filter?: boolean;
  width?: string;
  fields?: UtmFieldType[];

  static findFieldByNameInArray(fieldName: string, fields: UtmFieldType[]): UtmFieldType | null {
    for (const field of fields) {
      // Check if the current field matches the fieldName
      if (field.field === fieldName) {
        return field;
      }

      // If there are nested fields, search recursively
      if (field.fields && field.fields.length > 0) {
        const foundField = UtmFieldType.findFieldByNameInArray(fieldName, field.fields);
        if (foundField) {
          return foundField;
        }
      }
    }
    return null;
  }
}
