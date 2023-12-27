import {UTM_COLOR_THEME} from '../../../../../constants/utm-color.const';

export class Color {
  color?: string[];

  constructor(color?: string[]) {
    this.color = color ? color : UTM_COLOR_THEME;
  }
}
