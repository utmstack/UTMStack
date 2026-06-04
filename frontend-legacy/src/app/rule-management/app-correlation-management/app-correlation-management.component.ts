import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, NavigationEnd, Router} from '@angular/router';
import { filter } from 'rxjs/operators';
import {Actions, RouteValues} from './models/config.type';
import {ConfigService} from './services/config.service';

@Component({
  selector: 'app-correlation-management',
  templateUrl: './app-correlation-management.component.html',
  styleUrls: ['./app-correlation-management.component.scss']
})
export class AppCorrelationManagementComponent implements OnInit {

  routes = RouteValues;
  action: Actions;

  constructor(private router: Router,
              private configService: ConfigService) {
  }

  ngOnInit() {

    this.updateActionBasedOnUrl(this.router.url);

    this.router.events
        .pipe(filter(event => event instanceof NavigationEnd))
        .subscribe(() => {
          this.updateActionBasedOnUrl(this.router.url);
        });
  }

  add() {
    this.configService.onAction(this.action);
  }

  updateActionBasedOnUrl(url: string): void {
    const urlSegments = url.split('/');
    const lastFragment = urlSegments[urlSegments.length - 1];

    switch (lastFragment) {
      case this.routes.types: {
        this.action = Actions.CREATE_TYPE;
        break;
      }
      case this.routes.patterns: {
        this.action = Actions.CREATE_PATTERN;
        break;
      }
      default:
        this.action = Actions.CREATE_ASSET;
    }
  }


}
