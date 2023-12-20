import {Routes} from '@angular/router';
import {activateRoute} from './activate/activate.route';
import {registerRoute} from './register/register.route';


const ACCOUNT_ROUTES = [
  activateRoute,
  registerRoute];

export const accountState: Routes = [
  {
    path: '',
    children: ACCOUNT_ROUTES
  }
];
