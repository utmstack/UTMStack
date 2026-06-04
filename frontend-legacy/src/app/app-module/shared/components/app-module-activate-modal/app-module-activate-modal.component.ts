import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {SYSTEM_MENU_ICONS_PATH} from '../../../../shared/constants/menu_icons.constants';
import {UtmOpenModuleModalService} from '../../../../shared/services/config/utm-open-module-modal.service';
import {UtmModulesService} from '../../services/utm-modules.service';
import {UtmModuleType} from '../../type/utm-module.type';

@Component({
  selector: 'app-app-module-activate-modal',
  templateUrl: './app-module-activate-modal.component.html',
  styleUrls: ['./app-module-activate-modal.component.scss']
})
export class AppModuleActivateModalComponent implements OnInit {
  modules: UtmModuleType[];
  loading = true;
  iconPath = SYSTEM_MENU_ICONS_PATH;
  active: number;

  constructor(public activeModal: NgbActiveModal,
              private router: Router,
              private spinner: NgxSpinnerService,
              private utmOpenModuleModalService: UtmOpenModuleModalService,
              private utmModulesService: UtmModulesService,
              public modalService: NgbModal) {
  }

  ngOnInit() {
    this.getModules();
  }

  getModules() {
    this.utmModulesService.getModules({page: 0, size: 100}).subscribe(response => {
      this.loading = false;
      this.modules = response.body;
    });
  }

  setUpModule(module: UtmModuleType) {
    this.activeModal.close();
    this.setModalAsView();
    this.spinner.show('loadingSpinner');
    this.router.navigate(['/modules/module-management'], {
      queryParams: {
        setUpModule: module.moduleName
      }
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

  setModalAsView() {
    this.utmOpenModuleModalService.update({id: 1, moduleModalShown: true}).subscribe(response => {
      this.activeModal.close();
    });
  }

  isActive(module: UtmModuleType) {
    return this.active === module.id;
  }
}
