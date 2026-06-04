import {Color} from '../chart-properties/color/color';
import {DataZoom} from '../chart-properties/datazoom/data-zoom';
import {Grid} from '../chart-properties/grid/grid';
import {Legend} from '../chart-properties/legend/legend';
import {Toolbox} from '../chart-properties/toolbox/toolbox';

export class UtmBarOptionType {
  legend?: Legend;
  colors?: Color;
  toolbox?: Toolbox;
  grid?: Grid;
  dataZoom?: DataZoom;
}
