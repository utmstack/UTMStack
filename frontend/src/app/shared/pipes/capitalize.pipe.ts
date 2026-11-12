import {Pipe, PipeTransform} from '@angular/core';

@Pipe({
    standalone: false,
    name: 'capitalize',
})
export class CapitalizePipe implements PipeTransform {

  transform(value: string, args?: any): any {
    return value.charAt(0).toUpperCase() + value.substring(1, value.length);
  }

}
