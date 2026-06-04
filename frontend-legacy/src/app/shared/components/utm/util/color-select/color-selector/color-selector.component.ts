import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {ColorType} from '../color.type';
import {COLORS} from '../colors.const';

@Component({
  selector: 'app-color-selector',
  templateUrl: './color-selector.component.html',
  styleUrls: ['./color-selector.component.scss']
})
export class ColorSelectorComponent implements OnInit {
  @Input() multiple: boolean;
  @Output() colorChange = new EventEmitter<string[]>();
  @Output() singleColorChange = new EventEmitter<string>();
  @Input() colorSelected: string[];
  @Input() color: string;
  colors: ColorType[] = COLORS;
  colorSearch: string;
  colorSelect: string;

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
    if (!this.colorSelected) {
      this.colorSelected = [];
    }
  }

  searchColor($event: any) {
    if (!$event) {
      this.colors = COLORS;
    } else {
      this.colors = COLORS.filter(value =>
        value.group.toLowerCase().includes($event.toString().toLowerCase())
      );
    }
  }

  emitColorSelect() {
    if (this.multiple) {
      this.colorChange.emit(this.colorSelected);
    } else {
      this.singleColorChange.emit(this.color);
    }
    this.activeModal.close();
  }

  selectColor(color: string) {
    if (this.multiple) {
      const index = this.colorSelected.findIndex(value => value === color);
      if (index === -1) {
        this.colorSelected.push(color);
      } else {
        this.colorSelected.splice(index, 1);
      }
    } else {
      this.color = color;
    }
  }

  deleteColor() {
    const index = this.colorSelected.findIndex(value => value === this.colorSelect);
    if (index !== -1) {
      this.colorSelected.splice(index, 1);
    }
    this.colorSelect = undefined;
  }

  activeSortingColors(color: string) {
    this.colorSelect = color;
  }

  moveColorElement(move) {
    let oldIndex = this.colorSelected.findIndex(value => value === this.colorSelect);
    let newIndex = oldIndex - move;
    while (oldIndex < 0) {
      oldIndex += this.colorSelected.length;
    }
    while (newIndex < 0) {
      newIndex += this.colorSelected.length;
    }
    if (newIndex >= this.colorSelected.length) {
      let k = newIndex - this.colorSelected.length;
      while ((k--) + 1) {
        this.colorSelected.push(undefined);
      }
    }
    this.colorSelected.splice(newIndex, 0, this.colorSelected.splice(oldIndex, 1)[0]);
    return this.colorSelected;
  }

  isLastIndex(): boolean {
    const index = this.colorSelected.findIndex(value => value === this.colorSelect);
    return index + 1 === this.colorSelected.length;
  }

}
