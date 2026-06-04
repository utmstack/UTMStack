import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ElasticSearchIndexService} from '../../../shared/services/elasticsearch/elasticsearch-index.service';

@Component({
  selector: 'app-index-delete',
  templateUrl: './index-delete.component.html',
  styleUrls: ['./index-delete.component.css']
})
export class IndexDeleteComponent implements OnInit {
  @Input() indexes: string[];
  @Output() indexesDeleted = new EventEmitter<boolean>();
  deleting = false;

  constructor(public activeModal: NgbActiveModal,
              private toastService: UtmToastService,
              private elasticIndexService: ElasticSearchIndexService) {
  }

  ngOnInit() {
  }

  deleteIndex() {
    this.deleting = true;
    this.elasticIndexService.deleteIndexes(this.indexes).subscribe(() => {
      this.toastService.showSuccessBottom((this.indexes.length === 1 ? 'Index' : 'Indexes ') + 'deleted successfully');
      this.deleting = false;
      this.activeModal.close();
      this.indexesDeleted.emit(true);
    }, error => {
      this.deleting = false;
      this.toastService.showError('Error', 'Error while trying to delete index, please try again or contact with us');
    });
  }
}
