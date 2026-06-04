export class WebfluxEvent {
  id: number;
  when: any;

  constructor(jsonData) {
    Object.assign(this, jsonData);
  }
}
