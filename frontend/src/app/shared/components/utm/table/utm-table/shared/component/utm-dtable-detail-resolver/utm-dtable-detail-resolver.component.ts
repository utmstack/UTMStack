import {Component, ComponentFactoryResolver, Input, OnInit, Type, ViewChild} from '@angular/core';
import {TableDetailContentDirective} from '../../directive/table-detail-content.directive';
import {TableSkeletonInterface} from '../../type/table-skeleton';

@Component({
  selector: 'app-utm-dtable-detail-resolver',
  templateUrl: './utm-dtable-detail-resolver.component.html',
  styleUrls: ['./utm-dtable-detail-resolver.component.scss']
})
export class UtmDtableDetailResolverComponent implements OnInit {
  @Input() row;
  @Input() component: Type<any>;
  @ViewChild(TableDetailContentDirective) contentContainer: TableDetailContentDirective;

  constructor(private componentFactoryResolver: ComponentFactoryResolver) {
  }

  ngOnInit() {
    const componentFactory = this.componentFactoryResolver.resolveComponentFactory(this.component);
    const componentRef = this.contentContainer.viewContainerRef.createComponent(componentFactory);
    (componentRef.instance as TableSkeletonInterface).row = this.row;
  }
}
