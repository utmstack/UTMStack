export class ToolboxFeatureMagicType {
  show?: boolean;
  type?: string[];
  title?: {
    line?: string,
    bar?: string,
    stack?: string,
    tiled?: string
  };
  option?: any;


  constructor(show?: boolean,
              type?: string[],
              title?: { line?: string; bar?: string; stack?: string; tiled?: string },
              option?: any) {
    this.show = show;
    this.type = type;
    this.title = title;
    this.option = option;
  }
}
