import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {Router} from '@angular/router';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {UTM_CHART_ICONS} from '../../../shared/constants/icons-chart.const';
import {ChartTypeEnum} from '../../../shared/enums/chart-type.enum';
import {RouteCallbackEnum} from '../../../shared/enums/route-callback.enum';
import {InputClassResolve} from '../../../shared/util/input-class-resolve';
import {DashboardStatusEnum} from '../../dashboard-builder/shared/enums/dashboard-status.enum';
import {cleanVisualizationData} from '../../shared/util/visualization/visualization-cleaner.util';
import {VisualizationService} from '../shared/services/visualization.service';

@Component({
  selector: 'app-visualization-save',
  templateUrl: './visualization-save.component.html',
  styleUrls: ['./visualization-save.component.scss']
})
export class VisualizationSaveComponent implements OnInit {
  @Input() callback: RouteCallbackEnum;
  @Input() visualization: VisualizationType;
  @Input() mode: string;
  @Output() visualizationCreated = new EventEmitter<number>();
  visualizationToSave: VisualizationType;
  visSaveForm: FormGroup;
  creating = false;
  saveMode: boolean;

  constructor(public activeModal: NgbActiveModal,
              public inputClassResolve: InputClassResolve,
              private fb: FormBuilder,
              private router: Router,
              private utmToastService: UtmToastService,
              private visualizationService: VisualizationService) {
  }

  ngOnInit() {
    this.initFormSaveVis();
    cleanVisualizationData(this.visualization).then(visualization => {
      this.visualizationToSave = visualization;
      if (this.mode === 'edit') {
        this.visSaveForm.get('name').setValue(this.visualizationToSave.name);
        this.visSaveForm.get('description').setValue(this.visualizationToSave.description);
      }
    });
  }

  initFormSaveVis() {
    this.visSaveForm = this.fb.group(
      {
        name: ['', Validators.required],
        description: [''],
      }
    );
  }


  saveVisualization() {
    this.visualizationToSave.name = this.visSaveForm.get('name').value;
    this.visualizationToSave.description = this.visSaveForm.get('description').value;
    if (this.visualizationToSave.chartType !== ChartTypeEnum.TEXT_CHART) {
      this.convertPropertiesToString();
    }
    this.creating = true;

    if (this.mode === 'edit' && !this.saveMode) {
      this.editVisualization();
    } else {
      this.createVisualization();
    }
  }

  createVisualization() {
    this.visualizationToSave.id = null;
    this.visualizationService.create(this.visualizationToSave).subscribe(response => {
      this.operationSuccessfully(response.body, 'created');
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating visualization',
        error1.error.statusText);
    });
  }

  editVisualization() {
    this.visualizationService.update(this.visualizationToSave).subscribe(response => {
      this.operationSuccessfully(response.body, 'edited');
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error editing visualization',
        error1.error.statusText);
    });
  }

  operationSuccessfully(visualization: VisualizationType, mode: 'edited' | 'created') {
    this.utmToastService.showSuccessBottom('Visualization ' + mode + ' successfully');
    if (this.visualizationToSave.chartType !== ChartTypeEnum.TEXT_CHART) {
      this.convertPropertiesToObject();
    }
    this.activeModal.close();
    this.visualizationCreated.emit(visualization.id);
    this.navigateOnSave(visualization);
  }

  navigateOnSave(visualization: VisualizationType) {
    if (this.callback && this.callback === RouteCallbackEnum.DASHBOARD) {
      this.router.navigate(['/creator/dashboard/builder'],
        {
          queryParams:
            {
              addVisualization: visualization.id,
              mode: DashboardStatusEnum.DRAFT
            }
        });
    } else {
      this.router.navigate(['/creator/visualization/list']);
    }
  }

  chartIconResolver(chartType: string) {
    return UTM_CHART_ICONS[chartType];
  }

  convertPropertiesToString() {
    this.visualizationToSave.chartConfig = JSON.stringify(this.visualizationToSave.chartConfig);
    this.visualizationToSave.chartAction = JSON.stringify(this.visualizationToSave.chartAction);
    // this.visualizationToSave.filterType = JSON.stringify(this.visualizationToSave.filterType);
  }

  convertPropertiesToObject() {
    this.visualizationToSave.chartConfig = JSON.parse(this.visualizationToSave.chartConfig);
    this.visualizationToSave.chartAction = JSON.parse(this.visualizationToSave.chartAction);
    // if (typeof this.visualizationToSave.filterType === 'string') {
    //   this.visualizationToSave.filterType = JSON.parse(this.visualizationToSave.filterType);
    // }
  }

  saveAsNew($event: boolean) {
    this.saveMode = $event;
    if ($event) {
      this.visSaveForm.get('name').setValue(this.visualizationToSave.name + '-Clone');
    } else {
      this.visSaveForm.get('name').setValue(this.visualizationToSave.name);
    }
  }
}
