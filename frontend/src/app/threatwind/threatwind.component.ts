import {Component, OnInit} from '@angular/core';

@Component({
  selector: 'app-threatwind',
  templateUrl: './threatwind.component.html',
  styleUrls: ['./threatwind.component.css']
})
export class ThreatwindComponent implements OnInit {

  // Threat intelligence lives on threatwinds.com now; the page opens it in a
  // new tab instead of embedding galaxy.threatwinds.com in an iframe.
  url = 'https://threatwinds.com';
  opened = false;

  constructor() { }

  ngOnInit() {
    // A top-level navigation (the route load) is a user gesture, so this is
    // not blocked as a popup in practice. The in-page button below is the
    // fallback if a browser does block it.
    const win = window.open(this.url, '_blank');
    this.opened = win !== null;
  }

}
