export class ToolboxFeatureSaveImage {
  show?: boolean;
  type?: string;
  name?: string;
  title?: string;

  constructor(show?: boolean, type?: string, name?: string, title?: string) {
    this.show = show;
    this.type = type;
    this.name = name;
    this.title = title;
  }
}
