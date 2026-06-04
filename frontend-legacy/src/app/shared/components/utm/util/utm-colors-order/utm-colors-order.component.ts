import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-utm-colors-order',
  templateUrl: './utm-colors-order.component.html',
  styleUrls: ['./utm-colors-order.component.scss']
})
export class UtmColorsOrderComponent implements OnInit {
  @Input() colors: string[];
  @Output() orderColorChange = new EventEmitter<string[]>();
  colorSelect: string;

  constructor() {
  }

  ngOnInit() {
  }

  moveColorElement(move) {
    let oldIndex = this.colors.findIndex(value => value === this.colorSelect);
    let newIndex = oldIndex - move;
    while (oldIndex < 0) {
      oldIndex += this.colors.length;
    }
    while (newIndex < 0) {
      newIndex += this.colors.length;
    }
    if (newIndex >= this.colors.length) {
      let k = newIndex - this.colors.length;
      while ((k--) + 1) {
        this.colors.push(undefined);
      }
    }
    this.colors.splice(newIndex, 0, this.colors.splice(oldIndex, 1)[0]);
    return this.colors;
  }

  isLastIndex(): boolean {
    const index = this.colors.findIndex(value => value === this.colorSelect);
    return index + 1 === this.colors.length;
  }

  activeSortingColors(color: string) {
    this.colorSelect = color;
  }

  emitColorSelect() {
  }
}
