import {Directive, ViewContainerRef} from '@angular/core';

@Directive({
  selector: '[appTabContent]'
})
export class TabContentDirective {

  constructor(public viewContainerRef: ViewContainerRef) {
  }

}
