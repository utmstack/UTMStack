import {Component, OnInit} from '@angular/core';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ThemeChangeBehavior} from '../../shared/behaviors/theme-change.behavior';
import {UtmAppThemeService} from '../../shared/services/theme/utm-app-theme.service';
import {AppThemeType} from '../../shared/types/theme/app-theme.type';

@Component({
  selector: 'app-app-theme',
  templateUrl: './app-theme.component.html',
  styleUrls: ['./app-theme.component.scss']
})
export class AppThemeComponent implements OnInit {
  themeImages: AppThemeType[];
  loading = true;

  constructor(private utmAppThemeService: UtmAppThemeService,
              private utmToastService: UtmToastService,
              private themeChangeBehavior: ThemeChangeBehavior) {
  }

  ngOnInit() {
    this.loadConfigs();
  }

  loadConfigs() {
    this.utmAppThemeService.getTheme({page: 0, size: 1000}).subscribe(response => {
      this.themeImages = response.body;
      this.loading = false;
    });
  }

  setDefault() {
    this.utmAppThemeService.resetTheme().subscribe(response => {
      this.utmToastService.showSuccessBottom('Theme changed successfully');
      this.themeChangeBehavior.$themeChange.next(true);
      this.loading = false;
      this.loadConfigs();
    });
  }
}
