export class ToolboxFeatureDataZoom {
  show?: boolean;
  title?: {
    zoom?: string,
    back?: string
  };


  constructor(show?: boolean, title?: { zoom?: string; back?: string }) {
    this.show = show;
    this.title = title;
  }
}
