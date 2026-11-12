import {Component, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';

@Component({
    standalone: false,
    selector: 'app-asset-view-help',
    templateUrl: './asset-view-help.component.html',
    styleUrls: ['./asset-view-help.component.css'],
})
export class AssetViewHelpComponent implements OnInit {

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }

}
