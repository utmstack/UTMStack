import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {LeafletMapType} from '../../../../../shared/chart/types/map/leaflet/leaflet-map.type';
import {InputClassResolve} from '../../../../../shared/util/input-class-resolve';

@Component({
  selector: 'app-map-leaflet-options',
  templateUrl: './map-leaflet-options.component.html',
  styleUrls: ['./map-leaflet-options.component.scss']
})
export class MapLeafletOptionsComponent implements OnInit {
  @Output() mapLeafletOptions = new EventEmitter<LeafletMapType>();
  formLeaflet: FormGroup;
  viewLeafletOptions = false;
  layerControlPositions: { label: string, value: string }[] =
    [
      {label: 'Top left', value: 'topleft'},
      {label: 'Top right', value: 'topright'},
      {label: 'Bottom left', value: 'bottomleft'},
      {label: 'Bottom right', value: 'bottomright'},
    ];

  constructor(private fb: FormBuilder, public inputClass: InputClassResolve) {
  }

  ngOnInit() {
    this.initFormLeafletOptions();
    this.formLeaflet.valueChanges.subscribe(value => {
      if (value.tiles) {
        this.mapLeafletOptions.emit(value);
      }
    });
  }

  initFormLeafletOptions() {
    this.formLeaflet = this.fb.group({
      center: [[0, 0], Validators.required],
      zoom: [1, Validators.required],
      roam: [true],
      layerControl: this.fb.group({
        position: ['topleft']
      }),
      tiles: [],
      lat: [0, [Validators.pattern('^-?([1-8]?[1-9]|[1-9]0)\.{1}\d{1,6}')]],
      lon: [0, [Validators.pattern('^-?([1-8]?[1-9]|[1-9]0)\.{1}\d{1,6}')]]
    });
  }

  viewLeafletOptionProperties() {
    this.viewLeafletOptions = this.viewLeafletOptions ? false : true;
  }

  setCenter() {
    this.formLeaflet.get('center').setValue([this.formLeaflet.get('lat').value, this.formLeaflet.get('lon').value]);
  }
}
