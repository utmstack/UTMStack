import {AfterViewInit, Component, OnInit} from '@angular/core';
import {ActivatedRoute, NavigationEnd, Router} from '@angular/router';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {RestartApiBehavior} from '../../shared/behaviors/restart-api.behavior';
import {UtmConfigParamsService} from '../../shared/services/config/utm-config-params.service';
import {UtmConfigSectionService} from '../../shared/services/config/utm-config-section.service';
import {SectionConfigParamType} from '../../shared/types/configuration/section-config-param.type';
import {SectionConfigType} from '../../shared/types/configuration/section-config.type';

@Component({
  selector: 'app-app-config',
  templateUrl: './app-config.component.html',
  styleUrls: ['./app-config.component.scss']
})
export class AppConfigComponent implements OnInit {
  sections: SectionConfigType[] = [];
  private loading = true;
  confValid = true;
  saving: any;
  configToSave: SectionConfigParamType[] = [];
  fragment: string | null;

  constructor(private utmConfigSectionService: UtmConfigSectionService,
              private route: ActivatedRoute,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.route.queryParams.subscribe( params => {
      this.getSections(params.sections ? JSON.parse(params.sections) : []);
      this.fragment = params && params.id ? params.id : null;
    });
  }

  scrollToFragment(fragment: string): void {
    const element = document.getElementById(fragment);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
    }
  }
  getSections(sections: number[] ) {
    this.loading = true;
    this.utmConfigSectionService.query({page: 0, size: 10000, 'moduleNameShort.specified': false}).subscribe(response => {
      this.loading = false;
      this.sections = sections.length > 0 ? sections.map(id => response.body.find(s => s.id === id)) : response.body;
      setTimeout(() => {
        if (this.fragment) {
          this.scrollToFragment(this.fragment);
        }
      }, 300);
    }, error => {
      this.toastService.showError('Error', 'Error getting application configurations sections');
    });
  }

}
