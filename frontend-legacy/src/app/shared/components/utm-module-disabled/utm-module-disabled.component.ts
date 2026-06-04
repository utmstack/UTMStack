import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {UtmModuleEnum} from '../../enums/utm-module.enum';

@Component({
  selector: 'app-utm-module-disabled',
  templateUrl: './utm-module-disabled.component.html',
  styleUrls: ['./utm-module-disabled.component.scss']
})
export class UtmModuleDisabledComponent implements OnInit {
  moduleDisable = '';

  constructor(private activatedRoute: ActivatedRoute) {
    activatedRoute.queryParams.subscribe(params => {
      this.moduleDisable = UtmModuleEnum[params.module];
    });
  }

  ngOnInit() {
  }

}
