import {Component, OnInit} from '@angular/core';
import {NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import {EventDataTypeEnum} from '../../data-management/alert-management/shared/enums/event-data-type.enum';
import {FILTER_RULE_FIELDS} from '../models/rule.constant';
import {FilterService} from "../services/filter.service";
import {SERVER_API_URL} from "../../app.constants";
import {ITEMS_PER_PAGE} from "../../shared/constants/pagination.constants";
import {ConfigService} from "../app-correlation-management/services/config.service";
import {Actions} from "../app-correlation-management/models/config.type";


@Component({
    selector: 'app-rules',
    templateUrl: './app-rule.component.html',
    styleUrls: ['./app-rule.component.css']
})
export class AppRuleComponent implements OnInit {

    constructor(private filterService: FilterService,
                private configService: ConfigService) {
    }

    dataType: EventDataTypeEnum = EventDataTypeEnum.ALERT;
    fieldFilters = FILTER_RULE_FIELDS;
    filterUrl = `${SERVER_API_URL}api/correlation-rule/search-property-values`;
    requestParam = {
        page: 0,
        size: ITEMS_PER_PAGE,
        sort: 'id,desc',
    };

    ngOnInit() {}

    addRule() {
      this.configService.onAction(Actions.CREATE_RULE);
    }

  handleResponse(modal: NgbModalRef) {
    modal.result.then((result: boolean) => {
      if (result) {

      }
    });
  }

    loadPage($event: number) {

    }

    onItemsPerPageChange($event: number) {

    }


    resetAllFilters() {
        this.filterService.resetAllFilters();
    }

    /*onFilterChange($event: { prop: string, values: any }) {
        this.requestParam = {
            ...this.requestParam,
            [$event.prop]: $event.values.length > 0 ? $event.values : null
        };
        console.log(this.requestParam);

       /!* this.assetFiltersBehavior.$assetAppliedFilter.next(this.requestParam);
        this.assetFiltersBehavior.$assetFilter.next(this.requestParam);*!/
        this.getAssets();
    }*/
}
