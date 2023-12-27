export interface ISource {
  id?: number;
  name?: string;
  checked: boolean;
}

export class Source implements ISource {
  constructor(public id?: number, public name?: string, public checked: boolean = false) {
    this.id = id ? id : null;
    this.name = name ? name : null;
    this.checked = checked;
  }
}
