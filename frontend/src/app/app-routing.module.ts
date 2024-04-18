import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from './core/auth/user-route-access-service';
import {ConfirmIdentityComponent} from './shared/components/auth/confirm-identity/confirm-identity.component';
import {LoginComponent} from './shared/components/auth/login/login.component';
import {
  PasswordResetFinishComponent
} from './shared/components/auth/password-reset/finish/password-reset-finish.component';
import {TotpComponent} from './shared/components/auth/totp/totp.component';
import {
  WelcomeToUtmstackComponent
} from './shared/components/getting-started/welcome-to-utmstack/welcome-to-utmstack.component';
import {NotFoundComponent} from './shared/components/not-found/not-found.component';
import {UtmLiteVersionComponent} from './shared/components/utm-lite-version/utm-lite-version.component';
import {UtmModuleDisabledComponent} from './shared/components/utm-module-disabled/utm-module-disabled.component';
import {ADMIN_ROLE, USER_ROLE} from './shared/constants/global.constant';

const routes: Routes = [
  {path: '', redirectTo: 'login', pathMatch: 'full'},
  {
    path: 'dashboard',
    loadChildren: './dashboard/dashboard.module#UtmDashboardModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'data',
    loadChildren: './data-management/data-management.module#DataManagementModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE, USER_ROLE]}
  },
  {
    path: 'explore',
    loadChildren: './filebrowser/filebrowser.module#FileBrowserModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE, USER_ROLE]}
  },
  {
    path: 'profile',
    loadChildren: './account/account.module#UtmAccountModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE, USER_ROLE]}
  },
  // {
  //   path: 'vulnerability-scanner',
  //   loadChildren: './vulnerability-scanner/vulnerability-scanner.module#VulnerabilityScannerModule',
  //   canActivate: [UserRouteAccessService, LiteModeRouteAccessService, ActiveModuleRouteAccessService],
  //   data: {authorities: [USER_ROLE, ADMIN_ROLE], module: UtmModulesEnum.VULNERABILITIES}
  // },
  {
    path: 'data-sources',
    loadChildren: './assets-discover/assets-discover.module#AssetsDiscoverModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'management',
    loadChildren: './admin/admin.module#AdminModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE]}
  },
  {
    path: 'creator',
    loadChildren: './graphic-builder/graphic-builder.module#GraphicBuilderModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'app-management',
    loadChildren: './app-management/app-management.module#AppManagementModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'integrations',
    loadChildren: './app-module/app-module.module#AppModuleModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'variables',
    loadChildren: './automation-variables/automation-variables.module#AutomationVariablesModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'active-directory',
    loadChildren: './active-directory/active-directory.module#ActiveDirectoryModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'discover',
    loadChildren: './log-analyzer/log-analyzer.module#LogAnalyzerModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE, USER_ROLE]}
  },
  {
    path: 'compliance',
    loadChildren: './compliance/compliance.module#ComplianceModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'data-parsing',
    loadChildren: './logstash/logstash.module#LogstashModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },

  {
    path: 'incident-response',
    loadChildren: './incident-response/incident-response.module#IncidentResponseModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },

  {
    path: 'incident',
    loadChildren: './incident/incident.module#IncidentModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  // {
  //   path: 'reports',
  //   loadChildren: './report/report.module#ReportModule',
  //   canActivate: [UserRouteAccessService],
  //   data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  // },
  {
    path: 'iframe',
    loadChildren: './data-management/alert-management/alert-management.module#AlertManagementModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'getting-started',
    component: WelcomeToUtmstackComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE, USER_ROLE]}
  },
  {
    path: 'threat-intelligence',
    loadChildren: './threatwind/threatwind.module#ThreatWindModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE, USER_ROLE]}
  },
  {path: '', component: LoginComponent},
  {path: 'totp', component: TotpComponent},
  {path: 'reset/finish', component: PasswordResetFinishComponent},
  {path: 'page-not-found', component: NotFoundComponent},
  {path: 'confirm-identity/:id', component: ConfirmIdentityComponent},
  {path: 'module-disabled', component: UtmModuleDisabledComponent},
  {path: 'lite-mode', component: UtmLiteVersionComponent},
  {path: '**', redirectTo: 'page-not-found'},

];

@NgModule({
  imports: [RouterModule.forRoot(routes, {
    useHash: false,
    scrollPositionRestoration: 'top',
    onSameUrlNavigation: 'reload',
    anchorScrolling: 'enabled',
    scrollOffset: [0, 64],
  })],
  exports: [RouterModule]
})
export class AppRoutingModule {
}

