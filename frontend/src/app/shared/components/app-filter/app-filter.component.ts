import {Component, ContentChild, Input, OnInit, TemplateRef} from '@angular/core';
import {TemplateSelectorDirective} from '../../directives/template-selector/template-selector.directive';

@Component({
    standalone: false,
    selector: 'app-filter',
    templateUrl: './app-filter.component.html',
    styleUrls: ['./app-filter.component.scss'],
})
export class AppFilterComponent<T extends {id: number}> {
  @Input() list: T[] | null = null;
  @ContentChild(TemplateSelectorDirective, {static: false, read: TemplateRef})
  templateRef: TemplateRef<{$implicit: T, index: number}>;

  constructor() { }

  trackByFn(index: number, item: T) {
    return item.id;
  }
}
