import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {JhiBase64Service} from 'ng-jhipster';

@Component({
  selector: 'app-utm-input-file-upload',
  templateUrl: './utm-input-file-upload.component.html',
  styleUrls: ['./utm-input-file-upload.component.scss']
})
export class UtmInputFileUploadComponent implements OnInit {
  @Input() acceptTypes: string;
  @Output() fileEmit = new EventEmitter<any>();
  @Input() img: any;

  constructor(private base64Service: JhiBase64Service, public sanitizer: DomSanitizer) {
  }

  ngOnInit() {
  }

  uploadFile(targetElement: any) {
    const reader = new FileReader();
    reader.readAsDataURL(targetElement[0]);
    reader.onload = () => {
      // this.img = 'data:' + targetElement[0].type + ';base64,' + this.base64Service.encode(reader.result.toString());
      this.img = reader.result;
      // console.log(this.img);
      this.fileEmit.emit(this.img);
    };
  }

  removeImg() {
    this.fileEmit.emit(null);
  }
}
