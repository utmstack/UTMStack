import {Pipe, PipeTransform} from '@angular/core';

@Pipe({
  name: 'highlight'
})
export class HighlightPipe implements PipeTransform {
  transform(text: string, search?: string, cssClass: string = 'text-white bg-primary-300'): string {
    if (search) {
      let pattern = search.replace(/[\-\[\]\/\{\}\(\)\*\+\?\.\\\^\$\|]/g, '\\$&');
      pattern = pattern.split(' ').filter((t) => {
        return t.length > 0;
      }).join('|');
      const regex = new RegExp(pattern, 'gi');
      return search ? text.replace(regex, (match) => `<mark class="${cssClass} p-0">${match}</mark>`) : text;
    } else {
      return text;
    }
  }
}
