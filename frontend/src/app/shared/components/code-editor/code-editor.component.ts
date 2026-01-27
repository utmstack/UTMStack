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

const SQL_KEYWORDS = [
  'CREATE', 'DROP', 'ALTER', 'TRUNCATE',
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
  @Input() queryError: string | null = null;
  @Input() customKeywords: string[] = [];

  @Output() indexPatternChange = new EventEmitter<string>();
  @Output() execute = new EventEmitter<string>();
  @Output() clearData = new EventEmitter<void>();

  sqlQuery = '';
  errorMessage = '';
  successMessage = '';

  private editorInstance?: monaco.editor.IStandaloneCodeEditor;
  private completionProvider?: monaco.IDisposable;

  private updatingFromOutside = false;

  readonly defaultOptions: ConsoleOptions = {
    value: this.sqlQuery,
    language: 'sql',
    theme: 'myCustomTheme',
    minimap: { enabled: false },
    renderLineHighlight: 'none',
    scrollbar: { vertical: 'auto', horizontal: 'hidden' },
    overviewRulerLanes: 0,
    wordWrap: 'on',
    automaticLayout: true,
    lineNumbers: 'on',
    cursorSmoothCaretAnimation: 'off'
  };

  constructor(private sqlValidationService: SqlValidationService) {}

  ngOnInit(): void {
    this.consoleOptions = { ...this.defaultOptions, ...this.consoleOptions };
  }

  private onChange = (_: any) => {};
  private onTouched = () => {};

  ngOnDestroy(): void {
    if (this.completionProvider) {
      this.completionProvider.dispose();
      this.completionProvider = undefined;
    }
  }

  onEditorInit(editor: monaco.editor.IStandaloneCodeEditor) {
    this.editorInstance = editor;

    if (this.sqlQuery) {
      editor.setValue(this.sqlQuery);
    }

    this.completionProvider = monaco.languages.registerCompletionItemProvider('sql', {
      provideCompletionItems: () => {
        const allKeywords = Array.from(new Set([
          ...SQL_KEYWORDS,
          ...(this.customKeywords || [])
        ]));

        const suggestions = allKeywords.map(k => ({
          label: k,
          kind: monaco.languages.CompletionItemKind.Text,
          insertText: k
        }));

        return { suggestions };
      }
    });

    editor.onDidChangeModelContent(() => {
      if (this.updatingFromOutside) {
        return;
      }

      const val = editor.getValue();
      this.sqlQuery = val;
      this.onChange(val);
      this.onTouched();
    });
  }

  writeValue(value: any): void {
    this.updatingFromOutside = true;

    this.sqlQuery = value || '';

    if (this.editorInstance) {
      this.editorInstance.setValue(this.sqlQuery);
    }

    this.updatingFromOutside = false;
  }

  registerOnChange(fn: any): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn;
  }

  setDisabledState?(isDisabled: boolean): void {
    if (this.editorInstance) {
      this.editorInstance.updateOptions({ readOnly: isDisabled });
    }
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
    this.onChange('');
  }

  formatQuery(): void {
    this.clearMessages();
    this.sqlQuery = this.formatSql(this.sqlQuery);
    if (this.editorInstance) {
      this.editorInstance.setValue(this.sqlQuery);
    }
    this.onChange(this.sqlQuery);
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
    this.errorMessage = '';
    this.successMessage = '';
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
}
