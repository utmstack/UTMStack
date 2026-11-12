import {Directive, ViewContainerRef} from '@angular/core';

@Directive({
    standalone: false,
    selector: '[appTableDetailContent]',
})
export class TableDetailContentDirective {

  constructor(public viewContainerRef: ViewContainerRef) {
  }

}
