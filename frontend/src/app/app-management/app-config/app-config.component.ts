import {Component, OnInit} from '@angular/core';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {RestartApiBehavior} from '../../shared/behaviors/restart-api.behavior';
import {UtmConfigParamsService} from '../../shared/services/config/utm-config-params.service';
import {UtmConfigSectionService} from '../../shared/services/config/utm-config-section.service';
import {SectionConfigParamType} from '../../shared/types/configuration/section-config-param.type';
import {SectionConfigType} from '../../shared/types/configuration/section-config.type';
import {ActivatedRoute} from "@angular/router";

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

  constructor(private utmConfigSectionService: UtmConfigSectionService,
              private route: ActivatedRoute,
              private utmConfigParamsService: UtmConfigParamsService,
              private restartApiBehavior: RestartApiBehavior,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.route.queryParams.subscribe( params => {
      this.getSections(params.sections ? JSON.parse(params.sections) : []);
    });
  }
  getSections(sections: number[] ) {
    this.loading = true;
    this.utmConfigSectionService.query({page: 0, size: 10000, 'moduleNameShort.specified': false}).subscribe(response => {
      this.loading = false;
      console.log(response.body);
      this.sections = sections.length > 0 ? sections.map(id => response.body.find(s => s.id === id)) : response.body;
    }, error => {
      this.toastService.showError('Error', 'Error getting application configurations sections');
    });
  }

}
