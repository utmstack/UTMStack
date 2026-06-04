import {Component, ContentChild, Input, TemplateRef} from '@angular/core';
import {Step} from '../step';
import {StepDirective} from './step.component';

@Component({
  selector: 'app-utm-list',
  template: `
    <ng-container *ngFor="let item of items">
      <ng-template
        [ngTemplateOutlet]="stepTemplateRef"
        [ngTemplateOutletContext]="{ $implicit: item }">
      </ng-template>
    </ng-container>
  `
})

export class UtmListComponent {

  @Input() items: any[];

  @ContentChild(StepDirective, { read: TemplateRef })
  stepTemplateRef!: TemplateRef<unknown>;
  constructor() {
  }
}
