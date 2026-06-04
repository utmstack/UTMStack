import {AppLogSourceEnum} from '../enum/app-log-source.enum';
import {AppLogTypeEnum} from '../enum/app-log-type.enum';

export class AppLogType {
  source: AppLogSourceEnum;
  '@timestamp': string;
  message: string;
  type: AppLogTypeEnum;
  'es_metadata_id': string;
}
