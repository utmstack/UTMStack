import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {AdTrackerService} from '../shared/services/ad-tracker.service';
import {AdTrackerType} from '../shared/type/ad-tracker.type';

@Component({
  selector: 'app-ad-tracker-delete',
  templateUrl: './ad-tracker-delete.component.html',
  styleUrls: ['./ad-tracker-delete.component.scss']
})
export class AdTrackerDeleteComponent implements OnInit {
  @Input() tracker: AdTrackerType;
  @Output() trackerDeleted = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private adTrackerService: AdTrackerService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  deleteTracker() {
    this.adTrackerService.delete(this.tracker.id).subscribe(() => {
      this.trackerDeleted.emit('tracker deleted');
      this.utmToastService.showSuccessBottom('Tracker deleted successfully');
      this.activeModal.close();
    }, error1 => {
      this.utmToastService.showError('Error deleting tracker',
        'Error deleting tracker, check your network');
    });
  }
}
