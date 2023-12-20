import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ColorSelectorComponent} from './color-selector/color-selector.component';

@Component({
  selector: 'app-color-select',
  templateUrl: './color-select.component.html',
  styleUrls: ['./color-select.component.scss']
})
export class ColorSelectComponent implements OnInit {
  @Input() multiple: boolean;
  @Input() colors: string[];
  @Input() color: string;
  @Input() label: string;
  @Output() colorChange = new EventEmitter<string[]>();
  @Output() singleColor = new EventEmitter<string>();

  constructor(private modalService: NgbModal) {
  }

  ngOnInit(): void {
    if (!this.colors) {
      this.colors = [];
    }
  }

  selectColor() {
    const modal = this.modalService.open(ColorSelectorComponent, {centered: true});
    modal.componentInstance.multiple = this.multiple;
    modal.componentInstance.colorSelected = this.colors;
    modal.componentInstance.color = this.color;
    modal.componentInstance.colorChange.subscribe(colors => {
      this.colors = colors;
      this.color = colors;
      this.colorChange.emit(colors);
    });
    modal.componentInstance.singleColorChange.subscribe(color => {
      this.color = color;
      this.singleColor.emit(color);
    });
  }
}
