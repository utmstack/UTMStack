import {
  ALERT_CASE_ID_FIELD,
  ALERT_DESTINATION_ACCURACY_FIELD,
  ALERT_DESTINATION_COUNTRY_COORDINATES_FIELD,
  ALERT_DESTINATION_IP_FIELD,
  ALERT_FILTERS_FIELDS,
  ALERT_ID_FIELD,
  ALERT_NAME_FIELD,
  ALERT_SOURCE_ACCURACY_FIELD,
  ALERT_SOURCE_COUNTRY_COORDINATES_FIELD,
  ALERT_SOURCE_IP_FIELD,
  ALERT_STATUS_FIELD,
  ALERT_STATUS_LABEL_FIELD,
  EVENT_FILTERS_FIELDS,
  INCIDENT_FILTERS_FIELDS
} from '../../../../shared/constants/alert/alert-field.constant';
import {
  ALL_STATUS,
  AUTOMATIC_REVIEW,
  CLOSED,
  CLOSED_ICON,
  CLOSED_LABEL,
  IGNORED,
  IGNORED_ICON,
  IGNORED_LABEL,
  OPEN,
  OPEN_ICON,
  OPEN_LABEL,
  REVIEW,
  REVIEW_ICON,
  REVIEW_LABEL
} from '../../../../shared/constants/alert/alert-status.constant';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {AlertLocationType} from '../../../../shared/types/alert/alert-location.type';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';
import {getValueFromPropertyPath} from '../../../../shared/util/get-value-object-from-property-path.util';
import {EventDataTypeEnum} from '../enums/event-data-type.enum';

/**
 * Return ID of alert
 * @param alert Alert
 */
export function getID(alert: object): string {
  return getValueFromPropertyPath(alert, ALERT_ID_FIELD, null);
}

/**
 * Return Case ID of alert
 * @param alert Alert
 */
export function getCaseID(alert: object): string {
  return getValueFromPropertyPath(alert, ALERT_CASE_ID_FIELD, null);
}

export function getStatusName(status): string {
  let msg = '';
  switch (status) {
    case AUTOMATIC_REVIEW:
      msg = 'alertStatus.inReview';
      break;
    case REVIEW:
      msg = 'alertStatus.inReview';
      break;
    case IGNORED:
      msg = 'alertStatus.ignored';
      break;
    case CLOSED:
      msg = 'alertStatus.closed';
      break;
    case OPEN:
      msg = 'alertStatus.open';
      break;
  }
  return msg;
}

/**
 * return Label of field
 * @param filter Elasticsearch filter
 * @param dataType DataType to search label
 */
export function resolveFieldNameByFilter(filter: ElasticFilterType, dataType?: EventDataTypeEnum): string {
  const field = filter.field.replace('.keyword', '');
  let indexField: number;
  switch (dataType) {
    case EventDataTypeEnum.ALERT:
      indexField = ALERT_FILTERS_FIELDS.findIndex(value => value.field === field);
      if (indexField !== -1) {
        return ALERT_FILTERS_FIELDS[indexField].label;
      } else {
        return field;
      }
    case EventDataTypeEnum.FALSE_POSITIVE:
      indexField = ALERT_FILTERS_FIELDS.findIndex(value => value.field === field);
      if (indexField !== -1) {
        return ALERT_FILTERS_FIELDS[indexField].label;
      } else {
        return field;
      }
    case EventDataTypeEnum.INCIDENT:
      indexField = INCIDENT_FILTERS_FIELDS.findIndex(value => value.field === field);
      if (indexField !== -1) {
        return INCIDENT_FILTERS_FIELDS[indexField].label;
      } else {
        return field;
      }
    case EventDataTypeEnum.EVENT:
      indexField = EVENT_FILTERS_FIELDS.findIndex(value => value.field === field);
      if (indexField !== -1) {
        return EVENT_FILTERS_FIELDS[indexField].label;
      } else {
        return field;
      }
  }

}


/**
 * Return locations from alert
 * @param alert Alert to extract locations, return a Promise of AlertLocationType[]
 */
export function getLocationFromAlert(alert: object): Promise<AlertLocationType[]> {
  return new Promise<AlertLocationType[]>(resolve => {
    const locations: AlertLocationType[] = [];
    const sourceLocation: AlertLocationType = {
      location: getValueFromPropertyPath(alert, ALERT_SOURCE_COUNTRY_COORDINATES_FIELD, null),
      ip: getValueFromPropertyPath(alert, ALERT_SOURCE_IP_FIELD, null),
      alertName: getValueFromPropertyPath(alert, ALERT_NAME_FIELD, null),
      locationType: 'source',
      accuracy: getValueFromPropertyPath(alert, ALERT_SOURCE_ACCURACY_FIELD, null),
    };
    if (sourceLocation.location) {
      locations.push(sourceLocation);
    }
    const destinationLocation: AlertLocationType = {
      location: getValueFromPropertyPath(alert, ALERT_DESTINATION_COUNTRY_COORDINATES_FIELD, null),
      ip: getValueFromPropertyPath(alert, ALERT_DESTINATION_IP_FIELD, null),
      alertName: getValueFromPropertyPath(alert, ALERT_NAME_FIELD, null),
      locationType: 'destination',
      accuracy: getValueFromPropertyPath(alert, ALERT_DESTINATION_ACCURACY_FIELD, null),
    };
    if (destinationLocation.location) {
      locations.push(destinationLocation);
    }
    resolve(locations);
  });
}

/**
 * Resolve current status alert filter, always ALERT_STATUS_FIELD has priority hover ALERT_STATUS_LABEL_FIELD,
 * If in current filter exist both fields as index, get ALERT_STATUS_FIELD value as status
 * @param filters ElasticFilterType array with current filters
 */
export function getCurrentAlertStatus(filters: ElasticFilterType[]): number {
  // get index of ALERT_STATUS_FIELD, this is a numeric representation of alert status, find field and operator is is
  const indexStatus = filters.findIndex(value => value.field === ALERT_STATUS_FIELD
    && value.operator === ElasticOperatorsEnum.IS);
  // get index of ALERT_STATUS_LABEL_FIELD, this is a string representation of alert status
  const indexStatusLabel = filters.findIndex(value => value.field.includes(ALERT_STATUS_LABEL_FIELD)
    && value.operator === ElasticOperatorsEnum.IS);
  if (indexStatus !== -1) {
    /**
     * ALERT_STATUS_FIELD has priority so:
     * -check if there are any filter with ALERT_STATUS_FIELD
     * - if exist check if operator is IS_NOT,why? because when All filter is set filter ALERT_STATUS_FIELD
     *   is change to operator IS_NOT and value = 1, to avoid get alert in AUTOMATIC REVIEW
     *     - if operator is IS_NOT, then return ALL_STATUS
     *     - else convert convert current value of ALERT_STATUS_FIELD and return
     */
    if (filters[indexStatus].operator === ElasticOperatorsEnum.IS_NOT) {
      return ALL_STATUS;
    } else {
      return Number(filters[indexStatus].value);
    }
  } else {
    /**
     * If ALERT_STATUS_FIELD does not exist check if there any filter with ALERT_STATUS_LABEL_FIELD
     * - If not exist return all
     * - if exits map current string representation of status to numeric and return
     */
    if (indexStatusLabel === -1) {
      return ALL_STATUS;
    } else {
      switch (filters[indexStatusLabel].value) {
        case OPEN_LABEL:
          return OPEN;
        case REVIEW_LABEL:
          return REVIEW;
        case IGNORED_LABEL:
          return IGNORED;
        case CLOSED_LABEL:
          return CLOSED;
        default:
          return ALL_STATUS;
      }
    }
  }
}

export function setAlertPropertyValue(field: string, value: any, alert: object) {
  return alert[field] = value;
}

export function resolveStatusStyle(status: number): { icon: string, background: string, label: string } {
  switch (status) {
    case OPEN:
      return {icon: OPEN_ICON, background: 'border-success-400 text-success-400', label: 'alertStatus.open'};
    case REVIEW:
      return {icon: REVIEW_ICON, background: 'border-info-400 text-info-400', label: 'alertStatus.inReview'};
    case CLOSED:
      return {icon: CLOSED_ICON, background: 'border-blue-800 text-blue-800', label: 'alertStatus.closed'};
    case IGNORED:
      return {icon: IGNORED_ICON, background: 'border-warning-400 text-warning-400', label: 'alertStatus.ignored'};
    default:
      return {icon: 'icon-hammer', background: 'border-slate-800 text-slate-800', label: 'alertStatus.pending'};
  }
}
