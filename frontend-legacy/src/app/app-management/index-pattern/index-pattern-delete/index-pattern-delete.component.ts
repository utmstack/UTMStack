import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {IndexPatternService} from '../../../shared/services/elasticsearch/index-pattern.service';
import {UtmIndexPattern} from '../../../shared/types/index-pattern/utm-index-pattern';

@Component({
  selector: 'app-index-pattern-delete',
  templateUrl: './index-pattern-delete.component.html',
  styleUrls: ['./index-pattern-delete.component.scss']
})
export class IndexPatternDeleteComponent implements OnInit {
  @Input() indexPattern: UtmIndexPattern;
  @Output() indexPatternDeleted = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private indexPatternService: IndexPatternService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  deleteIndexPattern() {
    this.indexPatternService.delete(this.indexPattern.id)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Index pattern deleted successfully');
        this.activeModal.close();
        this.indexPatternDeleted.emit('deleted');
      }, (error) => {
        this.utmToastService.showError('Error deleting Index pattern',
          error.error.statusText);
      });
  }
}
