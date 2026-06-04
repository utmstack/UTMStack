import {AppThemeLocationEnum} from '../../enums/app-theme-location.enum';

export interface AppThemeType {
  shortName: AppThemeLocationEnum;
  systemImg: string;
  tooltipText: string;
  userImg: string;
}
