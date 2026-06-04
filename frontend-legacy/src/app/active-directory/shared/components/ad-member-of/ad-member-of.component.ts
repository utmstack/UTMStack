import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-ad-member-of',
  templateUrl: './ad-member-of.component.html',
  styleUrls: ['./ad-member-of.component.scss']
})
export class AdMemberOfComponent implements OnInit {
  @Input() members: string[];

  constructor() {
  }

  ngOnInit() {
  }

}
