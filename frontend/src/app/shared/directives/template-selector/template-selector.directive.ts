import { Directive } from '@angular/core';

@Directive({
    standalone: false,
    selector: 'ng-template[appTemplateSelector]',
})
export class TemplateSelectorDirective {

  constructor() { }

}
