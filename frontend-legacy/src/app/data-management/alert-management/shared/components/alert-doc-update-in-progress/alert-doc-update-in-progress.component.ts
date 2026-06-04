import {Component, OnInit} from '@angular/core';
import {AlertUpdateSolutionBehavior} from '../../behavior/alert-update-solution.behavior';

@Component({
  selector: 'app-alert-doc-update-in-progress',
  templateUrl: './alert-doc-update-in-progress.component.html',
  styleUrls: ['./alert-doc-update-in-progress.component.scss']
})
export class AlertDocUpdateInProgressComponent implements OnInit {
  updateDocSolStatus: 'init' | 'error' | 'done';
  showUp: boolean;

  constructor(private alertUpdateSolutionBehavior: AlertUpdateSolutionBehavior) {
  }

  ngOnInit() {
    this.alertUpdateSolutionBehavior.$updateSolution.subscribe(status => {
      if (status) {
        this.updateDocSolStatus = status;
        if (this.updateDocSolStatus && this.updateDocSolStatus !== 'done') {
          this.showUp = true;
        } else if (this.updateDocSolStatus === 'done') {
          setTimeout(() => {
            this.showUp = false;
          }, 2100);
        }
      }
    });
  }

  isNotDone() {
    setTimeout(() => {
      return this.updateDocSolStatus && this.updateDocSolStatus !== 'done';
    }, 2000);
  }
}
