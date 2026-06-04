import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {replaceBreakLine} from '../../../../util/string-util';

@Component({
  selector: 'app-utm-terminal-input',
  templateUrl: './utm-terminal-input.component.html',
  styleUrls: ['./utm-terminal-input.component.scss']
})
export class UtmTerminalInputComponent implements OnInit {
  /**
   * Text to view before line command
   */
  @Input() terminal: any;
  @Input() param: any;
  /**
   * Event to output
   */
  @Output() paramChange = new EventEmitter<string>();

  @Input() readonly = false;

  constructor() {
  }

  ngOnInit() {
    this.param = replaceBreakLine(this.param);
  }

}
