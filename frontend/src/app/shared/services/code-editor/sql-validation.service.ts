import { Injectable } from '@angular/core';

@Injectable({ providedIn: 'root' })
export class SqlValidationService {

  public validateSqlQuery(sqlQuery: string): string | null {
    const query = sqlQuery ? sqlQuery.trim() : '';
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
