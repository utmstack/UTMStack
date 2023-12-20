import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {IndexPatternService} from '../../../../services/elasticsearch/index-pattern.service';
import {UtmIndexPattern} from '../../../../types/index-pattern/utm-index-pattern';
import {IndexPatternCreateComponent} from '../index-pattern-create/index-pattern-create.component';

@Component({
  selector: 'app-index-pattern-select',
  templateUrl: './index-pattern-select.component.html',
  styleUrls: ['./index-pattern-select.component.css']
})
export class IndexPatternSelectComponent implements OnInit {
  patterns: UtmIndexPattern[];
  @Input() pattern: UtmIndexPattern;
  @Input() patterRegex;
  @Output() indexPatternChange = new EventEmitter<UtmIndexPattern>();

  constructor(public activeModal: NgbActiveModal,
              private modalService: NgbModal,
              private indexPatternService: IndexPatternService) {
  }

  ngOnInit() {
    if (!this.pattern) {
      this.getIndexPatterns('init');
    } else {
      this.getIndexPatterns();
    }
  }

  newIndexPattern() {
    const modal = this.modalService.open(IndexPatternCreateComponent, {centered: true});
    modal.componentInstance.indexPatternCreated.subscribe(pattern => {
      this.pattern = pattern;
      this.getIndexPatterns();
      this.indexPatternChange.emit(this.pattern);
    });
  }

  getIndexPatterns(init?: string) {
    const req = {
      page: 0,
      size: 1000,
      sort: 'id,asc',
      'isActive.equals': true,
    };
    this.indexPatternService.query(req).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers, init),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  onIndexPatternChange($event: any) {
    this.indexPatternChange.emit($event);
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }

  private onSuccess(data, headers, init) {
    this.patterns = data;
    if (init) {
      this.pattern = this.patterns[0];
      this.indexPatternChange.emit(this.pattern);
    }
  }
}
