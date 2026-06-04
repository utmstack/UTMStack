import {Component, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-asset-severity-help',
  templateUrl: './asset-severity-help.component.html',
  styleUrls: ['./asset-severity-help.component.scss']
})
export class AssetSeverityHelpComponent implements OnInit {

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }

}
