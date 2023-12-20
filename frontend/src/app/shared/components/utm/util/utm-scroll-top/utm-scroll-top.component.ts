import {Component, HostListener, OnInit} from '@angular/core';

@Component({
  selector: 'app-utm-scroll-top',
  templateUrl: './utm-scroll-top.component.html',
  styleUrls: ['./utm-scroll-top.component.css']
})
export class UtmScrollTopComponent implements OnInit {
  isScrolled = false;

  constructor() {
  }

  ngOnInit() {
  }

  @HostListener('window:scroll', ['$event']) // <- Add scroll listener to window
  scrolled(event: any): void {
    this.isScrolled = true;
  }

  scrollToTop() {
    this.isScrolled = false;
    window.scrollTo(0, 0);
  }
}
