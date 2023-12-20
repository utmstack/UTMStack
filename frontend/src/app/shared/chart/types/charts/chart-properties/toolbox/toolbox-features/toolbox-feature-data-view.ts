export class ToolboxFeatureDataView {
  show?: boolean;
  title?: string;
  readOnly?: boolean;
  lang?: string[];
  backgroundColor?: string;
  textareaColor?: string;
  textareaBorderColor?: string;
  textColor?: string;
  buttonColor?: string;
  buttonTextColor?: string;

  constructor(show?: boolean,
              title?: string,
              readOnly?: boolean,
              lang?: string[],
              backgroundColor?: string,
              textareaColor?: string, textareaBorderColor?: string,
              textColor?: string,
              buttonColor?: string,
              buttonTextColor?: string) {
    this.show = show;
    this.title = title;
    this.readOnly = readOnly;
    this.lang = lang;
    this.backgroundColor = backgroundColor;
    this.textareaColor = textareaColor;
    this.textareaBorderColor = textareaBorderColor;
    this.textColor = textColor;
    this.buttonColor = buttonColor;
    this.buttonTextColor = buttonTextColor;
  }
}
