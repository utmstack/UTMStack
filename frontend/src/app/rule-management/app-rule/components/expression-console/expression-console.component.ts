import {Component, ElementRef, EventEmitter, forwardRef, HostListener, Input, Output, ViewChild} from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

@Component({
  selector: 'app-expression-console',
  templateUrl: './expression-console.component.html',
  styleUrls: ['./expression-console.component.scss'],
  providers: [{
    provide: NG_VALUE_ACCESSOR,
    useExisting: forwardRef(() => ExpressionConsoleComponent),
    multi: true
  }]
})
export class ExpressionConsoleComponent implements ControlValueAccessor {
  @Input() variables: { as: string; of_type: string; get: string }[] = [];
  @Output() variablesChange = new EventEmitter<{ as: string; of_type: string; get: string }[]>();
  @ViewChild('console', {}) consoleRef!: ElementRef<HTMLTextAreaElement>;

  @Output() touched = new EventEmitter<void>();

  value = '';
  suggestions: string[] = [];
  suggestionsStyle: { [key: string]: string } = {};

  onChange: any = () => {};
  onTouched: any = () => {};

  writeValue(obj: any): void {
    this.value = obj || '';
    if (this.consoleRef) { this.consoleRef.nativeElement.innerText = this.value; }
  }

  registerOnChange(fn: any): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn;
  }

  onKeyup(event: KeyboardEvent): void {
    if (event.key === ' ') {
      this.suggestions = [];
      return;
    }

    if (event.key .trim() === '') {
      this.suggestions = [];
      return;
    }

    const textarea = this.consoleRef.nativeElement;
    const text = textarea.value;

    const lastWord = text.split(/\s+/).pop() || '';

    this.suggestions = this.variables
      .filter(v => v.as.startsWith(lastWord))
      .map(v => v.as);

    if (this.suggestions.length) {
      const cursorPos = this.getCursorPosition(textarea);
      this.suggestionsStyle = {
        left: `${cursorPos.x}px`,
        top: `${cursorPos.y + 20}px` // un poco más abajo
      };
    }
  }

  insertAlias(alias: string): void {
    const textarea = this.consoleRef.nativeElement;
    const cursorPos = textarea.selectionStart;
    const text = textarea.value;

    const beforeCursor = text.slice(0, cursorPos);
    const afterCursor = text.slice(cursorPos);

    const match = beforeCursor.match(/(\S+)$/);
    const wordStart = match ? cursorPos - match[0].length : cursorPos;

    const newText = text.slice(0, wordStart) + alias + ' ' + afterCursor;

    this.value = newText;
    this.suggestions = [];
    this.onChange(this.value);

    setTimeout(() => {
      textarea.focus();
      const pos = wordStart + alias.length + 1;
      textarea.selectionStart = textarea.selectionEnd = pos;
    });
  }

  removeVariable(index: number): void {
    this.variables = this.variables.filter((v, i) => i !== index);
    this.variablesChange.emit(this.variables);
  }

  getCursorPosition(textarea: HTMLTextAreaElement): { x: number, y: number } {
    const div = document.createElement('div');
    const style = window.getComputedStyle(textarea);

    div.style.position = 'absolute';
    div.style.whiteSpace = 'pre-wrap';
    div.style.visibility = 'hidden';
    div.style.font = style.font;
    div.style.padding = style.padding;
    div.style.border = style.border;
    div.style.overflow = 'auto';
    div.style.width = textarea.offsetWidth + 'px';
    div.style.lineHeight = style.lineHeight;
    div.style.letterSpacing = style.letterSpacing;
    div.style.textAlign = style.textAlign;

    const text = textarea.value.substring(0, textarea.selectionEnd);
    const span = document.createElement('span');
    span.textContent = '\u200b';
    div.textContent = text;
    div.appendChild(span);

    document.body.appendChild(div);

    const spanRect = span.getBoundingClientRect();

    document.body.removeChild(div);

    return { x: spanRect.left, y: spanRect.top };
  }

  @HostListener('document:click', ['$event'])
  onClickOutside(event: MouseEvent) {
    if (!this.consoleRef.nativeElement.contains(event.target as Node)) {
      this.suggestions = [];
    }
  }
}
