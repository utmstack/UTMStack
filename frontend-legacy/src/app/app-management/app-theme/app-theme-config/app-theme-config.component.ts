import {Component, Input, OnInit} from '@angular/core';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ThemeChangeBehavior} from '../../../shared/behaviors/theme-change.behavior';
import {AppThemeLocationEnum} from '../../../shared/enums/app-theme-location.enum';
import {UtmAppThemeService} from '../../../shared/services/theme/utm-app-theme.service';
import {AppThemeType} from '../../../shared/types/theme/app-theme.type';

@Component({
  selector: 'app-app-theme-config',
  templateUrl: './app-theme-config.component.html',
  styleUrls: ['./app-theme-config.component.css']
})
export class AppThemeConfigComponent implements OnInit {
  @Input() theme: AppThemeType;
  place = AppThemeLocationEnum;
  saving = false;
  changed = false;

  constructor(private utmAppThemeService: UtmAppThemeService,
              private themeChangeBehavior: ThemeChangeBehavior,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  onUpload($event: any[]) {
  }

  onImageChange($event: any) {
    this.theme.userImg = $event;
    this.changed = true;
  }

  saveImage() {
    this.saving = true;
    this.utmAppThemeService.update(this.theme).subscribe(response => {
      this.utmToastService.showSuccessBottom('Theme changed successfully');
      this.themeChangeBehavior.$themeChange.next(true);
      this.saving = false;
      this.changed = false;
    }, error => {
      this.saving = false;
      this.utmToastService.showError('Error saving theme', 'Error while trying to save application theme');
    });
  }

  getDescription() {
    switch (this.theme.shortName) {
      case AppThemeLocationEnum.HEADER:
        return 'Accepted extensions: .png,.svg,.jpg<br> Recommended size: 125x18 px';
      case AppThemeLocationEnum.LOGIN:
        return 'Accepted extensions: .png,.svg,.jpg<br> Recommended size: 100x130 px';
      case AppThemeLocationEnum.REPORT:
        return 'Accepted extensions: .png,.svg,.jpg<br> Recommended size: 30x40 px';
      case AppThemeLocationEnum.REPORT_COVER:
        return 'Accepted extensions: .png,.svg,.jpg<br> Recommended size: 1100x1600 px';
    }
  }
}
