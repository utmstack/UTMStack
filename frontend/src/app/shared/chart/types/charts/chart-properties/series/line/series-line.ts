import {ItemStyle} from '../../style/item-style';
import {SeriesMark} from '../series-properties/marks/series-mark';

export class SeriesLine {
  metricId?: number;
  name?: string;
  type?: 'line' | 'bar';
  data?: any;
  smooth?: boolean;
  markPoint?: SeriesMark;
  markLine?: SeriesMark;
  itemStyle?: ItemStyle;


  constructor(metricId?: number,
              name?: string,
              type?: 'line' | 'bar',
              data?: number[],
              smooth?: boolean,
              markPoint?: SeriesMark,
              markLine?: SeriesMark,
              itemStyle?: any) {
    this.metricId = metricId;
    this.name = name;
    this.type = type;
    this.data = data;
    this.smooth = smooth;
    this.markPoint = markPoint;
    this.markLine = markLine;
    this.itemStyle = itemStyle;
  }
}
