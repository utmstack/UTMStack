import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-threatwind',
  templateUrl: './threatwind.component.html',
  styleUrls: ['./threatwind.component.css']
})
export class ThreatwindComponent implements OnInit {

  height = window.innerHeight + 250;
  iframeSrc = 'https://galaxy.threatwinds.com/';

  constructor() { }

  ngOnInit() {
  }

}
