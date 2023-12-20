export class ToolboxFeatureBrush {
  type?: 'rect' | 'polygon' | 'lineX' | 'lineY' | 'keep' | 'clear';


  constructor(type?: 'rect' | 'polygon' | 'lineX' | 'lineY' | 'keep' | 'clear') {
    this.type = type;
  }
}
