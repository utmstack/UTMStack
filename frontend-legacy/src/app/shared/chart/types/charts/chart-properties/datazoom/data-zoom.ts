export class DataZoom {
  show?: boolean;
  orient?: 'horizontal' | 'vertical';
  type?: 'slider';
  start?: number;
  end?: number;
  height?: number | string;
  width?: number | string;
  borderColor?: string;
  fillerColor?: string;
  handleStyle?: {
    color?: string
  };
  left?: 'left' | 'center' | 'right' | string | number;
  top?: 'top' | 'middle' | 'bottom' | string | number;
  right?: string | number;
  bottom?: string | number;

  constructor(show?: boolean,
              orient?: 'horizontal' | 'vertical',
              type?: 'slider',
              start?: number,
              end?: number,
              height?: number | string,
              width?: number | string,
              borderColor?: string,
              fillerColor?: string,
              handleStyle?: { color?: string },
              left?: 'left' | 'center' | 'right' | string | number,
              top?: 'top' | 'middle' | 'bottom' | string | number,
              right?: string | number,
              bottom?: string | number) {
    this.show = show;
    this.orient = orient ? orient : 'horizontal';
    this.type = type ? type : 'slider';
    this.start = start ? start : 0;
    this.end = end ? end : 100;
    this.height = height ? height : 30;
    this.width = width ? width : null;
    this.borderColor = borderColor ? borderColor : '#ccc';
    this.fillerColor = fillerColor ? fillerColor : 'rgba(40,54,139,0.21)';
    this.handleStyle = handleStyle ? handleStyle : {color: '#004b8b'};
    this.left = left ? left : 'center';
    this.top = top ? top : 'bottom';
    this.right = right ? right : null;
    this.bottom = bottom ? bottom : -20;
  }
}
