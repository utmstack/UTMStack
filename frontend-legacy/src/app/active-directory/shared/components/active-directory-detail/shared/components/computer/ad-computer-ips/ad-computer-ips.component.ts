import {Component, Input, OnInit} from '@angular/core';
import {ActiveDirectoryType} from '../../../../../../types/active-directory.type';

@Component({
  selector: 'app-ad-computer-ips',
  templateUrl: './ad-computer-ips.component.html',
  styleUrls: ['./ad-computer-ips.component.scss']
})
export class AdComputerIpsComponent implements OnInit {
  @Input() adInfo: ActiveDirectoryType;

  constructor() {
  }

  ngOnInit() {
  }

}
