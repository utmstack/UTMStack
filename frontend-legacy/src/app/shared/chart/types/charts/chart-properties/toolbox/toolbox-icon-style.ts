export class ToolboxIconStyle {
  color: string;
  borderColor: string;
  borderWidth: number;
  borderType: 'solid' | 'dashed' | 'dotted';


  constructor(color: string,
              borderColor: string,
              borderWidth: number,
              borderType: 'solid' | 'dashed' | 'dotted') {
    this.color = color ? color : '#0d47a1';
    this.borderColor = borderColor ? borderColor : '#0d47a1';
    this.borderWidth = borderWidth ? borderWidth : 1;
    this.borderType = borderType ? borderType : 'solid';
  }
}
