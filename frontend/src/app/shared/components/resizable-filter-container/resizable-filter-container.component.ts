import {Component, EventEmitter, HostListener, Input, OnInit, Output} from '@angular/core';

@Component({
  selector: 'app-resizable-filter-container',
  templateUrl: './resizable-filter-container.component.html',
  styleUrls: ['./resizable-filter-container.component.scss']
})
export class ResizableFilterContainerComponent implements OnInit {

  @Input() minWidth = 350;
  @Output() filterWidthChange = new EventEmitter<number>();

  pageWidth = window.innerWidth;
  filterWidth: number;
  tableWidth: number;

  ngOnInit() {
    this.setInitialWidth();
  }

  @HostListener('window:resize', ['$event'])
  onResizeWindows(event: any) {
    this.pageWidth = event.target.innerWidth;
    this.setInitialWidth();
  }

  setInitialWidth() {
    if (this.pageWidth > 4000) {
      this.filterWidth = 400;
    } else if (this.pageWidth > 2500) {
      this.filterWidth = 350;
    } else if (this.pageWidth > 1980) {
      this.filterWidth = 350;
    } else {
      this.filterWidth = this.minWidth;
    }
    this.tableWidth = this.pageWidth - this.filterWidth - 51;
    this.filterWidthChange.emit(this.filterWidth);
  }

  onResize(event: any) {
    if (event.rectangle.width >= this.minWidth) {
      this.filterWidth = event.rectangle.width;
      this.tableWidth = this.pageWidth - this.filterWidth - 51;
      this.filterWidthChange.emit(this.filterWidth);
    }
  }

}
