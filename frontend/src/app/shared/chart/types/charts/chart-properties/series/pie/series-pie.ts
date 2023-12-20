import {ItemStyle} from '../../style/item-style';

export class SeriesPie {
  name?: string;
  data: { name: string, itemStyle?: any, value: number }[];
  type: 'pie' | 'line' | 'bar';
  radius?: string[];
  center?: string[];
  itemStyle?: ItemStyle;
  roseType?: 'radius' | 'area' | null;

  constructor(data: { name: string, itemStyle?: any, value: number }[],
              type: 'pie' | 'line' | 'bar',
              name?: string,
              radius?: string[],
              center?: string[],
              itemStyle?: ItemStyle,
              roseType?: 'radius' | 'area' | null) {
    this.type = type ? type : null;
    this.name = name ? name : null;
    this.radius = radius ? radius : null;
    this.center = center ? center : null;
    this.itemStyle = itemStyle ? itemStyle : null;
    this.data = data ? data : null;
    this.roseType = roseType ? roseType : null;
  }
}
