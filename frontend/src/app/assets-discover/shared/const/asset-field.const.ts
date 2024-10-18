import {ElasticDataTypesEnum} from '../../../shared/enums/elastic-data-types.enum';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {AssetFieldFilterEnum} from '../enums/asset-field-filter.enum';
import {AssetFieldEnum} from '../enums/asset-field.enum';
import {CollectorFieldFilterEnum} from '../enums/collector-field-filter.enum';

export const ASSETS_FIELDS: UtmFieldType[] = [
  // {
  //   label: 'Is alive',
  //   field: ASSET_IS_ALIVE,
  //   type: ElasticDataTypesEnum.BOOLEAN,
  //   visible: true
  // },
  {
    label: 'Source',
    field: AssetFieldEnum.ASSET_IP,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  },
  {
    label: 'Name',
    field: AssetFieldEnum.ASSET_NAME,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  },
  {
    label: 'Alias',
    field: AssetFieldEnum.ASSET_ALIAS,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  },
  // {
  //   label: 'Severity',
  //   field: AssetFieldEnum.ASSET_SEVERITY,
  //   type: ElasticDataTypesEnum.STRING,
  //   visible: true
  // },
  // {
  //   label: 'Status',
  //   field: AssetFieldEnum.ASSET_STATUS,
  //   type: ElasticDataTypesEnum.STRING,
  //   visible: true
  // },

  // {
  //   label: 'MAC',
  //   field: AssetFieldEnum.ASSET_MAC,
  //   type: ElasticDataTypesEnum.STRING,
  //   visible: true
  // },
  {
    label: 'Data Source',
    field: AssetFieldEnum.ASSET_OS,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  },
  // {
  //   label: 'Discovered by',
  //   field: AssetFieldEnum.ASSET_PROBE,
  //   type: ElasticDataTypesEnum.STRING,
  //   visible: true
  // },
  // {
  //   label: 'Discovered at',
  //   field: AssetFieldEnum.ASSET_DISCOVERED,
  //   type: ElasticDataTypesEnum.DATE,
  //   visible: true
  // },
  // {
  //   label: 'Metrics',
  //   field: AssetFieldEnum.ASSET_METRICS,
  //   type: ElasticDataTypesEnum.OBJECT,
  //   visible: true
  // },
  // {
  //   label: 'Discovery mode',
  //   field: AssetFieldEnum.ASSET_REGISTER_MODE,
  //   type: ElasticDataTypesEnum.STRING,
  //   visible: true
  // }

];

export const ASSETS_FIELDS_FILTERS: UtmFieldType[] = [
  // {
  //   label: 'Severity',
  //   field: AssetFieldFilterEnum.SEVERITY,
  //   type: ElasticDataTypesEnum.STRING,
  //   visible: true
  // },
  /*// {
  //   label: 'Status',
  //   field: AssetFieldFilterEnum.STATUS,
  //   type: ElasticDataTypesEnum.STRING,
  //   visible: true
  // },*/
  {
     label: 'Status',
     field: AssetFieldFilterEnum.ALIVE,
     type: ElasticDataTypesEnum.STRING,
      visible: true
  },
  {
    label: 'Group',
    field: AssetFieldFilterEnum.GROUP,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  },
  // {
  //   label: 'Discovered by',
  //   field: AssetFieldFilterEnum.PROBE,
  //   type: ElasticDataTypesEnum.STRING,
  //   visible: true
  // },
  {
    label: 'Type',
    field: AssetFieldFilterEnum.TYPE,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  },
  {
    label: 'Data Types',
    field: AssetFieldFilterEnum.DATA_TYPES,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  },
  // {
  //   label: 'Alias',
  //   field: AssetFieldFilterEnum.ALIAS,
  //   type: ElasticDataTypesEnum.STRING,
  //   visible: true
  // },
  // {
  //   label: 'Ports',
  //   field: AssetFieldFilterEnum.PORTS,
  //   type: ElasticDataTypesEnum.STRING,
  //   visible: true
  // },
  // {
  //   label: 'Data Source',
  //   field: AssetFieldFilterEnum.OS,
  //   type: ElasticDataTypesEnum.STRING,
  //   visible: true
  // },
];

export const COLLECTORS_FIELDS_FILTERS: UtmFieldType[] = [
  {
    label: 'Group',
    field: CollectorFieldFilterEnum.COLLECTOR_GROUP,
    type: ElasticDataTypesEnum.STRING,
    visible: true
  },
];

