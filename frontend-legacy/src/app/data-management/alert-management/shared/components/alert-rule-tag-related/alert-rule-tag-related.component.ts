import {Component, Input, OnInit} from '@angular/core';
import {AlertRuleType} from '../../../alert-rules/alert-rule.type';
import {AlertRulesService} from '../../services/alert-rules.service';

@Component({
  selector: 'app-alert-rule-tag-related',
  templateUrl: './alert-rule-tag-related.component.html',
  styleUrls: ['./alert-rule-tag-related.component.scss']
})
export class AlertRuleTagRelatedComponent implements OnInit {
  @Input() tagsId: number[];
  relatedRules: AlertRuleType[];
  loading = true;
  viewDetail: AlertRuleType;

  constructor(private alertRulesService: AlertRulesService) {
  }

  ngOnInit() {
    if (this.tagsId && this.tagsId.length > 0 && !this.relatedRules) {
      this.getRelatedRules();
    }
  }

  getRelatedRules() {
    this.loading = true;
    this.alertRulesService.getByIds(this.tagsId).subscribe(response => {
      this.relatedRules = response.body;
      this.loading = false;
    });
  }

}
