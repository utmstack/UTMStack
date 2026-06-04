import {Component, Input, OnInit} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';

@Component({
  selector: 'app-utm-pdf-preview',
  templateUrl: './utm-pdf-preview.component.html',
  styleUrls: ['./utm-pdf-preview.component.scss']
})
export class UtmPdfPreviewComponent implements OnInit {
  /**
   * BLOB PDF
   */
  @Input() pdf: any;
  @Input() width = '100%';
  @Input() height = '550px';
  preview: any;

  constructor(private sanitizer: DomSanitizer) {
  }

  ngOnInit() {
    this.preview = this.sanitizer.bypassSecurityTrustResourceUrl(URL.createObjectURL(this.pdf));
  }

}
