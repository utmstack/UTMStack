import {Directive, ElementRef, Input, Renderer2} from '@angular/core';
import {AppLogTypeEnum} from '../../../app-management/app-logs/shared/enum/app-log-type.enum';

@Directive({
  selector: '[appBadgeType]'
})
export class BadgeTypeDirective {

  constructor(private el: ElementRef,
              private renderer: Renderer2) { }

  @Input() set appBadgeType(type: string) {
    this.removePreviousClasses();
    const classNames = this.getClassByType(type);
    classNames.split(' ').forEach(cls => this.renderer.addClass(this.el.nativeElement, cls));
  }
  getClassByType(type: string) {
    switch (type) {
      case AppLogTypeEnum.ERROR:
        return 'border-danger-600 text-danger-600';
      case AppLogTypeEnum.INFO:
        return 'border-info-600 text-info-600';
      case AppLogTypeEnum.WARNING:
        return 'border-warning text-warning';
    }
  }
  private removePreviousClasses() {
    const classesToRemove = [
      'border-danger-600', 'text-danger-600',
      'border-info-600', 'text-info-600',
      'border-warning', 'text-warning'
    ];

    classesToRemove.forEach(cls => this.renderer.removeClass(this.el.nativeElement, cls));
  }
}
