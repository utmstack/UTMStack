import {Component, ContentChild, Input, TemplateRef} from '@angular/core';
import {Step} from '../step';
import {StepDirective} from './step.component';

@Component({
  selector: 'app-step-list',
  template: `
    <li *ngFor="let step of steps">
      <ng-template
        [ngTemplateOutlet]="stepTemplateRef"
        [ngTemplateOutletContext]="{ $implicit: step }">
      </ng-template>
    </li>
  `
})

export class StepListComponent {

  @Input() steps: Step[];

  @ContentChild(StepDirective, { read: TemplateRef })
  stepTemplateRef!: TemplateRef<unknown>;
  constructor() {
  }
}
