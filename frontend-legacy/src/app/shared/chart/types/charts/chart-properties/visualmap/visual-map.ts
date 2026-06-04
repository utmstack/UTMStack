export class VisualMap {
  type?: 'continuous' | 'piecewise';
  show?: boolean;
  calculable?: boolean;
  realtime?: boolean;
  inRange?: {
    color?: string[]
  };
  orient?: 'horizontal' | 'vertical';
  left?: 'left' | 'center' | 'right';
  top?: 'top' | 'middle' | 'bottom';
  bottom?: string | number;
  min?: number;
  max?: number;


  constructor(type?: 'continuous' | 'piecewise',
              show?: boolean,
              calculable?: boolean,
              realtime?: boolean,
              inRange?: { color?: string[] },
              orient?: 'horizontal' | 'vertical',
              left?: 'left' | 'center' | 'right',
              top?: 'top' | 'middle' | 'bottom',
              bottom?: string | number,
              min?: number,
              max?: number) {
    this.type = type ? type : 'continuous';
    this.show = show;
    this.calculable = calculable ? calculable : true;
    this.realtime = realtime ? realtime : true;
    this.inRange = inRange ? inRange : {color: ['#f5994e', '#c05050']};
    this.orient = orient ? orient : 'horizontal';
    this.left = left ? left : 'center';
    this.top = top ? top : 'bottom';
    this.bottom = bottom ? bottom : '15%';
    this.max = max;
    this.min = min;
  }
}
