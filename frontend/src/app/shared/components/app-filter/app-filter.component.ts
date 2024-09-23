import {Component, ContentChild, Input, OnInit, TemplateRef} from '@angular/core';
import {TemplateSelectorDirective} from '../../directives/template-selector/template-selector.directive';

@Component({
  selector: 'app-filter',
  templateUrl: './app-filter.component.html',
  styleUrls: ['./app-filter.component.css']
})
export class AppFilterComponent<T extends {id: number}> {
  @Input() list: T[] | null = null;
  @ContentChild(TemplateSelectorDirective, {read: TemplateRef})
  templateRef: TemplateRef<{$implicit: T, index: number}>;

  constructor() { }

  trackByFn(index: number, item: T) {
    return item.id;
  }
}
