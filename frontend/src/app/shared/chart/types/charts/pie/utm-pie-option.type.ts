import {Color} from '../chart-properties/color/color';
import {Grid} from '../chart-properties/grid/grid';
import {Legend} from '../chart-properties/legend/legend';
import {Toolbox} from '../chart-properties/toolbox/toolbox';

export class UtmPieOptionType {
  pieType?: 'donut' | 'pie' | string;
  legend?: Legend;
  color?: Color | string[];
  toolbox?: Toolbox;
  grid?: Grid;
}
