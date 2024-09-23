import {Component, OnInit} from '@angular/core';
import {NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import {SERVER_API_URL} from '../../app.constants';
import {EventDataTypeEnum} from '../../data-management/alert-management/shared/enums/event-data-type.enum';
import {ITEMS_PER_PAGE} from '../../shared/constants/pagination.constants';
import {Actions} from '../app-correlation-management/models/config.type';
import {ConfigService} from '../app-correlation-management/services/config.service';
import {FILTER_RULE_FIELDS} from '../models/rule.constant';
import {FilterService} from '../services/filter.service';


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

    resetAllFilters() {
      this.filterService.resetAllFilters();
    }
}
