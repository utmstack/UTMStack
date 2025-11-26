import {
  Component, EventEmitter,
  Input, OnChanges, OnInit, Output, SimpleChanges
} from '@angular/core';

interface MonacoOptions {
  value?: string;
  language?: string;
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
  automaticLayout: boolean;
}

@Component({
  selector: 'app-code-editor',
  templateUrl: './code-editor.component.html',
  styleUrls: ['./code-editor.component.scss']
})
export class CodeEditorComponent implements OnInit, OnChanges {
  @Input() consoleOptions?: MonacoOptions;
  @Output() execute = new EventEmitter<string>();
  @Input() queryError: string | null = null;

  isExecuting = false;
  sqlQuery = '';
  errorMessage = '';
  successMessage = '';

  readonly defaultOptions: MonacoOptions = {
    value: this.sqlQuery,
    language: 'sql',
    theme: 'myCustomTheme',
    minimap: {
      enabled: false
    },
    renderLineHighlight: 'none',
    scrollbar: {
      vertical: 'auto',
      horizontal: 'none'
    },
    overviewRulerLanes: 0,
    wordWrap: 'on',
    automaticLayout: true
  };

  constructor() {}

  ngOnInit(): void {
    this.consoleOptions = { ...this.defaultOptions, ...this.consoleOptions };
  }

  executeQuery(): void {
    this.resetMessages();

    const query = this.sqlQuery ? this.sqlQuery.trim() : '';
    if (!query) {
      this.errorMessage = 'The query cannot be empty.';
      return;
    }

    const validationError = this.validateSqlQuery(query);
    if (validationError) {
      this.errorMessage = validationError;
      return;
    }

    try {
      const cleanedQuery = query.replace(/\n/g, ' ');
      this.execute.emit(cleanedQuery);
    } catch (err) {
      this.errorMessage = err instanceof Error ? err.message : String(err);
    }
  }

  clearQuery(): void {
    this.sqlQuery = '';
    this.resetMessages();
  }

  formatQuery(): void {
    this.resetMessages();
    this.sqlQuery = this.formatSql(this.sqlQuery);
  }

  private formatSql(sql: string): string {
    const keywords = ['SELECT', 'FROM', 'WHERE', 'JOIN', 'LEFT', 'RIGHT', 'INNER', 'ON', 'GROUP BY', 'ORDER BY', 'LIMIT'];
    let formatted = sql;

    keywords.forEach(keyword => {
      const regex = new RegExp(`\\b${keyword}\\b`, 'gi');
      formatted = formatted.replace(regex, `\n${keyword}`);
    });

    return formatted.trim();
  }

  copyQuery(): void {
    this.resetMessages();
    (navigator as any).clipboard.writeText(this.sqlQuery);
    this.successMessage = 'Query copied to clipboard.';
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes.queryError && this.queryError) {
      this.resetMessages();
      this.errorMessage = this.queryError;
      this.isExecuting = false;
    }
  }

  resetMessages(): void {
    this.errorMessage = '';
    this.successMessage = '';
  }

  private validateSqlQuery(query: string): string | null {
    const trimmed = query.trim().replace(/;+\s*$/, '');
    const upper = trimmed.toUpperCase();

    const startPattern = /^\s*SELECT\b/i;
    if (!startPattern.test(trimmed)) {
      return 'Query must start with SELECT.';
    }

    const minimalPattern = /^\s*SELECT\s+.+\s+FROM\s+.+/is;
    if (!minimalPattern.test(trimmed)) {
      return 'Query must be at least: SELECT <columns> FROM <table>.';
    }

    const forbiddenPattern = /\b(INSERT|UPDATE|DELETE|DROP|ALTER|CREATE|REPLACE|TRUNCATE|MERGE|GRANT|REVOKE|EXEC|EXECUTE|COMMIT|ROLLBACK|INTO)\b/i;
    const commentPattern = /(--.*?$|\/\*.*?\*\/)/gm;
    const allowedFunctions = new Set(['COUNT', 'AVG', 'MIN', 'MAX', 'SUM']);

    if (forbiddenPattern.test(upper)) {
      return 'Query contains forbidden SQL keywords.';
    }
    if (commentPattern.test(trimmed)) {
      return 'Query must not contain SQL comments (-- or /* */).';
    }
    if (trimmed.includes(';')) {
      return 'Query must not contain internal semicolons.';
    }
    if (!this.balancedQuotes(trimmed)) {
      return 'Quotes are not balanced.';
    }
    if (!this.balancedParentheses(trimmed)) {
      return 'Parentheses are not balanced.';
    }

    const functions = this.extractFunctions(upper);
    for (const func of functions) {
      if (!allowedFunctions.has(func)) {
        return `Unsupported SQL function: ${func}.`;
      }
    }

    return null;
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
}
