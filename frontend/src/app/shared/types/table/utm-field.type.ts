import {ElasticDataTypesEnum} from '../../enums/elastic-data-types.enum';

export class UtmFieldType {
  label?: string;
  field?: string;
  type?: ElasticDataTypesEnum;
  visible?: boolean;
  customStyle?: string;

  constructor(label?: string,
              field?: string,
              type?: ElasticDataTypesEnum,
              visible?: boolean,
              customStyle?: string) {
    this.label = label;
    this.field = field;
    this.type = type;
    this.visible = visible;
    this.customStyle = customStyle;
  }
}
