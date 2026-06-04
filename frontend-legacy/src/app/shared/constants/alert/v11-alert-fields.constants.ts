import { ElasticDataTypesEnum } from '../../enums/elastic-data-types.enum';
import {UtmFieldType} from '../../types/table/utm-field.type';
import {
  ALERT_CASE_ID_FIELD,
  ALERT_CATEGORY_FIELD,
  ALERT_FIELDS,
  ALERT_GENERATED_BY_FIELD,
  ALERT_IMPACT_AVAILABILITY_FIELD,
  ALERT_IMPACT_CONFIDENTIALITY_FIELD, ALERT_IMPACT_INTEGRITY_FIELD,
  ALERT_INCIDENT_NAME_FIELD,
  ALERT_NAME_FIELD,
  ALERT_PROTOCOL_FIELD,
  ALERT_SENSOR_FIELD,
  ALERT_SEVERITY_FIELD_LABEL,
  ALERT_TAGS_FIELD,
  ALERT_TIMESTAMP_FIELD
} from './alert-field.constant';

// TARGET
export const ALERT_TARGET_IP_FIELD = 'target.ip';
export const ALERT_TARGET_BYTES_SENT_FIELD = 'target.bytesSent';
export const ALERT_TARGET_BYTES_RECEIVED_FIELD = 'target.bytesReceived';
export const ALERT_TARGET_PACKAGES_SENT_FIELD = 'target.packagesSent';
export const ALERT_TARGET_PACKAGES_RECEIVED_FIELD = 'target.packagesReceived';
export const ALERT_TARGET_URL_FIELD = 'target.url';
export const ALERT_TARGET_DOMAIN_FIELD = 'target.domain';
export const ALERT_TARGET_PORT_FIELD = 'target.port';
export const ALERT_TARGET_CIDR_FIELD = 'target.cidr';
export const ALERT_TARGET_MAC_FIELD = 'target.mac';
export const ALERT_TARGET_HOST_FIELD = 'target.host';
export const ALERT_TARGET_USER_FIELD = 'target.user';
export const ALERT_TARGET_GROUP_FIELD = 'target.group';

// Geolocation
export const ALERT_TARGET_GEOLOCATION_COUNTRY_FIELD = 'target.geolocation.country';
export const ALERT_TARGET_GEOLOCATION_CITY_FIELD = 'target.geolocation.city';
export const ALERT_TARGET_GEOLOCATION_LATITUDE_FIELD = 'target.geolocation.latitude';
export const ALERT_TARGET_GEOLOCATION_LONGITUDE_FIELD = 'target.geolocation.longitude';
export const ALERT_TARGET_GEOLOCATION_ASN_FIELD = 'target.geolocation.asn';
export const ALERT_TARGET_GEOLOCATION_ASO_FIELD = 'target.geolocation.aso';
export const ALERT_TARGET_GEOLOCATION_COUNTRY_CODE_FIELD = 'target.geolocation.countryCode';
export const ALERT_TARGET_GEOLOCATION_ACCURACY_FIELD = 'target.geolocation.accuracy';

// Certificates & Fingerprints
export const ALERT_TARGET_CERTIFICATE_FINGERPRINT_FIELD = 'target.certificateFingerprint';
export const ALERT_TARGET_JA3_FINGERPRINT_FIELD = 'target.ja3Fingerprint';
export const ALERT_TARGET_JARM_FINGERPRINT_FIELD = 'target.jarmFingerprint';
export const ALERT_TARGET_SSH_BANNER_FIELD = 'target.sshBanner';
export const ALERT_TARGET_SSH_FINGERPRINT_FIELD = 'target.sshFingerprint';

// Web & Email
export const ALERT_TARGET_COOKIE_FIELD = 'target.cookie';
export const ALERT_TARGET_JABBER_ID_FIELD = 'target.jabberId';
export const ALERT_TARGET_EMAIL_FIELD = 'target.email';
export const ALERT_TARGET_DKIM_FIELD = 'target.dkim';
export const ALERT_TARGET_DKIM_SIGNATURE_FIELD = 'target.dkimSignature';
export const ALERT_TARGET_EMAIL_ADDRESS_FIELD = 'target.emailAddress';
export const ALERT_TARGET_EMAIL_BODY_FIELD = 'target.emailBody';
export const ALERT_TARGET_EMAIL_DISPLAY_NAME_FIELD = 'target.emailDisplayName';
export const ALERT_TARGET_EMAIL_SUBJECT_FIELD = 'target.emailSubject';
export const ALERT_TARGET_EMAIL_THREAD_INDEX_FIELD = 'target.emailThreadIndex';
export const ALERT_TARGET_EMAIL_XMAILER_FIELD = 'target.emailXMailer';

// WHOIS
export const ALERT_TARGET_WHOIS_REGISTRANT_FIELD = 'target.whoisRegistrant';
export const ALERT_TARGET_WHOIS_REGISTRAR_FIELD = 'target.whoisRegistrar';

// Process
export const ALERT_TARGET_PROCESS_FIELD = 'target.process';
export const ALERT_TARGET_PROCESS_STATE_FIELD = 'target.processState';
export const ALERT_TARGET_COMMAND_FIELD = 'target.command';
export const ALERT_TARGET_WINDOWS_SCHEDULED_TASK_FIELD = 'target.windowsScheduledTask';
export const ALERT_TARGET_WINDOWS_SERVICE_DISPLAY_NAME_FIELD = 'target.windowsServiceDisplayName';
export const ALERT_TARGET_WINDOWS_SERVICE_NAME_FIELD = 'target.windowsServiceName';

// File
export const ALERT_TARGET_FILE_FIELD = 'target.file';
export const ALERT_TARGET_PATH_FIELD = 'target.path';
export const ALERT_TARGET_FILENAME_FIELD = 'target.filename';
export const ALERT_TARGET_SIZE_IN_BYTES_FIELD = 'target.sizeInBytes';
export const ALERT_TARGET_MIME_TYPE_FIELD = 'target.mimeType';

// Hashes
export const ALERT_TARGET_HASH_FIELD = 'target.hash';
export const ALERT_TARGET_AUTHENTIHASH_FIELD = 'target.authentihash';
export const ALERT_TARGET_CDHASH_FIELD = 'target.cdhash';
export const ALERT_TARGET_MD5_FIELD = 'target.md5';
export const ALERT_TARGET_SHA1_FIELD = 'target.sha1';
export const ALERT_TARGET_SHA224_FIELD = 'target.sha224';
export const ALERT_TARGET_SHA256_FIELD = 'target.sha256';
export const ALERT_TARGET_SHA384_FIELD = 'target.sha384';
export const ALERT_TARGET_SHA3224_FIELD = 'target.sha3224';
export const ALERT_TARGET_SHA3256_FIELD = 'target.sha3256';
export const ALERT_TARGET_SHA3384_FIELD = 'target.sha3384';
export const ALERT_TARGET_SHA3512_FIELD = 'target.sha3512';
export const ALERT_TARGET_SHA512_FIELD = 'target.sha512';
export const ALERT_TARGET_SHA512224_FIELD = 'target.sha512224';
export const ALERT_TARGET_SHA512256_FIELD = 'target.sha512256';
export const ALERT_TARGET_HEX_FIELD = 'target.hex';
export const ALERT_TARGET_BASE64_FIELD = 'target.base64';

// System & Vulnerability
export const ALERT_TARGET_OPERATING_SYSTEM_FIELD = 'target.operatingSystem';
export const ALERT_TARGET_CHROME_EXTENSION_FIELD = 'target.chromeExtension';
export const ALERT_TARGET_MOBILE_APP_ID_FIELD = 'target.mobileAppId';
export const ALERT_TARGET_CPE_FIELD = 'target.cpe';
export const ALERT_TARGET_CVE_FIELD = 'target.cve';

// Malware
export const ALERT_TARGET_MALWARE_FIELD = 'target.malware';
export const ALERT_TARGET_MALWARE_FAMILY_FIELD = 'target.malwareFamily';
export const ALERT_TARGET_MALWARE_TYPE_FIELD = 'target.malwareType';

// Keys
export const ALERT_TARGET_PGP_PRIVATE_KEY_FIELD = 'target.pgpPrivateKey';
export const ALERT_TARGET_PGP_PUBLIC_KEY_FIELD = 'target.pgpPublicKey';

// Resources
export const ALERT_TARGET_CONNECTIONS_FIELD = 'target.connections';
export const ALERT_TARGET_USED_CPU_PERCENT_FIELD = 'target.usedCpuPercent';
export const ALERT_TARGET_USED_MEM_PERCENT_FIELD = 'target.usedMemPercent';
export const ALERT_TARGET_TOTAL_CPU_UNITS_FIELD = 'target.totalCpuUnits';
export const ALERT_TARGET_TOTAL_MEM_FIELD = 'target.totalMem';

// ADVERSARY
export const ALERT_ADVERSARY_IP_FIELD = 'adversary.ip';
export const ALERT_ADVERSARY_BYTES_SENT_FIELD = 'adversary.bytesSent';
export const ALERT_ADVERSARY_BYTES_RECEIVED_FIELD = 'adversary.bytesReceived';
export const ALERT_ADVERSARY_PACKAGES_SENT_FIELD = 'adversary.packagesSent';
export const ALERT_ADVERSARY_PACKAGES_RECEIVED_FIELD = 'adversary.packagesReceived';
export const ALERT_ADVERSARY_URL_FIELD = 'adversary.url';
export const ALERT_ADVERSARY_DOMAIN_FIELD = 'adversary.domain';
export const ALERT_ADVERSARY_PORT_FIELD = 'adversary.port';
export const ALERT_ADVERSARY_CIDR_FIELD = 'adversary.cidr';
export const ALERT_ADVERSARY_MAC_FIELD = 'adversary.mac';
export const ALERT_ADVERSARY_HOST_FIELD = 'adversary.host';
export const ALERT_ADVERSARY_USER_FIELD = 'adversary.user';
export const ALERT_ADVERSARY_GROUP_FIELD = 'adversary.group';

// Geolocation
export const ALERT_ADVERSARY_GEOLOCATION_COUNTRY_FIELD = 'adversary.geolocation.country';
export const ALERT_ADVERSARY_GEOLOCATION_CITY_FIELD = 'adversary.geolocation.city';
export const ALERT_ADVERSARY_GEOLOCATION_LATITUDE_FIELD = 'adversary.geolocation.latitude';
export const ALERT_ADVERSARY_GEOLOCATION_LONGITUDE_FIELD = 'adversary.geolocation.longitude';
export const ALERT_ADVERSARY_GEOLOCATION_ASN_FIELD = 'adversary.geolocation.asn';
export const ALERT_ADVERSARY_GEOLOCATION_ASO_FIELD = 'adversary.geolocation.aso';
export const ALERT_ADVERSARY_GEOLOCATION_COUNTRY_CODE_FIELD = 'adversary.geolocation.countryCode';
export const ALERT_ADVERSARY_GEOLOCATION_ACCURACY_FIELD = 'adversary.geolocation.accuracy';

// Certificates & Fingerprints
export const ALERT_ADVERSARY_CERTIFICATE_FINGERPRINT_FIELD = 'adversary.certificateFingerprint';
export const ALERT_ADVERSARY_JA3_FINGERPRINT_FIELD = 'adversary.ja3Fingerprint';
export const ALERT_ADVERSARY_JARM_FINGERPRINT_FIELD = 'adversary.jarmFingerprint';
export const ALERT_ADVERSARY_SSH_BANNER_FIELD = 'adversary.sshBanner';
export const ALERT_ADVERSARY_SSH_FINGERPRINT_FIELD = 'adversary.sshFingerprint';

// Web & Email
export const ALERT_ADVERSARY_COOKIE_FIELD = 'adversary.cookie';
export const ALERT_ADVERSARY_JABBER_ID_FIELD = 'adversary.jabberId';
export const ALERT_ADVERSARY_EMAIL_FIELD = 'adversary.email';
export const ALERT_ADVERSARY_DKIM_FIELD = 'adversary.dkim';
export const ALERT_ADVERSARY_DKIM_SIGNATURE_FIELD = 'adversary.dkimSignature';
export const ALERT_ADVERSARY_EMAIL_ADDRESS_FIELD = 'adversary.emailAddress';
export const ALERT_ADVERSARY_EMAIL_BODY_FIELD = 'adversary.emailBody';
export const ALERT_ADVERSARY_EMAIL_DISPLAY_NAME_FIELD = 'adversary.emailDisplayName';
export const ALERT_ADVERSARY_EMAIL_SUBJECT_FIELD = 'adversary.emailSubject';
export const ALERT_ADVERSARY_EMAIL_THREAD_INDEX_FIELD = 'adversary.emailThreadIndex';
export const ALERT_ADVERSARY_EMAIL_XMAILER_FIELD = 'adversary.emailXMailer';

// WHOIS
export const ALERT_ADVERSARY_WHOIS_REGISTRANT_FIELD = 'adversary.whoisRegistrant';
export const ALERT_ADVERSARY_WHOIS_REGISTRAR_FIELD = 'adversary.whoisRegistrar';

// Process
export const ALERT_ADVERSARY_PROCESS_FIELD = 'adversary.process';
export const ALERT_ADVERSARY_PROCESS_STATE_FIELD = 'adversary.processState';
export const ALERT_ADVERSARY_COMMAND_FIELD = 'adversary.command';
export const ALERT_ADVERSARY_WINDOWS_SCHEDULED_TASK_FIELD = 'adversary.windowsScheduledTask';
export const ALERT_ADVERSARY_WINDOWS_SERVICE_DISPLAY_NAME_FIELD = 'adversary.windowsServiceDisplayName';
export const ALERT_ADVERSARY_WINDOWS_SERVICE_NAME_FIELD = 'adversary.windowsServiceName';

// File
export const ALERT_ADVERSARY_FILE_FIELD = 'adversary.file';
export const ALERT_ADVERSARY_PATH_FIELD = 'adversary.path';
export const ALERT_ADVERSARY_FILENAME_FIELD = 'adversary.filename';
export const ALERT_ADVERSARY_SIZE_IN_BYTES_FIELD = 'adversary.sizeInBytes';
export const ALERT_ADVERSARY_MIME_TYPE_FIELD = 'adversary.mimeType';

// Hashes
export const ALERT_ADVERSARY_HASH_FIELD = 'adversary.hash';
export const ALERT_ADVERSARY_AUTHENTIHASH_FIELD = 'adversary.authentihash';
export const ALERT_ADVERSARY_CDHASH_FIELD = 'adversary.cdhash';
export const ALERT_ADVERSARY_MD5_FIELD = 'adversary.md5';
export const ALERT_ADVERSARY_SHA1_FIELD = 'adversary.sha1';
export const ALERT_ADVERSARY_SHA224_FIELD = 'adversary.sha224';
export const ALERT_ADVERSARY_SHA256_FIELD = 'adversary.sha256';
export const ALERT_ADVERSARY_SHA384_FIELD = 'adversary.sha384';
export const ALERT_ADVERSARY_SHA3224_FIELD = 'adversary.sha3224';
export const ALERT_ADVERSARY_SHA3256_FIELD = 'adversary.sha3256';
export const ALERT_ADVERSARY_SHA3384_FIELD = 'adversary.sha3384';
export const ALERT_ADVERSARY_SHA3512_FIELD = 'adversary.sha3512';
export const ALERT_ADVERSARY_SHA512_FIELD = 'adversary.sha512';
export const ALERT_ADVERSARY_SHA512224_FIELD = 'adversary.sha512224';
export const ALERT_ADVERSARY_SHA512256_FIELD = 'adversary.sha512256';
export const ALERT_ADVERSARY_HEX_FIELD = 'adversary.hex';
export const ALERT_ADVERSARY_BASE64_FIELD = 'adversary.base64';

// System & Vulnerability
export const ALERT_ADVERSARY_OPERATING_SYSTEM_FIELD = 'adversary.operatingSystem';
export const ALERT_ADVERSARY_CHROME_EXTENSION_FIELD = 'adversary.chromeExtension';
export const ALERT_ADVERSARY_MOBILE_APP_ID_FIELD = 'adversary.mobileAppId';
export const ALERT_ADVERSARY_CPE_FIELD = 'adversary.cpe';
export const ALERT_ADVERSARY_CVE_FIELD = 'adversary.cve';

// Malware
export const ALERT_ADVERSARY_MALWARE_FIELD = 'adversary.malware';
export const ALERT_ADVERSARY_MALWARE_FAMILY_FIELD = 'adversary.malwareFamily';
export const ALERT_ADVERSARY_MALWARE_TYPE_FIELD = 'adversary.malwareType';

// Keys
export const ALERT_ADVERSARY_PGP_PRIVATE_KEY_FIELD = 'adversary.pgpPrivateKey';
export const ALERT_ADVERSARY_PGP_PUBLIC_KEY_FIELD = 'adversary.pgpPublicKey';

// Resources
export const ALERT_ADVERSARY_CONNECTIONS_FIELD = 'adversary.connections';
export const ALERT_ADVERSARY_USED_CPU_PERCENT_FIELD = 'adversary.usedCpuPercent';
export const ALERT_ADVERSARY_USED_MEM_PERCENT_FIELD = 'adversary.usedMemPercent';
export const ALERT_ADVERSARY_TOTAL_CPU_UNITS_FIELD = 'adversary.totalCpuUnits';
export const ALERT_ADVERSARY_TOTAL_MEM_FIELD = 'adversary.totalMem';




export const V11_ALERT_FIELDS: UtmFieldType[] = [
  // Core alert fields
  { label: 'Alert Name', field: ALERT_NAME_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Alert ID', field: ALERT_CASE_ID_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Severity', field: ALERT_SEVERITY_FIELD_LABEL, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Protocol', field: ALERT_PROTOCOL_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Category', field: ALERT_CATEGORY_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Sensor', field: ALERT_SENSOR_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Generated By', field: ALERT_GENERATED_BY_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Tags', field: ALERT_TAGS_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Time', field: ALERT_TIMESTAMP_FIELD, type: ElasticDataTypesEnum.DATE, visible: false },
  { label: 'Incident Name', field: ALERT_INCIDENT_NAME_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Impact Availability', field: ALERT_IMPACT_AVAILABILITY_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Impact Confidentiality', field: ALERT_IMPACT_CONFIDENTIALITY_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Impact Integrity', field: ALERT_IMPACT_INTEGRITY_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },

  // Adversary fields
  { label: 'Adversary IP', field: ALERT_ADVERSARY_IP_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Host', field: ALERT_ADVERSARY_HOST_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary User', field: ALERT_ADVERSARY_USER_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Group', field: ALERT_ADVERSARY_GROUP_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Domain', field: ALERT_ADVERSARY_DOMAIN_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary MAC', field: ALERT_ADVERSARY_MAC_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Port', field: ALERT_ADVERSARY_PORT_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary URL', field: ALERT_ADVERSARY_URL_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary CIDR', field: ALERT_ADVERSARY_CIDR_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Bytes Sent', field: ALERT_ADVERSARY_BYTES_SENT_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Bytes Received', field: ALERT_ADVERSARY_BYTES_RECEIVED_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Packages Sent', field: ALERT_ADVERSARY_PACKAGES_SENT_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Packages Received', field: ALERT_ADVERSARY_PACKAGES_RECEIVED_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },

  // Adversary geolocation
  { label: 'Adversary Country', field: ALERT_ADVERSARY_GEOLOCATION_COUNTRY_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary City', field: ALERT_ADVERSARY_GEOLOCATION_CITY_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Latitude', field: ALERT_ADVERSARY_GEOLOCATION_LATITUDE_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Longitude', field: ALERT_ADVERSARY_GEOLOCATION_LONGITUDE_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary ASN', field: ALERT_ADVERSARY_GEOLOCATION_ASN_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary ASO', field: ALERT_ADVERSARY_GEOLOCATION_ASO_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Country Code', field: ALERT_ADVERSARY_GEOLOCATION_COUNTRY_CODE_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Adversary Geolocation Accuracy', field: ALERT_ADVERSARY_GEOLOCATION_ACCURACY_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },

  // Target fields
  { label: 'Target IP', field: ALERT_TARGET_IP_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Host', field: ALERT_TARGET_HOST_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target User', field: ALERT_TARGET_USER_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Group', field: ALERT_TARGET_GROUP_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Domain', field: ALERT_TARGET_DOMAIN_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target MAC', field: ALERT_TARGET_MAC_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Port', field: ALERT_TARGET_PORT_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target URL', field: ALERT_TARGET_URL_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target CIDR', field: ALERT_TARGET_CIDR_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Bytes Sent', field: ALERT_TARGET_BYTES_SENT_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Bytes Received', field: ALERT_TARGET_BYTES_RECEIVED_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Packages Sent', field: ALERT_TARGET_PACKAGES_SENT_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Packages Received', field: ALERT_TARGET_PACKAGES_RECEIVED_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },

  // Target geolocation
  { label: 'Target Country', field: ALERT_TARGET_GEOLOCATION_COUNTRY_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target City', field: ALERT_TARGET_GEOLOCATION_CITY_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Latitude', field: ALERT_TARGET_GEOLOCATION_LATITUDE_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Longitude', field: ALERT_TARGET_GEOLOCATION_LONGITUDE_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target ASN', field: ALERT_TARGET_GEOLOCATION_ASN_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target ASO', field: ALERT_TARGET_GEOLOCATION_ASO_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Country Code', field: ALERT_TARGET_GEOLOCATION_COUNTRY_CODE_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
  { label: 'Target Geolocation Accuracy', field: ALERT_TARGET_GEOLOCATION_ACCURACY_FIELD, type: ElasticDataTypesEnum.STRING, visible: false },
];

