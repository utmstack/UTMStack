import {Component, Input, OnInit} from '@angular/core';
import {ActiveDirectoryUserComputer} from '../../../../../../types/active-directory-user-computer';
import {ActiveDirectoryType} from '../../../../../../types/active-directory.type';

@Component({
  selector: 'app-ad-user-computer',
  templateUrl: './ad-user-computer.component.html',
  styleUrls: ['./ad-user-computer.component.scss']
})
export class AdUserComputerComponent implements OnInit {
  @Input() adInfo: ActiveDirectoryType;
  computerDetail: ActiveDirectoryUserComputer;
  computers: ActiveDirectoryUserComputer[] = [];
  search: string;

  constructor() {
  }

  ngOnInit() {
    this.computers = this.adInfo.computerInformation;
  }

  viewComputerDetail(computer: ActiveDirectoryUserComputer) {
    this.computerDetail = computer;
  }

  onSearchComputer($event: string) {
    this.search = $event;
    if ($event) {
      this.computers = this.computers.filter(comp => comp.name.toLocaleLowerCase().includes($event.toLocaleLowerCase()) ||
        comp.objectSid.includes($event));
    } else {
      this.computers = this.adInfo.computerInformation;
    }
  }
}
