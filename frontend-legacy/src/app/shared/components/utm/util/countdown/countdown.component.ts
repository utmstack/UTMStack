import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-countdown',
  templateUrl: './countdown.component.html',
  styleUrls: ['./countdown.component.css']
})
export class CountdownComponent implements OnInit, OnDestroy {
  @Input() duration = 300;
  @Output() intervalEnd = new EventEmitter<void>();

  remainingSeconds: number;
  intervalId: any;

  get minutes(): number {
    return Math.floor(this.remainingSeconds / 60);
  }

  get seconds(): number {
    return this.remainingSeconds % 60;
  }

  ngOnInit() {
    this.remainingSeconds = this.duration;
    this.startCountdown();
  }

  startCountdown() {
    this.intervalId = setInterval(() => {
      if (this.remainingSeconds > 0) {
        this.remainingSeconds--;
      } else {
        clearInterval(this.intervalId);
        this.intervalEnd.emit();
      }
    }, 1000);
  }

  ngOnDestroy() {
    clearInterval(this.intervalId);
  }
}
