import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-active-directory-privileges',
  templateUrl: './active-directory-privileges.component.html',
  styleUrls: ['./active-directory-privileges.component.scss']
})
export class AdPrivilegesComponent implements OnInit {
  @Input() privileges: string[];

  constructor() {
  }

  ngOnInit() {
  }

}
