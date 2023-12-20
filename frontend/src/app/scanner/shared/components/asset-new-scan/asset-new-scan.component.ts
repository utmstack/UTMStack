import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ScheduleService} from '../../../scanner-config/schedule/shared/services/schedule.service';
import {TargetService} from '../../../scanner-config/target/shared/services/target.service';
import {TaskService} from '../../../scanner-config/task/shared/services/task.service';
import {TargetModel} from '../../model/target.model';

@Component({
  selector: 'app-asset-new-scan',
  templateUrl: './asset-new-scan.component.html',
  styleUrls: ['./asset-new-scan.component.scss']
})
export class AssetNewScanComponent implements OnInit {
  loadingTarget = true;
  targets: TargetModel[] = [];

  constructor(private taskService: TaskService,
              private utmToastService: UtmToastService,
              private scheduleService: ScheduleService,
              private targetService: TargetService) {
  }

  ngOnInit() {
    this.getTargets();
  }


  getTargets() {
    this.targetService.query({size: 1000}).subscribe(
      (res: HttpResponse<any>) => {
        this.targets = res.body;
        this.loadingTarget = false;
      },
      (res: HttpResponse<any>) => {
        this.loadingTarget = false;
      }
    );
  }

  newTarget() {

  }
}
