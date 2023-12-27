import {Axis} from '../chart-properties/axis/axis';
import {DataZoom} from '../chart-properties/datazoom/data-zoom';
import {Grid} from '../chart-properties/grid/grid';
import {Legend} from '../chart-properties/legend/legend';
import {SeriesHeatMap} from '../chart-properties/series/heatmap/series-heat-map';
import {Toolbox} from '../chart-properties/toolbox/toolbox';
import {Tooltip} from '../chart-properties/tooltip/tooltip';
import {VisualMap} from '../chart-properties/visualmap/visual-map';

export class HeatMapPropertiesType {
  legend?: Legend;
  toolbox?: Toolbox;
  xAxis?: Axis;
  yAxis?: Axis;
  series?: SeriesHeatMap[];
  color?: string[];
  grid?: Grid;
  tooltip?: Tooltip;
  dataZoom?: DataZoom;
  visualMap?: VisualMap;
}
