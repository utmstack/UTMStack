import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-utm-file-upload',
  templateUrl: './utm-file-upload.component.html',
  styleUrls: ['./utm-file-upload.component.scss']
})
export class UtmFileUploadComponent implements OnInit {
  files: any[] = [];
  @Input() acceptTypes: string;
  @Input() msg: string;
  @Output() fileEmit = new EventEmitter<any[]>();

  constructor() {
  }

  ngOnInit() {
  }

  uploadFile(event) {
    this.loadFileBeforeEmit(event).then(files => {
      this.emitFile(files).then(fil => {
        this.fileEmit.emit(fil);
      });
    });
  }

  loadFileBeforeEmit(event): Promise<any[]> {
    return new Promise<any[]>((resolve) => {
      for (const element of event) {
        this.files.push(element);
      }
      resolve(this.files);
    });
  }

  deleteAttachment(index) {
    this.files.splice(index, 1);
    this.emitFile(this.files).then(fil => {
      this.fileEmit.emit(fil);
    });
  }

  async emitFile(files): Promise<any[]> {
    return new Promise<any[]>((resolve) => {
      let arr: any[] = [];
      for (const file of files) {
        const reader = new FileReader();
        reader.readAsText(file);
        reader.onload = () => {
          arr = arr.concat(JSON.parse(reader.result.toString()));
        };
      }
      setTimeout(() => resolve(arr), 1000);
    });
  }

}
