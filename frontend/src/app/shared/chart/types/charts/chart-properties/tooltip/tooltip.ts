import {TextStyle} from '../style/text-style';

export class Tooltip {
  trigger: 'item' | 'axis' | 'none' | 'category';
  formatter?: any;
  backgroundColor?: string;
  padding?: number;
  textStyle?: TextStyle;

  constructor(trigger: 'item' | 'axis' | 'none' | 'category',
              formatter?: any,
              backgroundColor?: string,
              padding?: number,
              textStyle?: TextStyle) {
    this.formatter = formatter;
    this.trigger = trigger ? trigger : 'item';
    this.backgroundColor = backgroundColor ? backgroundColor : 'rgba(0,75,139,0.86)';
    this.padding = padding ? padding : null;
    this.textStyle = textStyle ? textStyle : new TextStyle(13, 'Roboto, sans-serif');
  }

}
