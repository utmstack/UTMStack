export class SeriesScatter {
  name?: string;
  type?: 'scatter' | 'effectScatter';
  coordinateSystem?: 'leaflet';
  data?: any[];
  rippleEffect?: {
    brushType?: 'stroke' | 'fill'
  };
  showEffectOn?: 'render';
  hoverAnimation?: true;
  symbolSize?: any;
  label?: {
    normal?: {
      formatter?: '{b}';
      position?: 'right';
      show?: false
    };
    emphasis?: {
      show?: true
    }
  };
  itemStyle?: {
    normal?: {
      color?: string
      shadowBlur?: number,
      shadowColor?: string
    }
  };
  zlevel?: number;
}
