import { Directive, Input, OnDestroy, OnInit, TemplateRef, ViewContainerRef } from '@angular/core';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { EnterpriseFeatures, VersionInfoService, VersionType } from 'src/app/shared/services/version/version-info.service';

@Directive({
  selector: '[appIsEnterpriseModule]'
})
export class IsEnterpriseModuleDirective implements OnInit, OnDestroy {
  @Input('appIsEnterpriseModule') module: string;
  @Input('appIsEnterpriseModuleElse') elseTpl: TemplateRef<any> | null = null;
  @Input('appIsEnterpriseModuleMenuName') menuName = '';
  @Input('appIsEnterpriseModuleCssIcon') cssIcon: string = '';

  private destroy$ = new Subject<void>();

  constructor(
    private tpl: TemplateRef<any>,
    private vcr: ViewContainerRef,
    private versionTypeService: VersionInfoService
  ) {}

  ngOnInit() {
    this.versionTypeService.versionType$
      .pipe(takeUntil(this.destroy$))
      .subscribe(versionType => this.render(versionType));
  }

  private render(versionType: VersionType) {
    this.vcr.clear();

    const context = {
      $implicit: this.menuName,
      menuName: this.menuName,
      cssIcon: this.cssIcon
    };

    if (versionType === VersionType.ENTERPRISE && EnterpriseFeatures.includes(this.module)) {
      const viewRef = this.vcr.createEmbeddedView(this.tpl, context);
      viewRef.detectChanges();
    } else if (this.elseTpl) {
      const viewRef = this.vcr.createEmbeddedView(this.elseTpl, context);
      viewRef.detectChanges();
    } else {

    }
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
