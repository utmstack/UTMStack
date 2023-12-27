export class Grid {
  left?: 'left' | 'center' | 'right' | string | number;
  top?: 'top' | 'middle' | 'bottom' | string | number;
  right?: string | number;
  bottom?: string | number;
  width?: string | number;
  height?: string | number;
  backgroundColor?: string;

  constructor(left?: 'left' | 'center' | 'right' | string | number,
              top?: 'top' | 'middle' | 'bottom' | string | number,
              right?: string | number,
              bottom?: string | number,
              width?: string | number,
              height?: string | number,
              backgroundColor?: string) {
    this.left = left ? left : 0;
    this.top = top ? top : 0;
    this.right = right ? right : 0;
    this.bottom = bottom ? bottom : 0;
    this.width = width ? width : null;
    this.height = height ? height : null;
    this.backgroundColor = backgroundColor ? backgroundColor : null;
  }
}
