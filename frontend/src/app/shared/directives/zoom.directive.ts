import {Directive, HostListener} from '@angular/core';

@Directive({
  selector: '[appZoom]'
})
export class ZoomDirective {
  private i = 1;
  private range = 0.5;

  constructor() {}

  @HostListener('mousewheel', ['$event']) onMousewheel(event) {
    if (event.ctrlKey === true) {
      event.preventDefault();
      if (event.wheelDelta > 0) {
        event.srcElement.style.setProperty('transition', 'all 200ms ease-in');
        event.srcElement.style.setProperty(
          'transform',
          `scale(${this.i + this.range})`
        );
      }
      if (event.wheelDelta < 0) {
        event.srcElement.style.setProperty('transition', 'all 200ms ease-out');
        event.srcElement.style.setProperty(
          'transform',
          `scale(${this.i - this.range})`
        );
      }
    }
  }
}
