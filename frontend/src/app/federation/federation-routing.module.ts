import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {WelcomeComponent} from './pages/welcome/welcome.component';

const routes: Routes = [
  {
    path: 'federation',
    children: [
      {path: '', redirectTo: 'welcome', pathMatch: 'full'},
      {path: 'welcome', component: WelcomeComponent},
      {path: 'instances', component: WelcomeComponent}
    ]
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class FederationRoutingModule {}
