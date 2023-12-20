import {Component, ComponentFactoryResolver, Input, OnInit, ViewChild} from '@angular/core';
import {TabContentDirective} from '../../directive/tab-content.directive';
import {SkeletonInterface} from '../../type/skeleton.interface';
import {TabType} from '../../type/tab.type';

@Component({
  selector: 'app-tab-content',
  templateUrl: './tab-content.component.html',
  styleUrls: ['./tab-content.component.scss']
})
export class TabContentComponent implements OnInit {
  @Input() tab;
  @ViewChild(TabContentDirective)
  contentContainer: TabContentDirective;

  constructor(private componentFactoryResolver: ComponentFactoryResolver) {
  }

  ngOnInit() {
    const tab: TabType = this.tab;
    const componentFactory = this.componentFactoryResolver.resolveComponentFactory(
      tab.component
    );
    const componentRef = this.contentContainer.viewContainerRef.createComponent(componentFactory);
    (componentRef.instance as SkeletonInterface).data = tab.tabData;
    (componentRef.instance as SkeletonInterface).uuid = tab.uuid;
  }
}
