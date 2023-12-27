import {ItemStyle} from '../../style/item-style';
import {SeriesMark} from '../series-properties/marks/series-mark';


export class SeriesBar {
  name?: string;
  data: { name?: string, value?: number }[];
  type: 'pie' | 'line' | 'bar';
  itemStyle?: ItemStyle;
  markPoint?: SeriesMark;

  constructor(data: { name?: string, value?: number, data?: any[] }[],
              type: 'pie' | 'line' | 'bar',
              name?: string,
              itemStyle?: ItemStyle,
              markPoint?: SeriesMark) {
    this.type = type ? type : null;
    this.name = name ? name : null;
    this.itemStyle = itemStyle ? itemStyle : null;
    this.data = data ? data : null;
    this.markPoint = markPoint ? markPoint : new SeriesMark();
  }
}
