import { Directive } from '@angular/core';

@Directive({
  selector: 'ng-template[appTemplateSelector]'
})
export class TemplateSelectorDirective {

  constructor() { }

}
