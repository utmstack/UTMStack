import {Component, Directive, Input, TemplateRef} from '@angular/core';
import {Step} from '../step';

@Directive({
  selector: 'ng-template[stepTemplateRef]'
})

export class StepDirective {
  constructor(public templateRef: TemplateRef<unknown>) {
  }
}

@Component({
  selector: 'app-step',
  template: `
      <p class="step-guide">
        <span class="step_number">
          <ng-content select ="[stepNumber]"></ng-content>
        </span>
        <ng-content></ng-content>
      </p>
  `
})

export class StepComponent{

  @Input() step: Step;
  constructor() {
  }
}
