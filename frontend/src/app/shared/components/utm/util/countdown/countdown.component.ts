import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-countdown',
  templateUrl: './countdown.component.html',
  styleUrls: ['./countdown.component.css']
})
export class CountdownComponent implements OnInit, OnDestroy {
  @Input() minute = 4;
  @Input() sec = 59;
  interval;
  @Output() intervalEnd = new EventEmitter<boolean>();

  constructor() {
  }

  ngOnInit() {
    this.count();
  }

  ngOnDestroy() {
    clearInterval(this.interval);
  }

  count() {
    this.interval = setInterval(() => {
      this.sec--;
      if (this.sec === 0) {
        this.minute--;
        this.sec = 60;
        if (this.minute === 0) {
          clearInterval(this.interval);
          this.intervalEnd.emit(true);
        }
      }
    }, 1000);
  }
}
