import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {AccountService} from '../../../../core/auth/account.service';
import {Account} from '../../../../core/user/account.model';
import {IncidentNoteType} from '../../../../shared/types/incident/incident-note.type';
import {UtmIncidentNoteService} from '../../services/utm-incident-note.service';

@Component({
  selector: 'app-incident-notes',
  templateUrl: './incident-notes.component.html',
  styleUrls: ['./incident-notes.component.scss']
})
export class IncidentNotesComponent implements OnInit, OnDestroy {
  @Input() incidentId: number;
  account: Account;
  notes: IncidentNoteType[] = [];
  loading = true;
  noteText = '';
  interval: any;

  constructor(private accountService: AccountService, private incidentNoteService: UtmIncidentNoteService) {

  }

  ngOnInit() {
    this.getNotes();
    this.accountService.identity().then(account => {
      this.account = account;
    });
    this.interval = setInterval(() => {
      this.getNotes();
    }, 10000);
  }

  ngOnDestroy() {
    clearInterval(this.interval);
  }

  getNotes() {
    this.incidentNoteService.query({page: 0, size: 1000, sort: 'id,asc', 'incidentId.equals': this.incidentId}).subscribe(res => {
      this.notes = res.body;
      this.loading = false;
    });
  }

  sendNote() {
    if (this.noteText) {
      this.incidentNoteService.create({incidentId: this.incidentId, noteText: this.noteText}).subscribe(res => {
        this.getNotes();
        this.noteText = '';
      });
    }
  }

  keyDownFunction($event: KeyboardEvent) {
    if ($event.code === 'Enter') {
      this.sendNote();
    }
  }
}
