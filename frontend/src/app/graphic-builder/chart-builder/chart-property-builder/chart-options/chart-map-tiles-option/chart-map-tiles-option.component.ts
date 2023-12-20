import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {LeafletTilesType} from '../../../../../shared/chart/types/map/leaflet/leaflet-tiles.type';
import {ChartTypeEnum} from '../../../../../shared/enums/chart-type.enum';
import {InputClassResolve} from '../../../../../shared/util/input-class-resolve';

@Component({
  selector: 'app-chart-map-tiles-option',
  templateUrl: './chart-map-tiles-option.component.html',
  styleUrls: ['./chart-map-tiles-option.component.scss']
})
export class ChartMapTilesOptionComponent implements OnInit {
  @Output() mapTilesOption = new EventEmitter<LeafletTilesType[]>();
  @Input() chartType: ChartTypeEnum;
  chartTypeEnum = ChartTypeEnum;
  formTiles: FormGroup;
  viewMapTiles = false;

  constructor(private fb: FormBuilder, public inputClass: InputClassResolve) {
  }

  get tiles() {
    return this.formTiles.get('tiles') as FormArray;
  }

  ngOnInit() {
    this.initFormTiles();
    this.formTiles.valueChanges.subscribe(val => {
      this.mapTilesOption.emit(val.tiles);
    });
  }

  deleteTiles(index: number) {
    this.tiles.removeAt(index);
  }

  addTile() {
    this.tiles.push(this.fb.group({
        label: ['', Validators.required],
        urlTemplate: ['', Validators.required],
        options: this.fb.group({
          attribution: ['']
        })
      }, {updateOn: 'change'})
    );
  }

  viewTilesOptionProperties() {
    this.viewMapTiles = this.viewMapTiles ? false : true;
  }

  private initFormTiles() {
    this.formTiles = this.fb.group({
      tiles: this.fb.array([])
    });
    this.tiles.push(this.fb.group({
        label: ['Open Street Map', Validators.required],
        urlTemplate: ['https://{s}.tile.openstreetmap.fr/hot/{z}/{x}/{y}.png', Validators.required],
        options: this.fb.group({
          attribution: ['']
        })
      }, {updateOn: 'change'})
    );
    this.mapTilesOption.emit(this.tiles.value);
  }

}
