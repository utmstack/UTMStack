import {Component, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {AlertTags} from '../../../shared/types/alert/alert-tag.type';
import {calcMultipleColors} from '../../../shared/util/color-mixing';
import {AlertTagService} from '../shared/services/alert-tag.service';
import {AlertTagsCreateComponent} from './alert-tags-create/alert-tags-create.component';

@Component({
  selector: 'app-alert-tags',
  templateUrl: './alert-tags-management.component.html',
  styleUrls: ['./alert-tags-management.component.scss']
})
export class AlertTagsManagementComponent implements OnInit {
  tags: AlertTags[] = [];
  loading = true;

  constructor(private alertTagService: AlertTagService,
              private modalService: NgbModal,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
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

  editTag(tag: any) {
    const modalRef = this.modalService.open(AlertTagsCreateComponent, {centered: true, size: 'sm'});
    modalRef.componentInstance.tag = tag;
    modalRef.componentInstance.addTag.subscribe((created) => {
      this.utmToastService.showSuccessBottom('Tag renamed to ' + created);
      this.getTags();
    });
  }


  addTag() {
    const modalRef = this.modalService.open(AlertTagsCreateComponent, {centered: true, size: 'sm'});
    modalRef.componentInstance.addTag.subscribe((created) => {
      this.utmToastService.showSuccessBottom('Tag ' + created + ' added');
      this.getTags();
    });
  }

  deleteTag(tag: any) {
    this.alertTagService.delete(tag.id).subscribe(() => {
      const index = this.tags.findIndex(value => value.id === tag.id);
      this.tags.splice(index, 1);
      this.utmToastService.showSuccessBottom('Tag ' + tag.name + ' deleted successfully');
    });
  }

  openDeleteConfirmation(tag: any) {
    const deleteModalRef = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModalRef.componentInstance.header = 'Confirm delete operation';
    deleteModalRef.componentInstance.message = 'Are you sure that you want to delete the tag: ' + tag.tagName;
    deleteModalRef.componentInstance.confirmBtnText = 'Delete';
    deleteModalRef.componentInstance.confirmBtnIcon = 'icon-database-remove';
    deleteModalRef.componentInstance.confirmBtnType = 'delete';
    deleteModalRef.result.then(() => {
      this.deleteTag(tag);
    });
  }

  getMixing() {
    const colors = this.tags.map(value => value.tagColor).filter(value => value !== null);
    return calcMultipleColors(colors);
  }
}
