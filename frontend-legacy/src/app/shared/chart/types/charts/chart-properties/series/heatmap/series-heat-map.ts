import {SeriesMark} from '../series-properties/marks/series-mark';

export class SeriesHeatMap {
  metricId?: number;
  name?: string;
  type?: 'heatmap';
  data?: any[];
  markPoint?: SeriesMark;
  markLine?: SeriesMark;
}
