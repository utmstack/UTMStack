import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-ir-full-log',
  templateUrl: './ir-full-log.component.html',
  styleUrls: ['./ir-full-log.component.scss']
})
export class IrFullLogComponent implements OnInit {
  @Input() fullLog: string;

  constructor() {
  }

  ngOnInit() {
  }

}
