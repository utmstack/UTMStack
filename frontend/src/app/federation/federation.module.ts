import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {InlineSVGModule} from 'ng-inline-svg';
import {NgxEchartsModule} from 'ngx-echarts';
import {NgxGaugeModule} from 'ngx-gauge';
import {FederationOverviewGridComponent} from './components/federation-overview-grid/federation-overview-grid.component';
import {FederationSidebarComponent} from './components/federation-sidebar/federation-sidebar.component';
import {
  FederationSidebarInstancesComponent
} from './components/federation-sidebar-instances/federation-sidebar-instances.component';
import {FederationUserMenuComponent} from './components/federation-user-menu/federation-user-menu.component';
import {InstanceFormModalComponent} from './components/instance-form-modal/instance-form-modal.component';
import {InstanceOverviewCardComponent} from './components/instance-overview-card/instance-overview-card.component';
import {TeamUserFormModalComponent} from './components/team-user-form-modal/team-user-form-modal.component';
import {
  TeamMembersPaginationComponent
} from './pages/team-members/team-members-pagination.component';
import {TeamMembersSearchComponent} from './pages/team-members/team-members-search.component';
import {TeamMembersTableComponent} from './pages/team-members/team-members-table.component';
import {TeamMembersPageComponent} from './pages/team-members/team-members.page.component';
import {WelcomeComponent} from './pages/welcome/welcome.component';

@NgModule({
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    NgbModule,
    InlineSVGModule,
    NgxEchartsModule,
    NgxGaugeModule
  ],
  declarations: [
    FederationSidebarComponent,
    FederationSidebarInstancesComponent,
    FederationUserMenuComponent,
    InstanceFormModalComponent,
    TeamUserFormModalComponent,
    WelcomeComponent,
    FederationOverviewGridComponent,
    InstanceOverviewCardComponent,
    TeamMembersPageComponent,
    TeamMembersTableComponent,
    TeamMembersSearchComponent,
    TeamMembersPaginationComponent
  ],
  exports: [
    FederationSidebarComponent,
    FederationUserMenuComponent,
    FederationOverviewGridComponent
  ],
  entryComponents: [
    InstanceFormModalComponent,
    TeamUserFormModalComponent
  ]
})
export class FederationModule {}
