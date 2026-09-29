import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {
  TeamUser,
  TeamUserCreatePayload,
  TeamUserUpdatePayload
} from '../../domain/team-user.model';
import {FederationTeamService} from '../../services/federation-team.service';

interface TeamUserFormValue {
  login: string;
  email: string;
  first_name: string;
  last_name: string;
}

@Component({
  selector: 'app-federation-team-user-form-modal',
  templateUrl: './team-user-form-modal.component.html',
  styleUrls: ['./team-user-form-modal.component.scss']
})
export class TeamUserFormModalComponent implements OnInit {
  @Input() user: TeamUser | null = null;
  @Output() saved = new EventEmitter<TeamUser>();

  submitting = false;
  errorMessage: string | null = null;
  model: TeamUserFormValue = {
    login: '',
    email: '',
    first_name: '',
    last_name: ''
  };

  constructor(public activeModal: NgbActiveModal,
              private teamService: FederationTeamService) {}

  ngOnInit(): void {
    if (this.user) {
      this.model = {
        login: this.user.login,
        email: this.user.email || '',
        first_name: this.user.first_name || '',
        last_name: this.user.last_name || ''
      };
    }
  }

  get isEditMode(): boolean {
    return this.user !== null;
  }

  get title(): string {
    return this.isEditMode ? 'Edit team member' : 'Invite team member';
  }

  submit(): void {
    if (this.submitting) {
      return;
    }
    this.submitting = true;
    this.errorMessage = null;

    const action$ = this.isEditMode && this.user
      ? this.teamService.update(this.user.id, this.buildUpdatePayload())
      : this.teamService.create(this.buildCreatePayload());

    action$.subscribe({
      next: result => {
        this.submitting = false;
        this.saved.emit(result);
      },
      error: err => {
        this.submitting = false;
        this.errorMessage = this.extractError(err);
      }
    });
  }

  cancel(): void {
    this.activeModal.dismiss();
  }

  private buildCreatePayload(): TeamUserCreatePayload {
    return {
      login: this.model.login.trim(),
      email: this.model.email.trim(),
      first_name: this.model.first_name.trim(),
      last_name: this.model.last_name.trim(),
      lang_key: 'en'
    };
  }

  private buildUpdatePayload(): TeamUserUpdatePayload {
    return {
      email: this.model.email.trim(),
      first_name: this.model.first_name.trim(),
      last_name: this.model.last_name.trim()
    };
  }

  private extractError(err: {status?: number; error?: {message?: string}}): string {
    if (err && err.status === 409) {
      return 'Login or email is already in use.';
    }
    if (err && err.error && err.error.message) {
      return err.error.message;
    }
    return 'Failed to save team member.';
  }
}
