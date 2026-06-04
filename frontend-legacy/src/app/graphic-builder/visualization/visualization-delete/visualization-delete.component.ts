import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {VisualizationService} from '../shared/services/visualization.service';

@Component({
  selector: 'app-visualization-delete',
  templateUrl: './visualization-delete.component.html',
  styleUrls: ['./visualization-delete.component.scss']
})
export class VisualizationDeleteComponent implements OnInit {
  @Input() visualization: VisualizationType;
  @Input() multiple: boolean;
  @Input() selected: VisualizationType[];
  @Output() visualizationDeleted = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private visualizationService: VisualizationService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  deleteVisualization() {
    if (!this.multiple) {
      this.deleteSingleVisualization();
    } else {
      this.deleteMultipleVisualization();
    }
  }

  deleteSingleVisualization() {
    this.visualizationService.delete(this.visualization.id)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Visualization deleted successfully');
        this.activeModal.close();
        this.visualizationDeleted.emit('deleted');
      }, () => {
        this.utmToastService.showError('Error deleting visualization',
          'Error deleting visualization, please check your network and try again');
      });
  }

  deleteMultipleVisualization() {
    let idToDelete = '?';
    for (const vis of this.selected) {
      idToDelete += 'ids=' + vis.id + '&';
    }
    this.visualizationService.bulkDelete(idToDelete)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Visualizations deleted successfully');
        this.activeModal.close();
        this.visualizationDeleted.emit('deleted');
      }, () => {
        this.utmToastService.showError('Error deleting visualizations',
          'Error deleting visualizations, please check your network and try again');
      });
  }
}
