import {Component, EventEmitter,
  Input, OnChanges, OnInit, Output, SimpleChanges} from '@angular/core';

@Component({
  selector: 'app-code-editor',
  templateUrl: './code-editor.component.html',
  styleUrls: ['./code-editor.component.scss']
})
export class CodeEditorComponent implements OnInit, OnChanges {
  @Output() execute = new EventEmitter<string>();
  @Input() queryError: string | null = null;

  isExecuting = false;
  sqlQuery = '';
  errorMessage = '';
  successMessage = '';

  constructor() {}

  ngOnInit(): void {}

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
      this.execute.emit(query);
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
    const forbiddenPattern = /\b(INSERT|UPDATE|DELETE|DROP|ALTER|CREATE|REPLACE|TRUNCATE|MERGE|GRANT|REVOKE|EXEC|EXECUTE|COMMIT|ROLLBACK|INTO)\b/i;
    const commentPattern = /(--.*?$|\/\*.*?\*\/)/gm;
    const allowedFunctions = new Set(['COUNT', 'AVG', 'MIN', 'MAX', 'SUM']);

    if (!startPattern.test(trimmed)) {
      return 'Query must start with SELECT.';
    }
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

  private extractFunctions(upperQuery: string): Set<string> {
    const funcPattern = /\b([A-Z]+)\s*\(/g;
    const funcs = new Set<string>();
    const match = funcPattern.exec(upperQuery);
    while (match) {
      funcs.add(match[1]);
    }
    return funcs;
  }

}
