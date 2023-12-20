import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {TargetService} from '../../../scanner-config/target/shared/services/target.service';

@Component({
  selector: 'app-vul-used-by',
  templateUrl: './used-by.component.html',
  styleUrls: ['./used-by.component.scss']
})
export class UsedByComponent implements OnInit {
  @Input() using: {
    uuid: string,
    name: string
  }[] = [];
  @Input() type;
  @Input() name: any;
  loading = false;
  @Input() dependency: string;

  constructor(public activeModal: NgbActiveModal, private targetService: TargetService) {
  }

  ngOnInit() {
  }

}
