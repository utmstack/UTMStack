import {Directive, Input, OnDestroy, OnInit, TemplateRef, ViewContainerRef} from '@angular/core';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import {EnterpriseFeatures, VersionInfoService, VersionType} from 'src/app/shared/services/version/version-info.service';

@Directive({
  selector: '[appIsEnterpriseModule]'
})
export class IsEnterpriseModuleDirective implements OnInit, OnDestroy {
  @Input('appIsEnterpriseModule') module: string;
  @Input('appIsEnterpriseModuleElse') elseTpl: TemplateRef<any> | null = null;

  private destroy$ = new Subject<void>();

  constructor(
    private tpl: TemplateRef<any>,
    private vcr: ViewContainerRef,
    private versionTypeService: VersionInfoService
  ) {}

  ngOnInit() {
    this.versionTypeService.versionType$
      .pipe(takeUntil(this.destroy$))
      .subscribe(versionType => {
        if (versionType === VersionType.ENTERPRISE && EnterpriseFeatures.includes(this.module)) {
          this.vcr.createEmbeddedView(this.tpl);
        } else if (this.elseTpl) {
          this.vcr.createEmbeddedView(this.elseTpl);
        } else {
          this.vcr.clear();
        }
      });
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
