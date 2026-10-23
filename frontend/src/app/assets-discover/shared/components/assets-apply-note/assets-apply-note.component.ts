import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {UtmNetScanService} from '../../services/utm-net-scan.service';
import {NetScanType} from '../../types/net-scan.type';

@Component({
  selector: 'app-assets-apply-note',
  templateUrl: './assets-apply-note.component.html',
  styleUrls: ['./assets-apply-note.component.scss']
})
export class AssetsApplyNoteComponent implements OnInit {
  @Input() asset: NetScanType;
  @Input() showNote: boolean;
  @Output() applyNote = new EventEmitter<string>();
  @Output() focus = new EventEmitter<boolean>();
  creating = false;

  constructor(private utmNetScanService: UtmNetScanService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  addNote() {
    this.creating = true;
    // Dedicated note endpoint: the whole-entity update() below would persist
    // the list row as the entity and NULL out group_id/asset_type_id, which
    // the row (NetworkScanDTO) does not carry (#2967).
    this.utmNetScanService.updateNotes({value: this.asset.assetNotes, assetsIds: [this.asset.id]}).subscribe(response => {
      this.utmToastService.showSuccessBottom('Comment added successfully');
      this.applyNote.emit('success');
      this.focus.emit(false);
      this.creating = false;
    }, error => {
      this.utmToastService.showError('Error adding note',
        'Error adding comment, please try again');
      this.creating = false;
    });
  }

   onClick(event: Event) {
     event.stopPropagation();
   }

   onHidden() {
     this.focus.emit(false);
   }

   onShown() {
    this.focus.emit(true);
   }
}
