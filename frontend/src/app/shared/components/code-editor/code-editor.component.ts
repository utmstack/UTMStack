import {
  Component, EventEmitter, forwardRef,
  Input, OnDestroy, OnInit, Output
} from '@angular/core';
import {ControlValueAccessor, NG_VALUE_ACCESSOR} from '@angular/forms';
import {SqlValidationService} from '../../services/code-editor/sql-validation.service';

export interface ConsoleOptions {
  value?: string;
  language?: 'sql';
  theme?: 'vs' | 'vs-dark' | 'hc-black' | string;
  minimap?: { enabled: boolean };
  renderLineHighlight?: 'none' | 'line' | 'gutter' | 'all';
  scrollbar?: {
    vertical?: 'auto' | 'visible' | 'hidden';
    horizontal?: 'auto' | 'visible' | 'hidden';
    verticalScrollbarSize?: number;
    horizontalScrollbarSize?: number;
  };
  overviewRulerLanes?: number;
  wordWrap?: 'off' | 'on' | 'wordWrapColumn' | 'bounded';
  automaticLayout?: boolean;
  lineNumbers?: 'off' | 'on';
  cursorSmoothCaretAnimation?: 'off' | 'on';
}

const SQL_KEYWORDS = ['CREATE', 'DROP', 'ALTER', 'TRUNCATE',
  'SELECT', 'INSERT', 'UPDATE', 'DELETE',
  'COMMIT', 'ROLLBACK',
  'AND', 'OR', 'NOT', 'BETWEEN', 'IN', 'LIKE', 'EXISTS',
  'COUNT', 'SUM', 'AVG', 'MIN', 'MAX',
  'FROM', 'WHERE', 'GROUP BY', 'HAVING', 'ORDER BY', 'DISTINCT',
  'JOIN', 'INNER', 'LEFT', 'RIGHT', 'FULL', 'UNION', 'INTERSECT',
  'NULL', 'TRUE', 'FALSE',
  'AS', 'CASE', 'WHEN', 'THEN', 'END',
  'LIMIT', 'OFFSET'
];

@Component({
  selector: 'app-code-editor',
  templateUrl: './code-editor.component.html',
  styleUrls: ['./code-editor.component.scss'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => CodeEditorComponent),
      multi: true
    }
  ]
})
export class CodeEditorComponent implements OnInit, OnDestroy, ControlValueAccessor {
  @Input() showFullEditor = true;
  @Input() showSuggestions = false;
  @Input() consoleOptions?: ConsoleOptions;
  @Output() execute = new EventEmitter<string>();
  @Output() clearData = new EventEmitter<void>();
  @Output() indexPatternChange = new EventEmitter<string>();
  @Input() queryError: string | null = null;
  @Input() customKeywords: string[] = [];

  sqlQuery = '';
  errorMessage = '';
  successMessage = '';
  readonly defaultOptions: ConsoleOptions = {
    value: this.sqlQuery,
    language: 'sql',
    theme: 'myCustomTheme',
    minimap: {
      enabled: false
    },
    renderLineHighlight: 'none',
    scrollbar: {
      vertical: 'auto',
      horizontal: 'hidden'
    },
    overviewRulerLanes: 0,
    wordWrap: 'on',
    automaticLayout: true,
    lineNumbers: 'on',
    cursorSmoothCaretAnimation: 'off'
  };
  private completionProvider?: monaco.IDisposable;
  private onChange = (_: any) => {};
  private onTouched = () => {};

  constructor(private sqlValidationService: SqlValidationService) {}

  ngOnInit(): void {
    this.consoleOptions = { ...this.defaultOptions, ...this.consoleOptions };
  }

  onEditorInit(editorInstance: monaco.editor.IStandaloneCodeEditor) {
    this.completionProvider = monaco.languages.registerCompletionItemProvider('sql', {
      provideCompletionItems: () => {
        const allKeywords = Array.from(new Set([
          ...SQL_KEYWORDS,
          ...(this.customKeywords || [])
        ]));

        const suggestions: monaco.languages.CompletionItem[] = allKeywords.map(k => ({
          label: k,
          kind: monaco.languages.CompletionItemKind.Text,
          insertText: k,
        }));

        return { suggestions };
      }
    });

    editorInstance.onDidChangeModelContent(() => {
      const val = editorInstance.getValue();
      this.sqlQuery = val;
      this.onChange(val);
      this.onTouched();
    });
  }

  executeQuery(): void {
    this.clearMessages();
    this.errorMessage = this.sqlValidationService.validateSqlQuery(this.sqlQuery);
    if (this.errorMessage) {
      return;
    }

    try {
      const cleanedQuery = this.sqlQuery.replace(/\n/g, ' ');
      this.execute.emit(cleanedQuery);
    } catch (err) {
      this.errorMessage = err instanceof Error ? err.message : String(err);
    }
  }

  clearQuery(): void {
    this.sqlQuery = '';
    this.clearMessages();
    this.clearData.emit();
  }

  formatQuery(): void {
    this.clearMessages();
    this.sqlQuery = this.formatSql(this.sqlQuery);
  }

  private formatSql(sql: string): string {
    let formatted = sql;

    SQL_KEYWORDS.forEach(keyword => {
      const regex = new RegExp(`\\b${keyword}\\b`, 'gi');
      formatted = formatted.replace(regex, `\n${keyword}`);
    });

    return formatted.trim();
  }

  copyQuery(): void {
    this.clearMessages();
    (navigator as any).clipboard.writeText(this.sqlQuery);
    this.successMessage = 'Query copied to clipboard.';
  }

  clearMessages(): void {
    console.log('Query:', this.extractIndexPattern(this.sqlQuery));
    this.errorMessage = '';
    this.successMessage = '';
  }

  writeValue(value: any): void {
    this.sqlQuery = value || '';
  }

  registerOnChange(fn: any): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn;
  }

  setDisabledState?(isDisabled: boolean): void {
    // Optional: handle disabled state
  }

   extractIndexPattern(sql: string): string | null {
    const normalized = sql
      .replace(/\s+/g, ' ')
      .toLowerCase();

    const fromIndex = normalized.indexOf(' from ');
    if (fromIndex === -1) { return null; }

    const start = fromIndex + 6;

    const keywords = [' where ', ' group by ', ' order by ', ' limit ', ' having '];

    let end = normalized.length;
    for (const kw of keywords) {
      const idx = normalized.indexOf(kw, start);
      if (idx !== -1 && idx < end) {
        end = idx;
      }
    }

    const originalFragment = normalized.substring(start, end).trim();

    if (originalFragment.length > 0) {
     const indexPatternSelected = this.customKeywords.find(keyword => keyword === originalFragment);

     if (indexPatternSelected) {
       this.indexPatternChange.emit(indexPatternSelected);
     }
   }
  }

  ngOnDestroy(): void {
    if (this.completionProvider) {
      this.completionProvider.dispose();
      this.completionProvider = undefined;
    }
  }

}
