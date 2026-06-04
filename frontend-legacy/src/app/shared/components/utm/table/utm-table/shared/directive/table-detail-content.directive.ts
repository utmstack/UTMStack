import {Directive, ViewContainerRef} from '@angular/core';

@Directive({
  selector: '[appTableDetailContent]'
})
export class TableDetailContentDirective {

  constructor(public viewContainerRef: ViewContainerRef) {
  }

}
