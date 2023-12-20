import {Component, OnInit} from '@angular/core';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {UtmRolloverService} from './shared/service/utm-rollover.service';
import {UtmRolloverType} from './shared/type/utm-rollover.type';

@Component({
  selector: 'app-rollover-config',
  templateUrl: './rollover-config.component.html',
  styleUrls: ['./rollover-config.component.scss']
})
export class RolloverConfigComponent implements OnInit {
  loading = true;
  rollover: UtmRolloverType;
  saving = false;
  days: any;

  constructor(private utmRolloverService: UtmRolloverService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.getRollover();
  }

  getRollover() {
    this.loading = true;
    this.utmRolloverService.getRollover().subscribe(response => {
      this.loading = false;
      this.rollover = response.body;
      this.days = Number(this.rollover.deleteAfter.match(/\d+/g));
    });
  }

  saveRollover() {
    this.saving = true;
    this.rollover.deleteAfter = this.days + 'd';
    this.utmRolloverService.update({
      snapshotActive: this.rollover.snapshotActive,
      deleteAfter: this.rollover.deleteAfter
    }).subscribe(response => {
      this.toastService.showSuccessBottom('Rollover configuration saved successfully');
      this.saving = false;
    }, error => {
      this.toastService.showError('Error saving rollover', 'Error occurring while try to update rollover configuration');
      this.saving = false;
    });
  }
}
