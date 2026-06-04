import {HttpErrorResponse} from '@angular/common/http';
import {ChangeDetectionStrategy, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {LocalStorageService} from 'ngx-webstorage';
import {EMPTY, Observable} from 'rxjs';
import {catchError, concatMap, filter, map,tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {CpStandardService} from '../../services/cp-standard.service';
import {ComplianceStandardType} from '../../type/compliance-standard.type';

@Component({
  selector: 'app-utm-cp-standard-select',
  templateUrl: './utm-cp-standard.component.html',
  styleUrls: ['./utm-cp-standard.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class UtmCpStandardComponent implements OnInit {
  @Input() standardId: number;
  @Input() required: boolean;
  @Output() standardSelect = new EventEmitter<number>();
  standards$: Observable<ComplianceStandardType[]>;
  selectedStandard: ComplianceStandardType;
  title = 'Select framework';

  constructor(private standardService: CpStandardService,
              private toastService: UtmToastService,
              public activeModal: NgbActiveModal,
              private $localStorage: LocalStorageService) {
  }

  ngOnInit() {
    this.standards$ = this.standardService.onRefresh$
      .pipe(filter(refresh => !!refresh),
        concatMap(() => this.standardService.fetchData({page: 0, size: 1000})),
        map((res) => res.body),
        tap(standards=>{
            this.selectedStandard =standards.find(s=>s.id==this.standardId);
        }),
        catchError((err: HttpErrorResponse) => {
          this.toastService.showError('Error',
            'Unable to retrieve the list of compliance standards. Please try again or contact support.');
          return EMPTY;
        }));

    this.standardService.notifyRefresh(true);
  }

  onSelectChange($event: Event, standard: ComplianceStandardType) {
    const isChecked = ($event.target as HTMLInputElement).checked;

    if (isChecked) {
      this.selectedStandard = standard;
    } else {
      this.selectedStandard = null;
    }
  }

  confirmSelection() {
    this.activeModal.close(this.selectedStandard);
  }
}
