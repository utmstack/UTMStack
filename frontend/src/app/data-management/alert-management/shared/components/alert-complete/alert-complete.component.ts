import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {FALSE_POSITIVE_OBJECT} from '../../../../../shared/constants/alert/alert-field.constant';
import {CLOSED, IGNORED} from '../../../../../shared/constants/alert/alert-status.constant';
import {AlertTags} from '../../../../../shared/types/alert/alert-tag.type';
import {AlertRuleType} from '../../../alert-rules/alert-rule.type';
import {AlertUpdateTagBehavior} from '../../behavior/alert-update-tag.behavior';
import {AlertManagementService} from '../../services/alert-management.service';
import {AlertRulesService} from '../../services/alert-rules.service';
import {AlertRuleCreateComponent} from '../alert-rule-create/alert-rule-create.component';

@Component({
  selector: 'app-alert-complete',
  templateUrl: './alert-complete.component.html',
  styleUrls: ['./alert-complete.component.scss']
})
export class AlertCompleteComponent implements OnInit {
  @Input() alertsIDs: string[] = [];
  @Input() status: number;
  @Input() alert: any;
  @Input() canCreateRule = true;
  CLOSED = CLOSED;
  IGNORED = IGNORED;
  @Output() statusChange: EventEmitter<string> = new EventEmitter();
  @Output() statusClose: EventEmitter<string> = new EventEmitter();
  observations: string;
  creating: boolean;
  rule: AlertRuleType;

  constructor(public activeModal: NgbActiveModal,
              public modalService: NgbModal,
              private alertRulesService: AlertRulesService,
              private utmToastService: UtmToastService,
              private alertUpdateTagBehavior: AlertUpdateTagBehavior,
              private alertServiceManagement: AlertManagementService) {
  }

  ngOnInit() {
    this.observations = '';
  }

  changeStatus() {
    this.creating = true;
    this.alertServiceManagement.updateAlertStatus(this.alertsIDs,
      this.status, this.observations).subscribe(al => {
      if (this.canCreateRule && this.rule) {
        this.alertRulesService.create(this.rule).subscribe(value => {
          this.utmToastService.showSuccessBottom('Rule ' + this.rule.name + ' created successfully');
          this.alertServiceManagement.updateAlertTags(this.alertsIDs, this.rule.tags.map(value1 => value1.tagName), true)
            .subscribe(tagsResponse => {
              this.alertUpdateTagBehavior.$tagRefresh.next(true);
              this.utmToastService.showSuccessBottom('Tags added successfully');
            });
        });
      }
      this.statusChange.emit('success');
      this.activeModal.close();
      this.creating = false;
    });
  }

  createRule() {
    const modal = this.modalService.open(AlertRuleCreateComponent, {centered: true, size: 'lg'});
    const falsePositive: AlertTags[] = [FALSE_POSITIVE_OBJECT];
    if (this.rule) {
      modal.componentInstance.rule = this.rule;
    }
    modal.componentInstance.alert = this.alert;
    modal.componentInstance.isForComplete = true;
    modal.componentInstance.tags = falsePositive;
    modal.componentInstance.ruleAdd.subscribe(rule => {
      this.rule = rule;
    });
  }
}
