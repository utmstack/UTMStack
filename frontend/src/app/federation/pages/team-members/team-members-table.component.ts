import {Component, EventEmitter, Input, Output} from '@angular/core';
import {TeamUser} from '../../domain/team-user.model';

@Component({
  selector: 'app-team-members-table',
  templateUrl: './team-members-table.component.html',
  styleUrls: ['./team-members-table.component.scss']
})
export class TeamMembersTableComponent {
  @Input() users: TeamUser[] = [];
  @Input() loading = false;
  @Input() pendingActionId: number | null = null;

  @Output() editUser = new EventEmitter<TeamUser>();
  @Output() resendInvite = new EventEmitter<TeamUser>();
  @Output() disableTfa = new EventEmitter<TeamUser>();
  @Output() toggleActivation = new EventEmitter<TeamUser>();

  trackById(_index: number, item: TeamUser): number {
    return item.id;
  }

  displayName(user: TeamUser): string {
    const composed = [user.first_name, user.last_name].filter(part => !!part).join(' ').trim();
    return composed || user.login;
  }
}
