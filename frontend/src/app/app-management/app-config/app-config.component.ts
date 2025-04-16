import {Component, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {UtmConfigSectionService} from '../../shared/services/config/utm-config-section.service';
import {NetworkService} from '../../shared/services/network.service';
import {SectionConfigParamType} from '../../shared/types/configuration/section-config-param.type';
import {ApplicationConfigSectionEnum, SectionConfigType} from '../../shared/types/configuration/section-config.type';

@Component({
  selector: 'app-app-config',
  templateUrl: './app-config.component.html',
  styleUrls: ['./app-config.component.scss']
})
export class AppConfigComponent implements OnInit, OnDestroy {
  sections: SectionConfigType[] = [];
  private loading = true;
  confValid = true;
  saving: any;
  configToSave: SectionConfigParamType[] = [];
  fragment: string | null;
  destroy$ = new Subject<void>();
  isOnline = false;

  constructor(private utmConfigSectionService: UtmConfigSectionService,
              private route: ActivatedRoute,
              private toastService: UtmToastService,
              private networkService: NetworkService) {
  }

  ngOnInit() {
    this.route.queryParams.subscribe( params => {
      this.getSections(params.sections ? JSON.parse(params.sections) : []);
      this.fragment = params && params.id ? params.id : null;
    });

    this.networkService.isOnline$
      .pipe(takeUntil(this.destroy$))
      .subscribe( isOnline => this.isOnline = isOnline);
  }

  scrollToFragment(fragment: string): void {
    const element = document.getElementById(fragment);
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
    }
  }
  getSections(sections: number[] ) {
    this.loading = true;
    this.utmConfigSectionService.query({
      page: 0,
      size: 10000,
      sort: 'id,asc',
      'moduleNameShort.specified': false
    }).subscribe(response => {
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

  getActiveSection(){
    if (this.isOnline) {
      return this.sections.filter(section => section.id !== ApplicationConfigSectionEnum.SYSTEM_LICENSE);
    } else {
      return this.sections;
    }
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

}
