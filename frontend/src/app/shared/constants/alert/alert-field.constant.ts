import {ElasticDataTypesEnum} from '../../enums/elastic-data-types.enum';
import {UtmFieldType} from '../../types/table/utm-field.type';

export const ALERT_NAME_FIELD = 'name';
export const ALERT_DESCRIPTION_FIELD = 'description';
export const ALERT_CATEGORY_DESCRIPTION_FIELD = 'category_description';
export const ALERT_SEVERITY_DESCRIPTION_FIELD = 'severity_description';
export const ALERT_FULL_LOG_FIELD = 'full_log';
export const ALERT_SOLUTION_FIELD = 'solution';
export const ALERT_STATUS_FIELD = 'status';
export const ALERT_STATUS_FIELD_AUTO = 'status';
export const ALERT_STATUS_LABEL_FIELD = 'statusLabel';
export const ALERT_SEVERITY_FIELD = 'severity';
export const ALERT_IMPACT_FIELD = 'impact';
export const ALERT_SEVERITY_FIELD_LABEL = 'severityLabel';
export const ALERT_TAGS_FIELD = 'tags';
export const ALERT_NOTE_FIELD = 'notes';
export const ALERT_OBSERVATION_FIELD = 'statusObservation';
export const ALERT_TIMESTAMP_FIELD = '@timestamp';
export const ALERT_ID_FIELD = 'id';
export const ALERT_CASE_ID_FIELD = 'id';
export const ALERT_GLOBAL_FIELD = 'alertType';
export const ALERT_TACTIC_FIELD = 'tactic';
export const ALERT_CATEGORY_FIELD = 'category';
export const ALERT_METHOD_FIELD = 'method';
export const ALERT_SENSOR_FIELD = 'dataSource';
export const ALERT_PROTOCOL_FIELD = 'protocol';
export const LOG_RELATED_ID_EVENT_FIELD = 'logs';
export const ALERT_REFERENCE_FIELD = 'reference';
export const ALERT_RELATED_RULES_FIELD = 'TagRulesApplied';
export const ALERT_TARGET_FIELD = 'target';
export const ALERT_ADVERSARY_FIELD = 'adversary';
export const ALERT_TECHNIQUE_FIELD = 'technique';
export const ALERT_PARENT_ID = 'parentId';

// SOURCE
export const ALERT_SOURCE_HOSTNAME_FIELD = 'source.host';
export const ALERT_SOURCE_IP_FIELD = 'source.ip';
export const ALERT_SOURCE_PORT_FIELD = 'source.port';
export const ALERT_SOURCE_COUNTRY_FIELD = 'source.country';
export const ALERT_SOURCE_COUNTRY_CODE_FIELD = 'source.countryCode';
export const ALERT_SOURCE_COUNTRY_COORDINATES_FIELD = 'source.coordinates';
export const ALERT_SOURCE_CITY_FIELD = 'source.city';
export const ALERT_SOURCE_ASN_FIELD = 'source.asn';
export const ALERT_SOURCE_ASO_FIELD = 'source.aso';
export const ALERT_SOURCE_IS_SATELLITE_FIELD = 'source.isSatelliteProvider';
export const ALERT_SOURCE_IS_AN_PROXY_FIELD = 'source.isAnonymousProxy';
export const ALERT_SOURCE_ACCURACY_FIELD = 'source.accuracyRadius';
export const ALERT_SOURCE_USER_FIELD = 'source.user';

// DESTINATION
export const ALERT_DESTINATION_HOSTNAME_FIELD = 'destination.host';
export const ALERT_DESTINATION_IP_FIELD = 'destination.ip';
export const ALERT_DESTINATION_PORT_FIELD = 'destination.port';
export const ALERT_DESTINATION_COUNTRY_FIELD = 'destination.country';
export const ALERT_DESTINATION_COUNTRY_CODE_FIELD = 'destination.countryCode';
export const ALERT_DESTINATION_COUNTRY_COORDINATES_FIELD = 'destination.coordinates';
export const ALERT_DESTINATION_CITY_FIELD = 'destination.city';
export const ALERT_DESTINATION_ASN_FIELD = 'destination.asn';
export const ALERT_DESTINATION_ASO_FIELD = 'destination.aso';
export const ALERT_DESTINATION_IS_SATELLITE_FIELD = 'destination.isSatelliteProvider';
export const ALERT_DESTINATION_IS_AN_PROXY_FIELD = 'destination.isAnonymousProxy';
export const ALERT_DESTINATION_ACCURACY_FIELD = 'destination.accuracyRadius';
export const ALERT_DESTINATION_USER_FIELD = 'destination.user';

// IMPACT
export const ALERT_IMPACT_AVAILABILITY_FIELD = 'impact.availability';
export const ALERT_IMPACT_CONFIDENTIALITY_FIELD = 'impact.confidentiality';
export const ALERT_IMPACT_INTEGRITY_FIELD = 'impact.integrity';


// TARGET
export const ALERT_TARGET_IP_FIELD = 'target.ip';
export const ALERT_TARGET_FILE_FIELD = 'target.file';
export const ALERT_TARGET_HOST_FIELD = 'target.host';
export const ALERT_TARGET_USER_FIELD = 'target.user';
export const ALERT_TARGET_BYTES_SENT_FIELD = 'target.bytesSent';
export const ALERT_TARGET_URL_FIELD = 'target.url';
export const ALERT_TARGET_DOMAIN_FIELD = 'target.domain';
export const ALERT_TARGET_GEOLOCATION_ASN_FIELD = 'target.geolocation.asn';
export const ALERT_TARGET_GEOLOCATION_ASO_FIELD = 'target.geolocation.aso';
export const ALERT_TARGET_GEOLOCATION_CITY_FIELD = 'target.geolocation.city';
export const ALERT_TARGET_GEOLOCATION_COUNTRY_FIELD = 'target.geolocation.country';
export const ALERT_TARGET_GEOLOCATION_COUNTRY_CODE_FIELD = 'target.geolocation.countryCode';
export const ALERT_TARGET_GEOLOCATION_LATITUDE_FIELD = 'target.geolocation.latitude';
export const ALERT_TARGET_GEOLOCATION_LONGITUDE_FIELD = 'target.geolocation.longitude';

// ADVERSARY
export const ALERT_ADVERSARY_IP_FIELD = 'adversary.ip';
export const ALERT_ADVERSARY_FILE_FIELD = 'adversary.file';
export const ALERT_ADVERSARY_HOST_FIELD = 'adversary.host';
export const ALERT_ADVERSARY_USER_FIELD = 'adversary.user';
export const ALERT_ADVERSARY_BYTES_SENT_FIELD = 'adversary.bytesSent';
export const ALERT_ADVERSARY_URL_FIELD = 'adversary.url';
export const ALERT_ADVERSARY_DOMAIN_FIELD = 'adversary.domain';
export const ALERT_ADVERSARY_GEOLOCATION_ASN_FIELD = 'adversary.geolocation.asn';
export const ALERT_ADVERSARY_GEOLOCATION_ASO_FIELD = 'adversary.geolocation.aso';
export const ALERT_ADVERSARY_GEOLOCATION_CITY_FIELD = 'adversary.geolocation.city';
export const ALERT_ADVERSARY_GEOLOCATION_COUNTRY_FIELD = 'adversary.geolocation.country';
export const ALERT_ADVERSARY_GEOLOCATION_COUNTRY_CODE_FIELD = 'adversary.geolocation.countryCode';
export const ALERT_ADVERSARY_GEOLOCATION_LATITUDE_FIELD = 'adversary.geolocation.latitude';
export const ALERT_ADVERSARY_GEOLOCATION_LONGITUDE_FIELD = 'adversary.geolocation.longitude';


export const ALERT_GENERATED_BY_FIELD = 'dataType';

// INCIDENT AND ALERT
export const ALERT_INCIDENT_FLAG_FIELD = 'isIncident';
export const ALERT_INCIDENT_DATE_FIELD = 'incidentDetail.creationDate';
export const ALERT_INCIDENT_USER_FIELD = 'incidentDetail.createdBy';
export const ALERT_INCIDENT_OBSERVATION_FIELD = 'incidentDetail.incidentObservation';
export const ALERT_INCIDENT_NAME_FIELD = 'incidentDetail.incidentName';
export const ALERT_INCIDENT_ID_FIELD = 'incidentDetail.incidentId';
export const ALERT_INCIDENT_MODULE_FIELD = 'incidentDetail.incidentSource';
export const EVENT_IS_ALERT = 'isAlert';

export const FALSE_POSITIVE_OBJECT = {id: 1, tagName: 'False positive', tagColor: '#f44336', systemOwner: true};

export const ALERT_FIELDS: UtmFieldType[] = [
  {
    label: 'Alert name',
    field: ALERT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Severity',
    field: ALERT_SEVERITY_FIELD_LABEL,
    type: ElasticDataTypesEnum.NUMBER,
    visible: true,
  },
  {
    label: 'Status',
    field: ALERT_STATUS_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: true,
  },
  {
    label: 'Time',
    field: ALERT_TIMESTAMP_FIELD,
    type: ElasticDataTypesEnum.DATE,
    visible: true,
  },
  {
    label: 'Sensor',
    field: ALERT_SENSOR_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Target',
    field: ALERT_TARGET_FIELD,
    type: ElasticDataTypesEnum.OBJECT,
    visible: true,
    fields: [
      {
        label: 'Target IP',
        field: ALERT_TARGET_IP_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: true,
      },
      {
        label: 'Target URL',
        field: ALERT_TARGET_URL_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Target Bytes Sent',
        field: ALERT_TARGET_BYTES_SENT_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Target Domain',
        field: ALERT_TARGET_DOMAIN_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Target ASN',
        field: ALERT_TARGET_GEOLOCATION_ASN_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: true,
      },
      {
        label: 'Target ASO',
        field: ALERT_TARGET_GEOLOCATION_ASO_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: true,
      },
      {
        label: 'Target Latitude',
        field: ALERT_TARGET_GEOLOCATION_LATITUDE_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Target Longitude',
        field: ALERT_TARGET_GEOLOCATION_LONGITUDE_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Target File',
        field: ALERT_TARGET_FILE_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Target Host',
        field: ALERT_TARGET_HOST_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Target User',
        field: ALERT_TARGET_USER_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      }
    ]
  },
  {
    label: 'Adversary',
    field: ALERT_ADVERSARY_FIELD,
    type: ElasticDataTypesEnum.OBJECT,
    visible: true,
    fields: [
      {
        label: 'Adversary IP',
        field: ALERT_ADVERSARY_IP_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: true,
      },
      {
        label: 'Adversary URL',
        field: ALERT_ADVERSARY_URL_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Adversary Bytes Sent',
        field: ALERT_ADVERSARY_BYTES_SENT_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Adversary Domain',
        field: ALERT_ADVERSARY_DOMAIN_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Adversary ASN',
        field: ALERT_ADVERSARY_GEOLOCATION_ASN_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: true,
      },
      {
        label: 'Adversary ASO',
        field: ALERT_ADVERSARY_GEOLOCATION_ASO_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: true,
      },
      {
        label: 'Adversary Latitude',
        field: ALERT_ADVERSARY_GEOLOCATION_LATITUDE_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Adversary Longitude',
        field: ALERT_ADVERSARY_GEOLOCATION_LONGITUDE_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Adversary File',
        field: ALERT_ADVERSARY_FILE_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Adversary Host',
        field: ALERT_ADVERSARY_HOST_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Adversary User',
        field: ALERT_ADVERSARY_URL_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      }
    ]
  },
  {
    label: 'Impact',
    field: ALERT_IMPACT_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
    fields: [
      {
        label: 'Availability',
        field: ALERT_IMPACT_AVAILABILITY_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: true,
      },
      {
        label: 'Confidentiality',
        field: ALERT_IMPACT_CONFIDENTIALITY_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
      {
        label: 'Integrity',
        field: ALERT_IMPACT_INTEGRITY_FIELD,
        type: ElasticDataTypesEnum.STRING,
        visible: false,
      },
    ]
  },
  {
    label: 'ID',
    field: ALERT_CASE_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  /*{
    label: 'Tactic',
    field: ALERT_TACTIC_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },*/
  {
    label: 'Technique',
    field: ALERT_TECHNIQUE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Note',
    field: ALERT_NOTE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Protocol',
    field: ALERT_PROTOCOL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Generated by',
    field: ALERT_GENERATED_BY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Category',
    field: ALERT_CATEGORY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Tags',
    field: ALERT_TAGS_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Observation',
    field: ALERT_OBSERVATION_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Incident ID',
    field: ALERT_INCIDENT_ID_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Incident Name',
    field: ALERT_INCIDENT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
];


export const ALERT_FILTERS_FIELDS: UtmFieldType[] = [
  {
    label: 'Alert Name',
    field: ALERT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    customStyle: 'text-blue-800',
    visible: true,
  },
  {
    label: 'ID',
    field: ALERT_CASE_ID_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Severity',
    field: ALERT_SEVERITY_FIELD_LABEL,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Protocol',
    field: ALERT_PROTOCOL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Generated by',
    field: ALERT_GENERATED_BY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Availability',
    field: ALERT_IMPACT_AVAILABILITY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Confidentiality',
    field: ALERT_IMPACT_CONFIDENTIALITY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Integrity',
    field: ALERT_IMPACT_INTEGRITY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Adversary IP',
    field: ALERT_ADVERSARY_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Adversary URL',
    field: ALERT_ADVERSARY_URL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Adversary Bytes Sent',
    field: ALERT_ADVERSARY_BYTES_SENT_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Adversary Domain',
    field: ALERT_ADVERSARY_DOMAIN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Adversary ASN',
    field: ALERT_ADVERSARY_GEOLOCATION_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Adversary ASO',
    field: ALERT_ADVERSARY_GEOLOCATION_ASO_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Adversary Latitude',
    field: ALERT_ADVERSARY_GEOLOCATION_LATITUDE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Adversary Longitude',
    field: ALERT_ADVERSARY_GEOLOCATION_LONGITUDE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Target IP',
    field: ALERT_TARGET_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Target URL',
    field: ALERT_TARGET_URL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Target Bytes Sent',
    field: ALERT_TARGET_BYTES_SENT_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Target Domain',
    field: ALERT_TARGET_DOMAIN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Target ASN',
    field: ALERT_TARGET_GEOLOCATION_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Target ASO',
    field: ALERT_TARGET_GEOLOCATION_ASO_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Target Latitude',
    field: ALERT_TARGET_GEOLOCATION_LATITUDE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Target Longitude',
    field: ALERT_TARGET_GEOLOCATION_LONGITUDE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Category',
    field: ALERT_CATEGORY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Sensor',
    field: ALERT_SENSOR_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Time',
    field: ALERT_TIMESTAMP_FIELD,
    type: ElasticDataTypesEnum.DATE,
    visible: false,
  },
  {
    label: 'Incident Name',
    field: ALERT_INCIDENT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Tags',
    field: ALERT_TAGS_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
];

export const EVENT_FIELDS: UtmFieldType[] = [
  {
    label: 'Alert name',
    field: ALERT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Severity',
    field: ALERT_SEVERITY_FIELD_LABEL,
    type: ElasticDataTypesEnum.NUMBER,
    visible: true,
  },
  {
    label: 'Time',
    field: ALERT_TIMESTAMP_FIELD,
    type: ElasticDataTypesEnum.DATE,
    visible: true,
  },
  {
    label: 'Sensor',
    field: ALERT_SENSOR_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Source IP',
    field: ALERT_SOURCE_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Destination IP',
    field: ALERT_DESTINATION_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'ID',
    field: ALERT_CASE_ID_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Note',
    field: ALERT_NOTE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Protocol',
    field: ALERT_PROTOCOL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Generated by',
    field: ALERT_GENERATED_BY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source hostname',
    field: ALERT_SOURCE_HOSTNAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source port',
    field: ALERT_SOURCE_PORT_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Source country',
    field: ALERT_SOURCE_COUNTRY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source city',
    field: ALERT_SOURCE_CITY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source ASN',
    field: ALERT_SOURCE_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination hostname',
    field: ALERT_DESTINATION_HOSTNAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination port',
    field: ALERT_DESTINATION_PORT_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Destination country',
    field: ALERT_DESTINATION_COUNTRY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination city',
    field: ALERT_DESTINATION_CITY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination ASN',
    field: ALERT_DESTINATION_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Category',
    field: ALERT_CATEGORY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Tags',
    field: ALERT_TAGS_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Observation',
    field: ALERT_OBSERVATION_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
];
export const EVENT_FILTERS_FIELDS: UtmFieldType[] = [
  {
    label: 'Alert name',
    field: ALERT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    customStyle: 'text-blue-800',
    visible: true,
  },
  {
    label: 'ID',
    field: ALERT_CASE_ID_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Severity',
    field: ALERT_SEVERITY_FIELD_LABEL,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Protocol',
    field: ALERT_PROTOCOL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Generated by',
    field: ALERT_GENERATED_BY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source hostname',
    field: ALERT_SOURCE_HOSTNAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source IP',
    field: ALERT_SOURCE_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source port',
    field: ALERT_SOURCE_PORT_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Source country',
    field: ALERT_SOURCE_COUNTRY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source ASN',
    field: ALERT_SOURCE_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination hostname',
    field: ALERT_DESTINATION_HOSTNAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination IP',
    field: ALERT_DESTINATION_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination port',
    field: ALERT_DESTINATION_PORT_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Destination country',
    field: ALERT_DESTINATION_COUNTRY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination ASN',
    field: ALERT_DESTINATION_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Category',
    field: ALERT_CATEGORY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Sensor',
    field: ALERT_SENSOR_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Time',
    field: ALERT_TIMESTAMP_FIELD,
    type: ElasticDataTypesEnum.DATE,
    visible: false,
  },
  {
    label: 'Tags',
    field: ALERT_TAGS_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  }
];

// Name (Nombre de la alerta o el evento, hoy rule)
// Severity
// Status
// Source Evidence (Alert o Event Security)
// Generated by (HIDS, NIDS,ETC)
// Created (Fecha en la que se agregó la alerta o evento como un incidente)
// Created by (Usuario que agregó la alerta o evento como un incidente)
// Observation (La observación que hizo el usuario cuando agregó como incidente)

export const INCIDENT_FIELDS: UtmFieldType[] = [
  {
    label: 'Alert name',
    field: ALERT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Severity',
    field: ALERT_SEVERITY_FIELD_LABEL,
    type: ElasticDataTypesEnum.NUMBER,
    visible: true,
  },
  {
    label: 'Status',
    field: ALERT_STATUS_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: true,
  },
  {
    label: 'Source evidence',
    field: ALERT_INCIDENT_MODULE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Time',
    field: ALERT_TIMESTAMP_FIELD,
    type: ElasticDataTypesEnum.DATE,
    visible: true,
  },
  {
    label: 'Sensor',
    field: ALERT_SENSOR_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source IP',
    field: ALERT_SOURCE_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination IP',
    field: ALERT_DESTINATION_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'ID',
    field: ALERT_CASE_ID_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Note',
    field: ALERT_NOTE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Protocol',
    field: ALERT_PROTOCOL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Generated by',
    field: ALERT_GENERATED_BY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Source hostname',
    field: ALERT_SOURCE_HOSTNAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source port',
    field: ALERT_SOURCE_PORT_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Source country',
    field: ALERT_SOURCE_COUNTRY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source city',
    field: ALERT_SOURCE_CITY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source ASN',
    field: ALERT_SOURCE_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination hostname',
    field: ALERT_DESTINATION_HOSTNAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination port',
    field: ALERT_DESTINATION_PORT_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Destination country',
    field: ALERT_DESTINATION_COUNTRY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination city',
    field: ALERT_DESTINATION_CITY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination ASN',
    field: ALERT_DESTINATION_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Category',
    field: ALERT_CATEGORY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Tags',
    field: ALERT_TAGS_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Observation',
    field: ALERT_OBSERVATION_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Created on',
    field: ALERT_INCIDENT_DATE_FIELD,
    type: ElasticDataTypesEnum.DATE,
    visible: true,
  },
  {
    label: 'Created by',
    field: ALERT_INCIDENT_USER_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Observation',
    field: ALERT_INCIDENT_OBSERVATION_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
];
export const INCIDENT_FILTERS_FIELDS: UtmFieldType[] = [
  {
    label: 'Alert name',
    field: ALERT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    customStyle: 'text-blue-800',
    visible: true,
  },
  {
    label: 'ID',
    field: ALERT_CASE_ID_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Severity',
    field: ALERT_SEVERITY_FIELD_LABEL,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Protocol',
    field: ALERT_PROTOCOL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source evidence',
    field: ALERT_INCIDENT_MODULE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Generated by',
    field: ALERT_GENERATED_BY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Created by',
    field: ALERT_INCIDENT_USER_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Source hostname',
    field: ALERT_SOURCE_HOSTNAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Source IP',
    field: ALERT_SOURCE_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source port',
    field: ALERT_SOURCE_PORT_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Source country',
    field: ALERT_SOURCE_COUNTRY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source ASN',
    field: ALERT_SOURCE_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination hostname',
    field: ALERT_DESTINATION_HOSTNAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination IP',
    field: ALERT_DESTINATION_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination port',
    field: ALERT_DESTINATION_PORT_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Destination country',
    field: ALERT_DESTINATION_COUNTRY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination ASN',
    field: ALERT_DESTINATION_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Category',
    field: ALERT_CATEGORY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Sensor',
    field: ALERT_SENSOR_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Time',
    field: ALERT_TIMESTAMP_FIELD,
    type: ElasticDataTypesEnum.DATE,
    visible: false,
  },
  {
    label: 'Tags',
    field: ALERT_TAGS_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  }
];


export const INCIDENT_AUTOMATION_ALERT_FIELDS: UtmFieldType[] = [
  {
    label: 'Alert name',
    field: ALERT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    customStyle: 'text-blue-800',
    visible: true,
  },
  {
    label: 'Severity',
    field: ALERT_SEVERITY_FIELD_LABEL,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Status',
    field: ALERT_STATUS_LABEL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Category',
    field: ALERT_CATEGORY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Sensor',
    field: ALERT_SENSOR_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Protocol',
    field: ALERT_PROTOCOL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source evidence',
    field: ALERT_INCIDENT_MODULE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Generated by',
    field: ALERT_GENERATED_BY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Created by',
    field: ALERT_INCIDENT_USER_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Source hostname',
    field: ALERT_SOURCE_HOSTNAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Source IP',
    field: ALERT_SOURCE_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source port',
    field: ALERT_SOURCE_PORT_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Source user',
    field: ALERT_SOURCE_USER_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source country',
    field: ALERT_SOURCE_COUNTRY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Source ASN',
    field: ALERT_SOURCE_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination hostname',
    field: ALERT_DESTINATION_HOSTNAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination IP',
    field: ALERT_DESTINATION_IP_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination port',
    field: ALERT_DESTINATION_PORT_FIELD,
    type: ElasticDataTypesEnum.NUMBER,
    visible: false,
  },
  {
    label: 'Destination user',
    field: ALERT_DESTINATION_USER_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination country',
    field: ALERT_DESTINATION_COUNTRY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Destination ASN',
    field: ALERT_DESTINATION_ASN_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },

];
