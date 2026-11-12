import {Pipe, PipeTransform} from '@angular/core';

@Pipe({
    standalone: false,
    name: 'highlight',
})
export class HighlightPipe implements PipeTransform {
  transform(text: string, search?: string, cssClass: string = 'text-white bg-primary-300'): string {
    if (!text) { return text; }
    const escape = (s: string): string => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
    const safeText = escape(text);
    if (!search) { return safeText; }
    const escapedTerms = search.split(' ').filter((t) => t.length > 0).map((t) => escape(t.replace(/[\-\[\]\/\{\}\(\)\*\+\?\.\\\^\$\|]/g, '\\$&')));
    if (escapedTerms.length === 0) { return safeText; }
    const regex = new RegExp(escapedTerms.join('|'), 'gi');
    return safeText.replace(regex, (match) => `<mark class="${cssClass} p-0">${match}</mark>`);
  }
}
