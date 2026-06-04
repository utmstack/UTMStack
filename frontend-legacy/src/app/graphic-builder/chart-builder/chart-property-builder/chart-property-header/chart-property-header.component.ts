import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-chart-property-header',
  templateUrl: './chart-property-header.component.html',
  styleUrls: ['./chart-property-header.component.scss']
})
export class ChartPropertyHeaderComponent implements OnInit {
  @Output() propertySetting = new EventEmitter<string>();
  @Input() headers: { name: string, label: string }[];
  @Input() confNotify: { icon: string, label: string, class: string };
  property = '';

  constructor() {
  }

  ngOnInit() {
    this.property = this.headers[0].name;
    this.propertySetting.emit(this.property);
  }

  viewProperty(property: string) {
    this.property = property;
    this.propertySetting.emit(this.property);
  }
}
