import {Component, Input, OnInit} from '@angular/core';
import {TargetModel} from '../../../shared/model/target.model';

@Component({
  selector: 'app-target-detail',
  templateUrl: './target-detail.component.html',
  styleUrls: ['./target-detail.component.scss']
})
export class TargetDetailComponent implements OnInit {
  @Input() target: TargetModel;

  constructor() {
  }

  ngOnInit() {
  }

  resolveSshCredential() {
    let credential = '';
    if (this.target.sshCredential.name !== null) {
      credential += this.target.sshCredential.name;
      if (this.target.sshCredential.port !== null) {
        credential += ' on port ' + this.target.sshCredential.port;
      }
    } else {
      credential = '';
    }
    return credential;
  }


}
