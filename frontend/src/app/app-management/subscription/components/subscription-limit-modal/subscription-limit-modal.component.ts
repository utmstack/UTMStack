import { Component, Input, OnInit } from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import { Router } from '@angular/router';

@Component({
  selector: 'app-subscription-limit-modal',
  templateUrl: './subscription-limit-modal.component.html',
  styleUrls: ['./subscription-limit-modal.component.scss']
})
export class SubscriptionLimitModalComponent implements OnInit {

  @Input() title: string;
  @Input() message: string;
  @Input() showUpgradeButton: boolean;
  @Input() icon: 'warning' | 'error' | 'info' = 'warning';

  constructor(
    public activeModal: NgbActiveModal,
    private router: Router
  ) { }

  ngOnInit(): void {
  }

  upgrade(): void {
    this.activeModal.dismiss('upgrade');
    this.router.navigate(['/profile/plans']);
  }
}
