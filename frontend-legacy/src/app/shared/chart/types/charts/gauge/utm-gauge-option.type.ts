import {Grid} from '../chart-properties/grid/grid';
import {SeriesGauge} from '../chart-properties/series/gauge/series-gauge';
import {Toolbox} from '../chart-properties/toolbox/toolbox';
import {Tooltip} from '../chart-properties/tooltip/tooltip';

export class UtmGaugeOptionType {
  grid?: Grid;
  tooltip?: Tooltip;
  toolbox?: Toolbox;
  serie?: SeriesGauge[];
}
