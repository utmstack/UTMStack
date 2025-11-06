import {Component, EventEmitter, Input, OnChanges, OnInit, Output, SimpleChanges} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {AlertTags} from '../../../../../../shared/types/alert/alert-tag.type';
import {UtmAlertType} from '../../../../../../shared/types/alert/utm-alert.type';
import {AlertUpdateTagBehavior} from '../../../behavior/alert-update-tag.behavior';
import {AlertRuleCreateComponent} from '../../alert-rule-create/alert-rule-create.component';
import { FALSE_POSITIVE_OBJECT } from 'src/app/shared/constants/alert/alert-field.constant';

@Component({
  selector: 'app-alert-tags-apply',
  templateUrl: './alert-tags-apply.component.html',
  styleUrls: ['./alert-tags-apply.component.scss'],
})
export class AlertTagsApplyComponent implements OnInit, OnChanges {
  @Input() showTagsLabel: boolean;
  @Input() alert: UtmAlertType;
  @Input() tags: AlertTags[];
  @Input() template: 'default' | 'menu-item' = 'default';
  @Input() action: any;
  selected: string[] = [];
  select: any;
  @Output() updateTagsEvent = new EventEmitter<boolean>();
  @Output() applyTagsEvent = new EventEmitter<{ tags: string[], automatic: boolean }>();
  icon: string;
  color: string;

  constructor(private modalService: NgbModal,
              private alertUpdateTagBehavior: AlertUpdateTagBehavior) {
  }

  ngOnInit() {
    this.selected = [];
    if (this.alert) {
      const alertTags = this.alert.tags;
      this.selected = alertTags ? alertTags : [];
      this.icon = this.getTagIcon();
      this.color = this.getColor();
    }
    this.alertUpdateTagBehavior.$updateTagForAlert.subscribe(id => {
      const alertid = this.alert.id;
      if (id && id === alertid) {
        this.icon = this.getTagIcon();
        this.color = this.getColor();
      }
    });
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes.alert || changes.tags) {
      this.icon = this.getTagIcon();
      this.color = this.getColor();
    }
  }

  addNewTagRule(isFalsePositive: boolean = false) {
    const modalRef = this.modalService.open(AlertRuleCreateComponent, {
      centered: true,
      size: 'lg',
      windowClass: 'alert-rule-modal'
    });
    modalRef.componentInstance.alert = this.alert;
    modalRef.componentInstance.action = 'select';
    modalRef.componentInstance.isFalsePositiveRule = isFalsePositive;
    modalRef.componentInstance.ruleAdd.subscribe((created) => {
      this.icon = this.getTagIcon();
      this.color = this.getColor();
      this.applyTagsEvent.emit({tags: [], automatic: true});
    });
  }


  getTagIcon() {
    if (this.selected.length === 0) {
      return 'icon-price-tag3';
    } else if (this.selected.length === 1) {
      return 'icon-price-tag2';
    } else {
      return 'icon-price-tags2';
    }
  }

  getColor() {
    if (this.selected.length === 0) {
      return '#0277bd';
    } else if (this.selected.length === 1 && this.tags) {
      const tag = this.selected[0];
      const index = this.tags.findIndex(value => value.tagName === tag);
      if (index !== -1) {
        const color = this.tags[index].tagColor;
        if (color) {
          return color;
        } else {
          return '#0277bd';
        }
      } else {
        return '#0277bd';
      }
    } else {
      return '#0277bd';
    }
  }

}
