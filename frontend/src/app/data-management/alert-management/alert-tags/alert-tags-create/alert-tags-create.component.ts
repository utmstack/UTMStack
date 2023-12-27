import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {AlertTags} from '../../../../shared/types/alert/alert-tag.type';
import {AlertTagService} from '../../shared/services/alert-tag.service';

@Component({
  selector: 'app-alert-tags-create',
  templateUrl: './alert-tags-create.component.html',
  styleUrls: ['./alert-tags-create.component.scss']
})
export class AlertTagsCreateComponent implements OnInit {
  name: string;
  tagColor = '#0277bd';
  @Output() addTag: EventEmitter<AlertTags> = new EventEmitter();
  @Input() tag: AlertTags;
  edit: boolean;
  typing: boolean;
  exist = false;
  private timer: any;

  constructor(public activeModal: NgbActiveModal,
              private tagService: AlertTagService) {
  }

  ngOnInit() {
    if (this.tag !== undefined) {
      this.edit = true;
      this.name = this.tag.tagName;
    } else {
      this.edit = false;
      this.name = '';
    }
  }

  createTag() {
    this.tagService.create({
      tagName: this.name,
      tagColor: this.tagColor
    }).subscribe(response => {
      this.addTag.emit(response.body);
      this.activeModal.close();
    });
  }

  editTag() {
    this.tagService.update({
      tagName: this.name,
      tagColor: this.tagColor,
      id: this.tag.id
    }).subscribe(response => {
      this.activeModal.close();
      this.addTag.emit(response.body);
    });
  }

  checkName() {
    this.typing = true;
    this.exist = true;
    if (this.timer) {
      clearTimeout(this.timer);
    }
    this.timer = setTimeout(() => {
      this.searchTag();
    }, 1000);
  }

  searchTag() {
    const req = {
      'tagName.equals': this.name
    };
    this.tagService.query(req).subscribe(response => {
      this.exist = response.body.length > 0 && !this.tag;
      this.typing = false;
    });
  }
}
