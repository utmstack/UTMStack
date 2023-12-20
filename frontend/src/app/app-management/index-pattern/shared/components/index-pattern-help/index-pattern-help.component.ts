import {Component, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-index-pattern-help',
  templateUrl: './index-pattern-help.component.html',
  styleUrls: ['./index-pattern-help.component.scss']
})
export class IndexPatternHelpComponent implements OnInit {

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }

}
