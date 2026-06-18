import {Component, OnDestroy, OnInit} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Subject} from 'rxjs';
import {debounceTime, distinctUntilChanged, takeUntil} from 'rxjs/operators';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {
  ModalConfirmationComponent
} from '../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {TeamUser, TeamUserPageInfo} from '../../domain/team-user.model';
import {FederationTeamService} from '../../services/federation-team.service';
import {TeamUserFormModalComponent} from '../team-user-form-modal/team-user-form-modal.component';

const DEFAULT_PAGE_SIZE = 20;

@Component({
  selector: 'app-federation-team-management-modal',
  templateUrl: './team-management-modal.component.html',
  styleUrls: ['./team-management-modal.component.scss']
})
export class TeamManagementModalComponent implements OnInit, OnDestroy {
  users: TeamUser[] = [];
  pageInfo: TeamUserPageInfo | null = null;
  loading = false;
  errorMessage: string | null = null;
  searchTerm = '';
  pendingActionId: number | null = null;

  private currentPage = 1;
  private readonly pageSize = DEFAULT_PAGE_SIZE;
  private searchSubject = new Subject<string>();
  private destroy$ = new Subject<void>();

  constructor(public activeModal: NgbActiveModal,
              private teamService: FederationTeamService,
              private modalService: NgbModal,
              private toast: UtmToastService) {}

  ngOnInit(): void {
    this.searchSubject
      .pipe(debounceTime(300), distinctUntilChanged(), takeUntil(this.destroy$))
      .subscribe(() => {
        this.currentPage = 1;
        this.load();
      });
    this.load();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  onSearchChange(value: string): void {
    this.searchTerm = value;
    this.searchSubject.next(value);
  }

  prevPage(): void {
    if (!this.pageInfo || !this.pageInfo.has_prev || this.loading) {
      return;
    }
    this.currentPage = this.pageInfo.page - 1;
    this.load();
  }

  nextPage(): void {
    if (!this.pageInfo || !this.pageInfo.has_next || this.loading) {
      return;
    }
    this.currentPage = this.pageInfo.page + 1;
    this.load();
  }

  openInvite(): void {
    const ref = this.modalService.open(TeamUserFormModalComponent, {
      centered: true,
      backdrop: 'static'
    });
    ref.componentInstance.saved.subscribe(() => {
      ref.close();
      this.toast.showSuccessBottom('Invitation sent.');
      this.currentPage = 1;
      this.load();
    });
  }

  openEdit(user: TeamUser): void {
    const ref = this.modalService.open(TeamUserFormModalComponent, {
      centered: true,
      backdrop: 'static'
    });
    ref.componentInstance.user = user;
    ref.componentInstance.saved.subscribe(() => {
      ref.close();
      this.toast.showSuccessBottom('Team member updated.');
      this.load();
    });
  }

  resendInvite(user: TeamUser): void {
    if (this.pendingActionId !== null) {
      return;
    }
    this.pendingActionId = user.id;
    this.teamService.resendInvite(user.id).subscribe({
      next: () => {
        this.pendingActionId = null;
        this.toast.showSuccessBottom('Invitation re-sent.');
      },
      error: err => {
        this.pendingActionId = null;
        this.toast.showError('Resend invite failed', this.extractError(err));
      }
    });
  }

  disableTfa(user: TeamUser): void {
    if (this.pendingActionId !== null) {
      return;
    }
    const ref = this.modalService.open(ModalConfirmationComponent, {
      backdrop: 'static',
      centered: true
    });
    ref.componentInstance.header = 'Disable 2FA';
    ref.componentInstance.message = `Disable 2FA for ${this.displayName(user)}?`;
    ref.componentInstance.confirmBtnText = 'Disable';
    ref.componentInstance.confirmBtnIcon = 'icon-shield-cross';
    ref.componentInstance.confirmBtnType = 'delete';
    ref.componentInstance.textDisplay =
      'The user will be able to sign in without a 2FA code until they re-enroll.';
    ref.componentInstance.textType = 'warning';
    ref.result.then(() => this.runDisableTfa(user), () => undefined);
  }

  toggleActivation(user: TeamUser): void {
    if (this.pendingActionId !== null) {
      return;
    }
    if (user.activated) {
      this.confirmAndDeactivate(user);
    } else {
      this.activate(user);
    }
  }

  trackById(_index: number, item: TeamUser): number {
    return item.id;
  }

  displayName(user: TeamUser): string {
    const composed = [user.first_name, user.last_name].filter(part => !!part).join(' ').trim();
    return composed || user.login;
  }

  private confirmAndDeactivate(user: TeamUser): void {
    const ref = this.modalService.open(ModalConfirmationComponent, {
      backdrop: 'static',
      centered: true
    });
    ref.componentInstance.header = 'Deactivate team member';
    ref.componentInstance.message = `Deactivate ${this.displayName(user)}?`;
    ref.componentInstance.confirmBtnText = 'Deactivate';
    ref.componentInstance.confirmBtnIcon = 'icon-cancel-circle2';
    ref.componentInstance.confirmBtnType = 'delete';
    ref.componentInstance.textDisplay = 'The user will lose access to this federation server.';
    ref.componentInstance.textType = 'warning';
    ref.result.then(() => this.runDeactivate(user), () => undefined);
  }

  private runDeactivate(user: TeamUser): void {
    this.pendingActionId = user.id;
    this.teamService.deactivate(user.id).subscribe({
      next: () => {
        this.pendingActionId = null;
        this.toast.showSuccessBottom('Team member deactivated.');
        this.load();
      },
      error: err => {
        this.pendingActionId = null;
        this.toast.showError('Deactivation failed', this.extractError(err));
      }
    });
  }

  private activate(user: TeamUser): void {
    this.pendingActionId = user.id;
    this.teamService.update(user.id, {
      email: user.email || '',
      first_name: user.first_name || '',
      last_name: user.last_name || '',
      activated: true
    }).subscribe({
      next: () => {
        this.pendingActionId = null;
        this.toast.showSuccessBottom('Team member activated.');
        this.load();
      },
      error: err => {
        this.pendingActionId = null;
        this.toast.showError('Activation failed', this.extractError(err));
      }
    });
  }

  private runDisableTfa(user: TeamUser): void {
    this.pendingActionId = user.id;
    this.teamService.disableTfa(user.id).subscribe({
      next: () => {
        this.pendingActionId = null;
        this.toast.showSuccessBottom('2FA disabled.');
        this.load();
      },
      error: err => {
        this.pendingActionId = null;
        this.toast.showError('Disable 2FA failed', this.extractError(err));
      }
    });
  }

  private load(): void {
    this.loading = true;
    this.errorMessage = null;
    this.teamService.list({
      page: this.currentPage,
      page_size: this.pageSize,
      search: this.searchTerm ? this.searchTerm.trim() : undefined
    }).subscribe({
      next: response => {
        this.loading = false;
        this.users = response.data || [];
        this.pageInfo = response.page_info;
      },
      error: err => {
        this.loading = false;
        this.users = [];
        this.pageInfo = null;
        this.errorMessage = this.extractError(err);
      }
    });
  }

  private extractError(err: {error?: {message?: string}}): string {
    if (err && err.error && err.error.message) {
      return err.error.message;
    }
    return 'Operation failed. Please try again.';
  }
}
