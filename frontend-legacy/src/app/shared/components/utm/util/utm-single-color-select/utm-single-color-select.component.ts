import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {ColorType} from '../color-select/color.type';
import {COLORS} from '../color-select/colors.const';

@Component({
  selector: 'app-utm-single-color-select',
  templateUrl: './utm-single-color-select.component.html',
  styleUrls: ['./utm-single-color-select.component.scss']
})
export class UtmSingleColorSelectComponent implements OnInit {
  @Output() singleColorChange = new EventEmitter<string>();
  @Input() color: string;
  @Input() label: string;
  colors: ColorType[] = COLORS;
  colorSelect: string;

  constructor() {
  }

  ngOnInit() {
  }

  searchColor($event: string) {
    if (!$event) {
      this.colors = COLORS;
    } else {
      this.colors = COLORS.filter(value =>
        value.group.toLowerCase().includes($event.toLowerCase())
      );
    }
  }

  selectColor(color: string) {
    this.colorSelect = color;
  }
}
