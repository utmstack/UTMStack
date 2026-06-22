import {Component, EventEmitter, Input, Output} from '@angular/core';
import {TeamUserPageInfo} from '../../domain/team-user.model';

@Component({
  selector: 'app-team-members-pagination',
  templateUrl: './team-members-pagination.component.html'
})
export class TeamMembersPaginationComponent {
  @Input() pageInfo: TeamUserPageInfo | null = null;
  @Input() loading = false;
  @Output() prev = new EventEmitter<void>();
  @Output() next = new EventEmitter<void>();
}
