import {ToolboxFeature} from './toolbox-feature';
import {ToolboxIconStyle} from './toolbox-icon-style';

export class Toolbox {
  show?: boolean;
  feature?: ToolboxFeature;
  orient?: 'horizontal' | 'vertical';
  itemSize?: number;
  showTitle?: boolean;
  left?: 'left' | 'center' | 'right';
  top?: 'top' | 'middle' | 'bottom';
  width?: number;
  height?: number;
  iconStyle?: ToolboxIconStyle;


  constructor(show?: boolean,
              feature?: ToolboxFeature,
              orient?: 'horizontal' | 'vertical',
              itemSize?: number,
              showTitle?: boolean,
              left?: 'left' | 'center' | 'right',
              top?: 'top' | 'middle' | 'bottom',
              width?: number,
              height?: number,
              iconStyle?: ToolboxIconStyle) {
    this.show = show;
    this.feature = feature ? feature : null;
    this.orient = orient ? orient : 'horizontal';
    this.itemSize = itemSize ? itemSize : 14;
    this.showTitle = showTitle ? showTitle : false;
    this.left = left ? left : 'right';
    this.top = top ? top : 'top';
    this.width = width ? width : null;
    this.height = height ? height : null;
    this.iconStyle = iconStyle ? iconStyle : null;
  }
}
