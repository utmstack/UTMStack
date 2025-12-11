import {
  Component, EventEmitter, forwardRef,
  Input, OnInit, Output
} from '@angular/core';
import {ControlValueAccessor, NG_VALUE_ACCESSOR} from "@angular/forms";

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
export class CodeEditorComponent implements OnInit, ControlValueAccessor {
  @Input() showHeader = true;
  @Input() consoleOptions?: ConsoleOptions;
  @Output() execute = new EventEmitter<string>();
  @Output() clearData = new EventEmitter<void>();
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
    lineNumbers: 'on'
  };
  private onChange = (_: any) => {};
  private onTouched = () => {};
  constructor() {}

  ngOnInit(): void {
    this.consoleOptions = { ...this.defaultOptions, ...this.consoleOptions };
  }

  onEditorInit(editor: monaco.editor.IStandaloneCodeEditor) {
    monaco.languages.registerCompletionItemProvider('sql', {
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
    editor.onDidChangeModelContent(() => {
      const val = editor.getValue();
      this.sqlQuery = val;
      this.onChange(val);
      this.onTouched();
    });
  }

  executeQuery(): void {
    //TODO: ELENA comprobar cambio para logExpplorer caso de cadena con solo espacios
    this.clearMessages();
    this.validateSqlQuery();
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
    this.errorMessage = '';
    this.successMessage = '';
  }

  public validateSqlQuery(): string | null {
    const query = this.sqlQuery ? this.sqlQuery.trim() : '';
    let message: string | null = null;

    const trimmed = query.replace(/;+\s*$/, '');
    const upper = trimmed.toUpperCase();

    const rules: { test: boolean; message: string }[] = [
      { test: !query, message: 'The query cannot be empty.' },
      { test: !/^\s*SELECT\b/i.test(trimmed), message: 'Query must start with SELECT.' },
      { test: !/^\s*SELECT\s+.+\s+FROM\s+.+/is.test(trimmed), message: 'Query must be at least: SELECT <columns> FROM <table>.' },
      {
        test: new RegExp(
          '\\b(INSERT|UPDATE|DELETE|DROP|ALTER|CREATE|REPLACE|TRUNCATE|MERGE|GRANT|REVOKE|EXEC|EXECUTE|COMMIT|ROLLBACK|INTO)\\b',
          'i'
        ).test(upper),
        message: 'Query contains forbidden SQL keywords.'
      },
      { test: /(--.*?$|\/\*.*?\*\/)/gm.test(trimmed), message: 'Query must not contain SQL comments (-- or /* */).' },
      { test: trimmed.includes(';'), message: 'Query must not contain internal semicolons.' },
      { test: !this.balancedQuotes(trimmed), message: 'Quotes are not balanced.' },
      { test: !this.balancedParentheses(trimmed), message: 'Parentheses are not balanced.' },
      { test: this.hasMisplacedCommas(trimmed), message: 'Query contains misplaced commas.' },
      { test: this.hasSubqueryWithoutAlias(trimmed), message: 'Subquery in FROM must have an alias.' },
    ];

    for (const rule of rules) {
      if (rule.test) {
        message = rule.message;
        break;
      }
    }

    if (!message) {
      const allowedFunctions = new Set(['COUNT', 'AVG', 'MIN', 'MAX', 'SUM']);
      const functions = this.extractFunctions(upper);
      for (const func of functions) {
        if (!allowedFunctions.has(func)) {
          message = `Unsupported SQL function: ${func}.`;
          break;
        }
      }
    }
    this.errorMessage = message;
    return message;
  }


  private balancedParentheses(query: string): boolean {
    let count = 0;
    for (const c of query) {
      if (c === '(') {
        count++;
      } else if (c === ')') {
        count--;
      }
      if (count < 0) {
        return false;
      }
    }
    return count === 0;
  }

  private balancedQuotes(query: string): boolean {
    let sq = 0;
    let dq = 0;
    let escaped = false;    for (const c of query) {
      if (escaped) { escaped = false; continue; }
      if (c === '\\') { escaped = true; continue; }
      if (c === '\'') {
        sq++;
      } else {
        if (c === '"') { dq++; }
      }
    }
    return (sq % 2 === 0) && (dq % 2 === 0);
  }

  private extractFunctions(upperQuery: string): string[] {
    const funcPattern = /\b(COUNT|AVG|MIN|MAX|SUM)\s*\(/g;
    const funcs: string[] = [];

    let match: RegExpExecArray | null = funcPattern.exec(upperQuery);
    while (match !== null) {
      funcs.push(match[1]);
      match = funcPattern.exec(upperQuery);
    }

    return funcs;
  }

  private hasMisplacedCommas(query: string): boolean {
    const upperQuery = query.toUpperCase();

    if (upperQuery.startsWith('SELECT ,') || upperQuery.includes(',,')) {
      return true;
    }

    if (/,\s*FROM/i.test(upperQuery)) {
      return true;
    }

    const selectPart = query
      .replace(/^SELECT\s+/i, '')
      .replace(/\s+FROM.*$/i, '')
      .trim();

    if (selectPart.startsWith(',') || selectPart.endsWith(',')) {
      return true;
    }

    const fields = selectPart.split(',');
    for (const f of fields) {
      if (f.trim() === '') {
        return true;
      }
    }

    return false;
  }

  private hasSubqueryWithoutAlias(query: string): boolean {
    const subqueryRegex = /FROM\s*\([^)]*\)/i;
    if (!subqueryRegex.test(query)) {
      return false;
    }
    const aliasRegex = /FROM\s*\([^)]*\)\s+(AS\s+\w+|\w+)/i;
    return !aliasRegex.test(query);
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
}
