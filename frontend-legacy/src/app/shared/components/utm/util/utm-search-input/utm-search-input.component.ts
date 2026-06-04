import {Component, ElementRef, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';

@Component({
  selector: 'app-utm-search-input',
  templateUrl: './utm-search-input.component.html',
  styleUrls: ['./utm-search-input.component.scss']
})
export class UtmSearchInputComponent implements OnInit {
  @Input() searching;
  @Input() placeholder: any;
  @Output() searchFor = new EventEmitter<string | number>();
  @Input() type: 'text' | 'password' | 'number' = 'text';
  @ViewChild('searcherInput') searcherInput: ElementRef;

  typing: boolean;
  valueSearch: string;
  private timer: any;

  constructor() {
  }

  ngOnInit() {
  }


  search($event: any) {
    this.typing = true;
    if (this.timer) {
      clearTimeout(this.timer);
    }
    this.timer = setTimeout(() => {
      this.searchFor.emit($event.target.value);
      this.typing = false;
    }, 2000);
  }
}
