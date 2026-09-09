import {getCurrentDateTimeString} from './date.util';

export function replaceBreakLine(doc: string): string {
  if (doc) {
    let msg = doc.split('\n').join('<br>');
    msg = String(msg).split('\t\t').join('&nbsp;');
    msg = String(msg).split('\t').join('&nbsp;&nbsp;');
    return msg;
  } else {
    return doc;
  }
}

/**
 * Escapes HTML special characters so the result is safe to render through a
 * `safe:'html'` binding. Intended for untrusted content (agent/IR command
 * output) that must keep its newlines but must NOT keep any markup.
 */
export function escapeHtml(value: string): string {
  if (!value) {
    return value;
  }
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

/**
 * Escapes untrusted text, then converts newlines/tabs for `innerHTML` display.
 * Escape happens first so attacker markup is neutralized before the intended
 * `<br>`/`&nbsp;` markers are inserted.
 */
export function escapeAndBreakLine(doc: string): string {
  return replaceBreakLine(escapeHtml(doc));
}

export function normalizeString(str: string): string {
  str = str.replace(/\W+(?!$)/g, '-').toLowerCase();
  str = str.replace(/\W$/, '').toLowerCase();
  return str;
}


export const getElementPrefix = (inputString): string => {
  const regex = /\b(IR(?:A)?-\d+)\s/g;
  return inputString.match(regex);
};

export const createElementPrefix = (prefix: string): string => {
  return `${prefix}-${getCurrentDateTimeString()} `;
};
