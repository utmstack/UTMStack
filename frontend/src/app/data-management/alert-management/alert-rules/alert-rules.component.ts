import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {FormBuilder} from '@angular/forms';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {SourcesService} from '../../../admin/sources/sources.service';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {AlertTags} from '../../../shared/types/alert/alert-tag.type';
import {TimeFilterType} from '../../../shared/types/time-filter.type';
import {AlertRulesService} from '../shared/services/alert-rules.service';
import {AlertTagService} from '../shared/services/alert-tag.service';
import {AlertRuleType} from './alert-rule.type';

@Component({
  selector: 'app-rules',
  templateUrl: './alert-rules.component.html',
  styleUrls: ['./alert-rules.component.scss']
})
export class AlertRulesComponent implements OnInit {
  loading = true;
  rules: AlertRuleType[];
  range: TimeFilterType;
  totalItems: number;
  itemsPerPage = ITEMS_PER_PAGE;
  request = {
    page: 0,
    size: ITEMS_PER_PAGE,
    sort: '',
    name: null,
    conditionField: null,
    conditionValue: null,
    tagIds: null
  };
  viewRule: AlertRuleType;
  tags: AlertTags[];


  constructor(private sourcesService: SourcesService,
              private fb: FormBuilder,
              private ruleService: AlertRulesService,
              private modalService: NgbModal,
              private utmToastService: UtmToastService,
              private alertTagService: AlertTagService) {
  }

  ngOnInit() {
    this.getRules();
    this.getTags();
  }

  getTags() {
    const request = {
      page: 0,
      size: 500
    };
    this.alertTagService.query(request).subscribe(cat => {
      this.tags = cat.body;
      this.loading = false;
    });
  }

  deleteRule(rule) {
    if (rule) {
      this.ruleService.delete(rule.id).subscribe(() => {
        this.utmToastService.showSuccessBottom('Rule ' + rule.name + ' deleted successfully');
        this.getRules();
      });
    }
  }

  deleteRuleAction(rule: AlertRuleType) {
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {backdrop: 'static', centered: true});
    deleteModalRef.componentInstance.header = 'Delete rule';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete the rule: \n' + rule.name;
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-stack-text';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.componentInstance.textDisplay = 'If you delete this rule, future alerts' +
      ' that meet its conditions will stop being tagged.';
    deleteModalRef.componentInstance.textType = 'warning';
    deleteModalRef.result.then(() => {
      this.deleteRule(rule);
    });
  }

  loadPage(page: number) {
    this.request.page = page - 1;
    this.getRules();
  }

  onItemsPerPageChange($event: number) {
    this.request.size = $event;
    this.getRules();
  }

  getRules() {
    this.ruleService.query(this.request).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.rules = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  onSort($event: SortEvent) {
    this.request.sort = $event.column + ',' + $event.direction;
    this.getRules();
  }


  searchByRule($event: string) {
    this.request.name = $event;
    this.getRules();
  }

  filterByTags($event: any) {
    this.request.tagIds = $event.length === 0 ? null : $event.map(value => value.id);
    this.getRules();
  }
}
