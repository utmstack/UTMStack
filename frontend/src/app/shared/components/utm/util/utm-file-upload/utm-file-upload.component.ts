import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import * as yaml from 'js-yaml';
import {UtmToastService} from '../../../../alert/utm-toast.service';

@Component({
  selector: 'app-utm-file-upload',
  templateUrl: './utm-file-upload.component.html',
  styleUrls: ['./utm-file-upload.component.scss']
})
export class UtmFileUploadComponent implements OnInit {
  files: any[] = [];
  @Input() acceptTypes: string;
  @Input() msg: string;
  @Input() multiple = true;
  @Input() shoPreview = false;
  @Input() validateFileSize = false;
  @Input() maxFileSize = 10 * 1024;
  maxFiles = 10;
  maxFilesError = false;
  @Output() fileEmit = new EventEmitter<any[]>();
  @Output() errorEmit = new EventEmitter<any>();
  messageError: string;
  previewFiles = [];

  constructor(private toastService: UtmToastService) {
  }

  ngOnInit() {}

  uploadFile(event) {

    if (event && event.length > this.maxFiles) {
      this.maxFilesError = true;
      return;
    } else {
      this.maxFilesError = false;
    }

    if (this.validateFileSize) {
      if (!event || event.length === 0) {
        return;
      }

      const invalidFile = Array.from(event).filter((file: any) => file.size > this.maxFileSize);
      event = Array.from(event).filter((file: any) => file.size < this.maxFileSize);
      invalidFile.forEach((file: any) =>
        this.toastService.showError('Error', `File "${file.name}" exceeds the ${this.maxFileSize} KB limit.`));
    }

    this.loadFileBeforeEmit(event)
      .then(files => this.emitFile(files))
      .then(docs => {
        this.messageError = null;
        this.previewFiles = docs;
        this.fileEmit.emit(docs);
      })
      .catch(error => {
        this.messageError = error;
        this.errorEmit.emit(error);
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
    if (this.files.length === 0) {
      this.previewFiles = [];
    }
    this.emitFile(this.files).then((fil) => {
      this.previewFiles = fil;
      this.fileEmit.emit(fil);
    });
  }

  async emitFile(files: any[]): Promise<any[]> {
    return new Promise<any[]>((resolve, reject) => {
      let arr: any[] = [];
      let filesProcessed = 0;

      const fileArray = Array.from(files);
      for (let index = 0; index < fileArray.length; index++) {
        const reader = new FileReader();
        reader.readAsText(fileArray[index]);

        reader.onload = () => {
          try {
            let parsedData;
            if (fileArray[index].type === 'application/json' || fileArray[index].name.endsWith('.json')) {
              parsedData = JSON.parse(reader.result as string);
            } else if (fileArray[index].type === 'application/x-yaml' || fileArray[index].name.endsWith('.yml')
              || fileArray[index].name.endsWith('.yaml')) {
              parsedData = yaml.load(reader.result as string);
            } else {
              return reject(`Unsupported file type: ${fileArray[index].name}`);
            }

            if (parsedData) {
              arr = arr.concat(parsedData);
            }

            filesProcessed++;
            if (filesProcessed === files.length) {
              resolve(arr);
            }
          } catch (error) {
            filesProcessed++;
            fileArray[index].error = `Error parsing file ${fileArray[index].name}: ${error.message}`;
          }
        };

        reader.onerror = () => {
          const errorMessage = `Error reading file: ${fileArray[index].name}`;
          console.error(errorMessage);
          reject(errorMessage);
        };
      }
    });
  }

  showError(file: any) {
    file.showError = !file.showError;
  }
}
