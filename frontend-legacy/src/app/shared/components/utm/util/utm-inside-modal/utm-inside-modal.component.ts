import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-utm-inside-modal',
  templateUrl: './utm-inside-modal.component.html',
  styleUrls: ['./utm-inside-modal.component.scss']
})
export class UtmInsideModalComponent implements OnInit {
  @Input() viewModalCondition: boolean;

  constructor() {
  }

  ngOnInit() {
  }

}
