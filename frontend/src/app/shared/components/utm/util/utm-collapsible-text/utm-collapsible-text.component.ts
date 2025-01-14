import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-utm-collapsible-text',
  templateUrl: './utm-collapsible-text.component.html',
  styleUrls: ['./utm-collapsible-text.component.scss']
})
export class UtmCollapsibleTextComponent implements OnInit {
  @Input() text: string = '';
  @Input() maxLength: number = 100;
  isCollapsed: boolean = true;

  ngOnInit() {
  }

  toggleCollapse() {
    this.isCollapsed = !this.isCollapsed;
  }

  getDisplayText(text: string): string {
    if (this.isCollapsed && text.length > this.maxLength) {
      return text.substring(0, this.maxLength) + '...';
    }
    return text;
  }
}
