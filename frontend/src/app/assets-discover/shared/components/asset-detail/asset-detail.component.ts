import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {NetScanType} from '../../types/net-scan.type';

@Component({
  selector: 'app-asset-detail',
  templateUrl: './asset-detail.component.html',
  styleUrls: ['./asset-detail.component.scss']
})
export class AssetDetailComponent implements OnInit {
  @Output() refreshData = new EventEmitter<boolean>();
  @Input() asset: NetScanType;


  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }


}
