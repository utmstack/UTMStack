import {Axis} from '../chart-properties/axis/axis';
import {UtmBarDataType} from './utm-bar-data.type';
import {UtmBarOptionType} from './utm-bar-option.type';

export class UtmBarProperties {
  data?: UtmBarDataType;
  options?: UtmBarOptionType;
  xAxis?: Axis;
  yAxis?: Axis;
}
