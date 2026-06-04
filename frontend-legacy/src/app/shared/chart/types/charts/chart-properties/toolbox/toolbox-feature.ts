import {ToolboxFeatureBrush} from './toolbox-features/toolbox-feature-brush';
import {ToolboxFeatureDataView} from './toolbox-features/toolbox-feature-data-view';
import {ToolboxFeatureDataZoom} from './toolbox-features/toolbox-feature-data-zoom';
import {ToolboxFeatureMagicType} from './toolbox-features/toolbox-feature-magic-type';
import {ToolboxFeatureMark} from './toolbox-features/toolbox-feature-mark';
import {ToolboxFeatureRestore} from './toolbox-features/toolbox-feature-restore';
import {ToolboxFeatureSaveImage} from './toolbox-features/toolbox-feature-save-image';

export class ToolboxFeature {
  saveAsImage?: ToolboxFeatureSaveImage;
  restore?: ToolboxFeatureRestore;
  dataView?: ToolboxFeatureDataView;
  dataZoom?: ToolboxFeatureDataZoom;
  magicType?: ToolboxFeatureMagicType;
  brush?: ToolboxFeatureBrush;
  mark?: ToolboxFeatureMark;


  constructor(saveAsImage?: ToolboxFeatureSaveImage,
              restore?: ToolboxFeatureRestore,
              dataView?: ToolboxFeatureDataView,
              dataZoom?: ToolboxFeatureDataZoom,
              magicType?: ToolboxFeatureMagicType,
              brush?: ToolboxFeatureBrush,
              mark?: ToolboxFeatureMark) {
    this.saveAsImage = saveAsImage ? saveAsImage : new ToolboxFeatureSaveImage(true);
    this.restore = restore ? restore : new ToolboxFeatureRestore(true);
    this.dataView = dataView ? dataView : new ToolboxFeatureDataView(true);
    this.dataZoom = dataZoom ? dataZoom : new ToolboxFeatureDataZoom(false);
    this.magicType = magicType ? magicType : new ToolboxFeatureMagicType(false);
    this.brush = brush ? brush : new ToolboxFeatureBrush('rect');
    this.mark = mark ? mark : new ToolboxFeatureMark(false);
  }
}
