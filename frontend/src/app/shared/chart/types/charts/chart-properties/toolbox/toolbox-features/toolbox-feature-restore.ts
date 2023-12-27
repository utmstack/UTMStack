export class ToolboxFeatureRestore {
  show?: boolean;
  title?: string;

  constructor(show?: boolean, title?: string) {
    this.show = show;
    this.title = title ? title : 'Restore';
  }
}
