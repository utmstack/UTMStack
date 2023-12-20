export class Legend {
  show?: boolean;
  data?: any[];
  type?: 'scroll' | 'plain';
  orient?: 'vertical' | 'horizontal';
  top?: 'top' | 'middle' | 'bottom';
  left?: 'left' | 'center' | 'right';
  right?: number;
  bottom?: number;
  itemHeight?: number;
  itemWidth?: number;
  icon: string;
  extraCssText?: string;

  constructor(show: boolean,
              data?: any[],
              type?: 'scroll' | 'plain',
              orient?: 'vertical' | 'horizontal',
              top?: 'top' | 'middle' | 'bottom',
              left?: 'left' | 'center' | 'right',
              right?: number,
              bottom?: number,
              itemHeight?: number,
              itemWidth?: number,
              icon?: string) {
    this.show = show;
    this.data = data;
    this.type = type ? type : 'plain';
    this.orient = orient ? orient : 'vertical';
    this.top = top ? top : 'middle';
    this.left = left ? left : 'left';
    this.right = right ? right : null;
    this.bottom = bottom ? bottom : null;
    this.itemHeight = itemHeight ? itemHeight : 8;
    this.itemWidth = itemWidth ? itemWidth : 8;
    this.icon = icon ? icon : null;
  }
}
