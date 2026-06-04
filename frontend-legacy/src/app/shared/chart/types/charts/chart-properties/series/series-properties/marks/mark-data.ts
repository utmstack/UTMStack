export class MarkData {
  type?: 'min' | 'max' | 'average';
  name?: string;

  constructor(type?: 'min' | 'max' | 'average',
              name?: string) {
    this.type = type;
    this.name = name;
  }
}
