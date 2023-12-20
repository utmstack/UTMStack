import {Component, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges} from '@angular/core';
import {UtmToastService} from '../../../../../../shared/alert/utm-toast.service';
import {ALERT_NOTE_FIELD} from '../../../../../../shared/constants/alert/alert-field.constant';
import {getValueFromPropertyPath} from '../../../../../../shared/util/get-value-object-from-property-path.util';
import {AlertUpdateHistoryBehavior} from '../../../behavior/alert-update-history.behavior';
import {AlertManagementService} from '../../../services/alert-management.service';
import {getID, setAlertPropertyValue} from '../../../util/alert-util-function';

@Component({
  selector: 'app-alert-apply-note',
  templateUrl: './alert-apply-note.component.html',
  styleUrls: ['./alert-apply-note.component.scss']
})
export class AlertApplyNoteComponent implements OnInit, OnChanges {
  @Input() alert: any;
  @Input() showNote: boolean;
  @Output() applyNote = new EventEmitter<string>();
  note: string;
  creating = false;

  constructor(private alertServiceManagement: AlertManagementService,
              private utmToastService: UtmToastService,
              private alertUpdateHistoryBehavior: AlertUpdateHistoryBehavior) {
  }

  ngOnInit() {
    this.getNoteValue();
  }

  ngOnChanges(changes: SimpleChanges) {
    this.getNoteValue();
  }

  getNoteValue() {
    if (this.alert) {
      const note = getValueFromPropertyPath(this.alert, ALERT_NOTE_FIELD, null);
      this.note = note ? note : '';
    }
  }

  addNote() {
    this.creating = true;
    this.alertServiceManagement.updateNotes(getID(this.alert), this.note).subscribe(response => {
      this.utmToastService.showSuccessBottom('Comment added successfully');
      this.applyNote.emit('success');
      this.creating = false;
      this.alert = setAlertPropertyValue(ALERT_NOTE_FIELD, this.note, this.alert);
      this.alertUpdateHistoryBehavior.$refreshHistory.next(true);
    }, error => {
      this.utmToastService.showError('Error adding note',
        'Error adding comment, please try again');
      this.creating = false;
    });
  }
}
