import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-utm-toggle',
  templateUrl: './utm-toggle.component.html',
  styleUrls: ['./utm-toggle.component.scss']
})
export class UtmToggleComponent implements OnInit {
  @Input() active: boolean;
  @Input() label: string;
  @Input() customClass: string;
  @Input() emitAtStart: boolean;
  @Output() toggleChange = new EventEmitter<boolean>();

  constructor() {
  }

  ngOnInit() {
    if (!this.active) {
      this.active = false;
    }
    if (this.emitAtStart === true) {
      this.toggleChange.emit(this.active);
    }
  }

  toggle() {
    this.active = this.active ? false : true;
    this.toggleChange.emit(this.active);
  }
}
