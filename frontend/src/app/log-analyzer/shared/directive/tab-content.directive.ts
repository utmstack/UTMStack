import {Directive, ViewContainerRef} from '@angular/core';

@Directive({
    standalone: false,
    selector: '[appTabContent]',
})
export class TabContentDirective {

  constructor(public viewContainerRef: ViewContainerRef) {
  }

}
