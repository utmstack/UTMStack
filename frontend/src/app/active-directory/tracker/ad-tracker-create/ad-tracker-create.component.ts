import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {WINLOGBET_PATTERN} from '../../shared/const/active-directory-index-const';
import {ActiveDirectoryTreeType} from '../../shared/types/active-directory-tree.type';
import {AdTrackerService} from '../shared/services/ad-tracker.service';
import {AdTrackerType} from '../shared/type/ad-tracker.type';

@Component({
  selector: 'app-active-directory-tracker-create',
  templateUrl: './ad-tracker-create.component.html',
  styleUrls: ['./ad-tracker-create.component.scss']
})
export class AdTrackerCreateComponent implements OnInit {
  @Input() targetTracking: ActiveDirectoryTreeType[];
  @Output() trackingCreated = new EventEmitter<string>();
  winlogBeatPattern = WINLOGBET_PATTERN;
  trackers: AdTrackerType[] = [];
  creating = false;

  constructor(public activeModal: NgbActiveModal,
              public modalService: NgbModal,
              private adTrackerService: AdTrackerService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  createTracker() {
    this.adTrackerService.create(this.trackers).subscribe(value => {
      this.trackingCreated.emit('notification created');
      this.utmToastService.showSuccessBottom('Tracker created successfully');
      this.activeModal.close();
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating tracker ',
        'Error creating tracker, check your network');
    });
  }


  onTableChange($event: AdTrackerType[]) {
    this.trackers = $event;
  }
}
