import {AfterViewInit, Component, ElementRef, Input, OnInit, ViewChild} from '@angular/core';
import * as hljs from 'highlight.js';
import * as bash from 'highlight.js/lib/languages/bash.js';
import * as golang from 'highlight.js/lib/languages/go.js';
import * as java from 'highlight.js/lib/languages/java.js';
import * as javascript from 'highlight.js/lib/languages/javascript.js';
import * as json from 'highlight.js/lib/languages/json.js';
import * as powershell from 'highlight.js/lib/languages/powershell.js';
import * as python from 'highlight.js/lib/languages/python.js';
import {UtmToastService} from '../../../../alert/utm-toast.service';

@Component({
  selector: 'app-utm-code-highlight',
  templateUrl: './utm-code-highlight.component.html',
  styleUrls: ['./utm-code-highlight.component.css']
})
export class UtmCodeHighlightComponent implements OnInit, AfterViewInit {
  @Input() codeType: 'python' | 'java' | 'javascript' | 'golang' | 'bash' | 'powershell' | 'json' = 'python';
  @Input() code: string;
  @Input() showCopyCode = true;
  @ViewChild('codeElement') codeElement: ElementRef;
  @Input() maskSecret: string;
  @Input() unmaskedSecret: string;
  private str: string;
  codeLines: number[] = [];

  constructor(private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    this.str = this.code;
    const codeArray = this.code.split('\n');
    this.codeLines = Array.from({length: codeArray.length}, (_, i) => i + 1);
  }

  ngAfterViewInit() {
    hljs.registerLanguage('python', python);
    hljs.registerLanguage('java', java);
    hljs.registerLanguage('javascript', javascript);
    hljs.registerLanguage('golang', golang);
    hljs.registerLanguage('bash', bash);
    hljs.registerLanguage('powershell', powershell);
    hljs.registerLanguage('json', json);
    hljs.highlightBlock(this.codeElement.nativeElement);
  }

  copyCode() {
    const selBox = document.createElement('textarea');
    const copyText = this.maskSecret !== '' ? this.str.replace(this.maskSecret, this.unmaskedSecret) : this.str;
    selBox.style.position = 'fixed';
    selBox.style.left = '0';
    selBox.style.top = '0';
    selBox.style.opacity = '0';
    selBox.value = copyText;
    document.body.appendChild(selBox);
    selBox.focus();
    selBox.select();
    document.execCommand('copy');
    document.body.removeChild(selBox);
    this.utmToastService.showInfo('Copied!', 'Code is in clipboard');
  }
}
