import {Axis} from '../chart-properties/axis/axis';
import {DataZoom} from '../chart-properties/datazoom/data-zoom';
import {Grid} from '../chart-properties/grid/grid';
import {Legend} from '../chart-properties/legend/legend';
import {SeriesLine} from '../chart-properties/series/line/series-line';
import {Toolbox} from '../chart-properties/toolbox/toolbox';
import {Tooltip} from '../chart-properties/tooltip/tooltip';

export class UtmLineBarOptionType {
  legend?: Legend;
  toolbox?: Toolbox;
  xAxis?: Axis;
  yAxis?: Axis;
  seriesOption?: SeriesLine[];
  series?: SeriesLine[];
  color?: string[];
  grid?: Grid;
  tooltip?: Tooltip;
  dataZoom?: DataZoom;

  constructor(legend: Legend,
              toolbox: Toolbox,
              xAxis: Axis,
              yAxis: Axis,
              seriesOption: SeriesLine[],
              series: SeriesLine[],
              color: string[],
              grid: Grid,
              tooltip: Tooltip,
              dataZoom: DataZoom) {
    this.legend = legend;
    this.toolbox = toolbox;
    this.xAxis = xAxis;
    this.yAxis = yAxis;
    this.seriesOption = seriesOption;
    this.series = series;
    this.color = color;
    this.grid = grid;
    this.tooltip = tooltip;
    this.dataZoom = dataZoom;
  }
}
