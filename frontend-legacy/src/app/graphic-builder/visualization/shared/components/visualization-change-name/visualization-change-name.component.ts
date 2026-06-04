import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {VisualizationType} from '../../../../../shared/chart/types/visualization.type';
import {UTM_CHART_ICONS} from '../../../../../shared/constants/icons-chart.const';
import {InputClassResolve} from '../../../../../shared/util/input-class-resolve';
import {VisualizationService} from '../../services/visualization.service';

@Component({
  selector: 'app-visualization-change-name',
  templateUrl: './visualization-change-name.component.html',
  styleUrls: ['./visualization-change-name.component.scss']
})
export class VisualizationChangeNameComponent implements OnInit {
  @Input() visualization: VisualizationType;
  @Output() visualizationEdited = new EventEmitter<string>();
  visSaveForm: FormGroup;
  creating = false;
  saveMode: boolean;

  constructor(public activeModal: NgbActiveModal,
              public inputClassResolve: InputClassResolve,
              private fb: FormBuilder,
              private utmToastService: UtmToastService,
              private visualizationService: VisualizationService) {
  }

  ngOnInit() {
    this.initFormSaveVis();
    this.visSaveForm.get('name').setValue(this.visualization.name);
  }

  initFormSaveVis() {
    this.visSaveForm = this.fb.group(
      {
        name: ['', Validators.required],
      }
    );
  }

  changeName() {
    this.visualization.name = this.visSaveForm.get('name').value;
    this.creating = true;
    if (typeof this.visualization.filterType !== 'string') {
      this.visualization.chartConfig = JSON.stringify(this.visualization.chartConfig);
    }
    if (typeof this.visualization.chartAction !== 'string') {
      this.visualization.chartAction = JSON.stringify(this.visualization.chartAction);
    }
    this.visualizationService.update(this.visualization).subscribe(vis => {
      this.utmToastService.showSuccessBottom('Visualization edited successfully');
      if (typeof this.visualization.filterType === 'string') {
        this.visualization.filterType = JSON.parse(this.visualization.filterType);
      }
      if (typeof this.visualization.chartAction === 'string') {
        this.visualization.chartAction = JSON.parse(this.visualization.chartAction);
      }
      this.visualization.chartConfig = JSON.parse(this.visualization.chartConfig);
      this.activeModal.close();
      this.visualizationEdited.emit(vis.body.id);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error editing visualization',
        error1.error.statusText);
    });
  }

  chartIconResolver(chartType: string) {
    return UTM_CHART_ICONS[chartType];
  }

}
