import {Axis} from '../../types/charts/chart-properties/axis/axis';
import {Color} from '../../types/charts/chart-properties/color/color';
import {DataZoom} from '../../types/charts/chart-properties/datazoom/data-zoom';
import {Grid} from '../../types/charts/chart-properties/grid/grid';
import {Legend} from '../../types/charts/chart-properties/legend/legend';
import {SeriesBar} from '../../types/charts/chart-properties/series/bar/series-bar';
import {SeriesGauge} from '../../types/charts/chart-properties/series/gauge/series-gauge';
import {SeriesHeatMap} from '../../types/charts/chart-properties/series/heatmap/series-heat-map';
import {SeriesLine} from '../../types/charts/chart-properties/series/line/series-line';
import {SeriesPie} from '../../types/charts/chart-properties/series/pie/series-pie';
import {SeriesScatter} from '../../types/charts/chart-properties/series/scatter/series-scatter';
import {SeriesTagCloud} from '../../types/charts/chart-properties/series/tagcloud/series-tag-cloud';
import {Toolbox} from '../../types/charts/chart-properties/toolbox/toolbox';
import {Tooltip} from '../../types/charts/chart-properties/tooltip/tooltip';
import {LeafletMapType} from '../../types/map/leaflet/leaflet-map.type';

export class ChartOption {
  toolbox?: Toolbox;
  color?: Color | string[];
  tooltip?: Tooltip;
  legend?: Legend;
  series?: SeriesPie | SeriesBar[] | SeriesGauge[] | SeriesTagCloud[] | SeriesLine[] | SeriesScatter[] | SeriesHeatMap[];
  calculable?: boolean;
  grid?: Grid;
  xAxis?: Axis;
  yAxis?: Axis;
  dataZoom?: DataZoom;
  leaflet?: LeafletMapType;

  constructor(series: SeriesPie | SeriesBar[] | SeriesGauge[],
              tooltip?: Tooltip,
              legend?: Legend,
              color?: Color,
              toolbox?: Toolbox,
              calculable?: boolean,
              grid?: Grid,
              xAxis?: Axis,
              yAxis?: Axis,
              dataZoom?: any) {
    this.color = color;
    this.tooltip = tooltip;
    this.legend = legend;
    this.series = series;
    this.toolbox = toolbox;
    this.calculable = calculable;
    this.grid = grid ? grid : new Grid();
    this.xAxis = xAxis ? xAxis : null;
    this.yAxis = yAxis ? yAxis : null;
    this.dataZoom = dataZoom ? dataZoom : null;
  }

}
