package services

import (
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

// logIndexMappings pins the canonical fields for v11-log-* documents.
// "event" is OpenSearch's flat_object type: it stores every sub-key as a keyword
// and never infers types, so heterogeneous vendor values can no longer produce
// mapper_parsing_exception conflicts. "controls" holds compliance control tags.
// Top-level Event fields keep real types so the UI/SQL/sort work on them.
//
// This mappings template is scoped to v11-log-* ONLY (NOT v11-alert-*): alert
// documents carry e.g. severity as an integer, so they must not inherit these
// log-side text mappings. It carries no settings — settings come from the
// settings-only template (utmstack_log_indexes) that v11 applies to both.
const logIndexMappings = `
{
  "@timestamp": {"type":"date"},
  "dataType": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "dataSource": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "action": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "protocol": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "actionResult": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "severity": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "connectionStatus": {"type":"keyword"},
  "statusCode": {"type":"long"},
  "origin": {"type":"object","dynamic":true,"properties":{
    "ip": {"type":"ip","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "port": {"type":"long"},
    "user": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "host": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "file": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "bytesSent": {"type":"double"},
    "bytesReceived": {"type":"double"}
  }},
  "target": {"type":"object","dynamic":true,"properties":{
    "ip": {"type":"ip","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "port": {"type":"long"},
    "user": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "host": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "path": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "file": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "bytesSent": {"type":"double"},
    "bytesReceived": {"type":"double"}
  }},
  "event": {"type":"flat_object"},
  "controls": {"type":"keyword"},
  "utm": {"type": "object", "dynamic": false, "properties": {"ActorContextId": {"type": "keyword"}, "AppPoolName": {"type": "keyword"}, "CorrelationId": {"type": "keyword"}, "DetectionMethod": {"type": "keyword"}, "EventData": {"type": "keyword"}, "ExternalAccess": {"type": "keyword"}, "ItemType": {"type": "keyword"}, "LogonError": {"type": "keyword"}, "ObjectId": {"type": "keyword"}, "OrganizationId": {"type": "keyword"}, "OrganizationName": {"type": "keyword"}, "OriginatingServer": {"type": "keyword"}, "P2Sender": {"type": "keyword"}, "Platform": {"type": "keyword"}, "Recipients": {"type": "keyword"}, "RequestId": {"type": "keyword"}, "SourceFileExtension": {"type": "keyword"}, "Subject": {"type": "keyword"}, "TargetUserOrGroupName": {"type": "keyword"}, "TargetUserOrGroupType": {"type": "keyword"}, "UserType": {"type": "keyword"}, "Verdict": {"type": "keyword"}, "Workload": {"type": "keyword"}, "accessType": {"type": "keyword"}, "activityId": {"type": "keyword"}, "additionalEventData": {"type": "object", "dynamic": false, "properties": {"AuthenticationMethod": {"type": "keyword"}, "CipherSuite": {"type": "keyword"}}}, "agentVersion": {"type": "keyword"}, "appAccessContextClientAppId": {"type": "keyword"}, "awsRegion": {"type": "keyword"}, "beatVersion": {"type": "keyword"}, "category": {"type": "keyword"}, "ciscoMsg": {"type": "keyword"}, "deviceIps": {"type": "keyword"}, "eventCategory": {"type": "keyword"}, "eventCode": {"type": "keyword"}, "eventDataAccessList": {"type": "keyword"}, "eventDataAccessMask": {"type": "keyword"}, "eventDataAlgorithmName": {"type": "keyword"}, "eventDataApplication": {"type": "keyword"}, "eventDataBootAppStatus": {"type": "keyword"}, "eventDataCA": {"type": "keyword"}, "eventDataCallerProcessId": {"type": "keyword"}, "eventDataCallerProcessName": {"type": "keyword"}, "eventDataCategoryId": {"type": "keyword"}, "eventDataClassId": {"type": "keyword"}, "eventDataClassName": {"type": "keyword"}, "eventDataClientProcessId": {"type": "keyword"}, "eventDataContextInfo": {"type": "keyword"}, "eventDataDCName": {"type": "keyword"}, "eventDataDeviceDescription": {"type": "keyword"}, "eventDataDeviceId": {"type": "keyword"}, "eventDataDisplayName": {"type": "keyword"}, "eventDataErrorCode": {"type": "keyword"}, "eventDataExceptionCode": {"type": "keyword"}, "eventDataFQDN": {"type": "keyword"}, "eventDataFailureReasonDescription": {"type": "keyword"}, "eventDataFilePath": {"type": "keyword"}, "eventDataFilterRTID": {"type": "keyword"}, "eventDataHandleId": {"type": "keyword"}, "eventDataHiveName": {"type": "keyword"}, "eventDataImagePath": {"type": "keyword"}, "eventDataIpPort": {"type": "keyword"}, "eventDataKeyFilePath": {"type": "keyword"}, "eventDataKeyName": {"type": "keyword"}, "eventDataKeyType": {"type": "keyword"}, "eventDataLastBootGood": {"type": "keyword"}, "eventDataLastShutdownGood": {"type": "keyword"}, "eventDataLayerName": {"type": "keyword"}, "eventDataLevel": {"type": "keyword"}, "eventDataLogonGuid": {"type": "keyword"}, "eventDataLogonType": {"type": "keyword"}, "eventDataMemberName": {"type": "keyword"}, "eventDataMemberSid": {"type": "keyword"}, "eventDataMessage": {"type": "keyword"}, "eventDataNewSd": {"type": "keyword"}, "eventDataNewSize": {"type": "keyword"}, "eventDataNewState": {"type": "keyword"}, "eventDataNewTime": {"type": "keyword"}, "eventDataNewValueType": {"type": "keyword"}, "eventDataObjectServer": {"type": "keyword"}, "eventDataObjectType": {"type": "keyword"}, "eventDataObjectValueName": {"type": "keyword"}, "eventDataOldSd": {"type": "keyword"}, "eventDataOldValueType": {"type": "keyword"}, "eventDataOperation": {"type": "keyword"}, "eventDataOperationType": {"type": "keyword"}, "eventDataOriginalSize": {"type": "keyword"}, "eventDataOutcome": {"type": "keyword"}, "eventDataParam1": {"type": "keyword"}, "eventDataParam2": {"type": "keyword"}, "eventDataParentProcessId": {"type": "keyword"}, "eventDataParentProcessName": {"type": "keyword"}, "eventDataPasswordLastSet": {"type": "keyword"}, "eventDataPreAuthType": {"type": "keyword"}, "eventDataPreviousTime": {"type": "keyword"}, "eventDataPrivilegeList": {"type": "keyword"}, "eventDataProcessID": {"type": "keyword"}, "eventDataProcessId": {"type": "keyword"}, "eventDataProfileChanged": {"type": "keyword"}, "eventDataProtocol": {"type": "keyword"}, "eventDataProviderName": {"type": "keyword"}, "eventDataReadOperation": {"type": "keyword"}, "eventDataRelativeTargetName": {"type": "keyword"}, "eventDataRequestId": {"type": "keyword"}, "eventDataReturnCode": {"type": "keyword"}, "eventDataRuleId": {"type": "keyword"}, "eventDataRuleName": {"type": "keyword"}, "eventDataRunspaceId": {"type": "keyword"}, "eventDataSamAccountName": {"type": "keyword"}, "eventDataScriptBlockId": {"type": "keyword"}, "eventDataServiceAccount": {"type": "keyword"}, "eventDataServiceFileName": {"type": "keyword"}, "eventDataServiceName": {"type": "keyword"}, "eventDataServiceSid": {"type": "keyword"}, "eventDataServiceType": {"type": "keyword"}, "eventDataShareLocalPath": {"type": "keyword"}, "eventDataShareName": {"type": "keyword"}, "eventDataShutdownType": {"type": "keyword"}, "eventDataSourceAddress": {"type": "keyword"}, "eventDataSourceHandleId": {"type": "keyword"}, "eventDataSourcePort": {"type": "keyword"}, "eventDataStartType": {"type": "keyword"}, "eventDataStatus": {"type": "keyword"}, "eventDataStatusDescription": {"type": "keyword"}, "eventDataSubjectLogonId": {"type": "keyword"}, "eventDataSubjectUserSid": {"type": "keyword"}, "eventDataTargetHandleId": {"type": "keyword"}, "eventDataTargetLogonGuid": {"type": "keyword"}, "eventDataTargetLogonId": {"type": "keyword"}, "eventDataTargetName": {"type": "keyword"}, "eventDataTargetServerName": {"type": "keyword"}, "eventDataTargetSid": {"type": "keyword"}, "eventDataTargetUserSid": {"type": "keyword"}, "eventDataTaskName": {"type": "keyword"}, "eventDataTemplateName": {"type": "keyword"}, "eventDataTicketEncryptionType": {"type": "keyword"}, "eventDataTicketOptions": {"type": "keyword"}, "eventDataTokenElevationType": {"type": "keyword"}, "eventDataTransactionId": {"type": "keyword"}, "eventDataUserPrincipalName": {"type": "keyword"}, "eventDataVendorIds": {"type": "keyword"}, "eventID": {"type": "keyword"}, "eventType": {"type": "keyword"}, "eventtype": {"type": "keyword"}, "host": {"type": "object", "dynamic": false, "properties": {"os": {"type": "object", "dynamic": false, "properties": {"version": {"type": "keyword"}}}}}, "hostOsVersion": {"type": "keyword"}, "insightDetails": {"type": "object", "dynamic": false, "properties": {"eventName": {"type": "keyword"}}}, "keywords": {"type": "keyword"}, "managementEvent": {"type": "keyword"}, "message": {"type": "keyword"}, "msg": {"type": "keyword"}, "opcode": {"type": "keyword"}, "osType": {"type": "keyword"}, "policyid": {"type": "keyword"}, "priority": {"type": "keyword"}, "process": {"type": "keyword"}, "product": {"type": "keyword"}, "providerGuid": {"type": "keyword"}, "providerName": {"type": "keyword"}, "readOnly": {"type": "keyword"}, "recipientAccountId": {"type": "keyword"}, "recordId": {"type": "keyword"}, "requestID": {"type": "keyword"}, "requestParameters": {"type": "object", "dynamic": false, "properties": {"Host": {"type": "keyword"}, "bucketName": {"type": "keyword"}, "key": {"type": "keyword"}, "logGroupName": {"type": "keyword"}, "policyArn": {"type": "keyword"}, "roleArn": {"type": "keyword"}, "roleName": {"type": "keyword"}, "roleSessionName": {"type": "keyword"}, "userName": {"type": "keyword"}}}, "resourceId": {"type": "keyword"}, "responseElements": {"type": "object", "dynamic": false, "properties": {"ConsoleLogin": {"type": "keyword"}, "assumedRoleUser": {"type": "object", "dynamic": false, "properties": {"arn": {"type": "keyword"}}}, "credentials": {"type": "object", "dynamic": false, "properties": {"accessKeyId": {"type": "keyword"}, "expiration": {"type": "keyword"}}}, "x": {"type": "keyword"}}}, "src_host": {"type": "keyword"}, "subtype": {"type": "keyword"}, "tags": {"type": "keyword"}, "task": {"type": "keyword"}, "threadId": {"type": "keyword"}, "tlsDetails": {"type": "object", "dynamic": false, "properties": {"tlsVersion": {"type": "keyword"}}}, "userAgent": {"type": "keyword"}, "userDataBackupPath": {"type": "keyword"}, "userDomain": {"type": "keyword"}, "userIdentity": {"type": "object", "dynamic": false, "properties": {"accessKeyId": {"type": "keyword"}, "invokedBy": {"type": "keyword"}, "principalId": {"type": "keyword"}, "sessionContext": {"type": "object", "dynamic": false, "properties": {"attributes": {"type": "object", "dynamic": false, "properties": {"mfaAuthenticated": {"type": "keyword"}}}, "sessionIssuer": {"type": "object", "dynamic": false, "properties": {"userName": {"type": "keyword"}}}}}}}}}
}`


// newBagMappings are the mappings added to EXISTING v11-log-* indices on
// upgrade (event/controls are new keys — safe to add in-place; typed
// top-levels are not, so they stay out of the retro PUT).
const newBagMappings = `{"event":{"type":"flat_object"},"controls":{"type":"keyword"},"utm":{"type": "object", "dynamic": false, "properties": {"ActorContextId": {"type": "keyword"}, "AppPoolName": {"type": "keyword"}, "CorrelationId": {"type": "keyword"}, "DetectionMethod": {"type": "keyword"}, "EventData": {"type": "keyword"}, "ExternalAccess": {"type": "keyword"}, "ItemType": {"type": "keyword"}, "LogonError": {"type": "keyword"}, "ObjectId": {"type": "keyword"}, "OrganizationId": {"type": "keyword"}, "OrganizationName": {"type": "keyword"}, "OriginatingServer": {"type": "keyword"}, "P2Sender": {"type": "keyword"}, "Platform": {"type": "keyword"}, "Recipients": {"type": "keyword"}, "RequestId": {"type": "keyword"}, "SourceFileExtension": {"type": "keyword"}, "Subject": {"type": "keyword"}, "TargetUserOrGroupName": {"type": "keyword"}, "TargetUserOrGroupType": {"type": "keyword"}, "UserType": {"type": "keyword"}, "Verdict": {"type": "keyword"}, "Workload": {"type": "keyword"}, "accessType": {"type": "keyword"}, "activityId": {"type": "keyword"}, "additionalEventData": {"type": "object", "dynamic": false, "properties": {"AuthenticationMethod": {"type": "keyword"}, "CipherSuite": {"type": "keyword"}}}, "agentVersion": {"type": "keyword"}, "appAccessContextClientAppId": {"type": "keyword"}, "awsRegion": {"type": "keyword"}, "beatVersion": {"type": "keyword"}, "category": {"type": "keyword"}, "ciscoMsg": {"type": "keyword"}, "deviceIps": {"type": "keyword"}, "eventCategory": {"type": "keyword"}, "eventCode": {"type": "keyword"}, "eventDataAccessList": {"type": "keyword"}, "eventDataAccessMask": {"type": "keyword"}, "eventDataAlgorithmName": {"type": "keyword"}, "eventDataApplication": {"type": "keyword"}, "eventDataBootAppStatus": {"type": "keyword"}, "eventDataCA": {"type": "keyword"}, "eventDataCallerProcessId": {"type": "keyword"}, "eventDataCallerProcessName": {"type": "keyword"}, "eventDataCategoryId": {"type": "keyword"}, "eventDataClassId": {"type": "keyword"}, "eventDataClassName": {"type": "keyword"}, "eventDataClientProcessId": {"type": "keyword"}, "eventDataContextInfo": {"type": "keyword"}, "eventDataDCName": {"type": "keyword"}, "eventDataDeviceDescription": {"type": "keyword"}, "eventDataDeviceId": {"type": "keyword"}, "eventDataDisplayName": {"type": "keyword"}, "eventDataErrorCode": {"type": "keyword"}, "eventDataExceptionCode": {"type": "keyword"}, "eventDataFQDN": {"type": "keyword"}, "eventDataFailureReasonDescription": {"type": "keyword"}, "eventDataFilePath": {"type": "keyword"}, "eventDataFilterRTID": {"type": "keyword"}, "eventDataHandleId": {"type": "keyword"}, "eventDataHiveName": {"type": "keyword"}, "eventDataImagePath": {"type": "keyword"}, "eventDataIpPort": {"type": "keyword"}, "eventDataKeyFilePath": {"type": "keyword"}, "eventDataKeyName": {"type": "keyword"}, "eventDataKeyType": {"type": "keyword"}, "eventDataLastBootGood": {"type": "keyword"}, "eventDataLastShutdownGood": {"type": "keyword"}, "eventDataLayerName": {"type": "keyword"}, "eventDataLevel": {"type": "keyword"}, "eventDataLogonGuid": {"type": "keyword"}, "eventDataLogonType": {"type": "keyword"}, "eventDataMemberName": {"type": "keyword"}, "eventDataMemberSid": {"type": "keyword"}, "eventDataMessage": {"type": "keyword"}, "eventDataNewSd": {"type": "keyword"}, "eventDataNewSize": {"type": "keyword"}, "eventDataNewState": {"type": "keyword"}, "eventDataNewTime": {"type": "keyword"}, "eventDataNewValueType": {"type": "keyword"}, "eventDataObjectServer": {"type": "keyword"}, "eventDataObjectType": {"type": "keyword"}, "eventDataObjectValueName": {"type": "keyword"}, "eventDataOldSd": {"type": "keyword"}, "eventDataOldValueType": {"type": "keyword"}, "eventDataOperation": {"type": "keyword"}, "eventDataOperationType": {"type": "keyword"}, "eventDataOriginalSize": {"type": "keyword"}, "eventDataOutcome": {"type": "keyword"}, "eventDataParam1": {"type": "keyword"}, "eventDataParam2": {"type": "keyword"}, "eventDataParentProcessId": {"type": "keyword"}, "eventDataParentProcessName": {"type": "keyword"}, "eventDataPasswordLastSet": {"type": "keyword"}, "eventDataPreAuthType": {"type": "keyword"}, "eventDataPreviousTime": {"type": "keyword"}, "eventDataPrivilegeList": {"type": "keyword"}, "eventDataProcessID": {"type": "keyword"}, "eventDataProcessId": {"type": "keyword"}, "eventDataProfileChanged": {"type": "keyword"}, "eventDataProtocol": {"type": "keyword"}, "eventDataProviderName": {"type": "keyword"}, "eventDataReadOperation": {"type": "keyword"}, "eventDataRelativeTargetName": {"type": "keyword"}, "eventDataRequestId": {"type": "keyword"}, "eventDataReturnCode": {"type": "keyword"}, "eventDataRuleId": {"type": "keyword"}, "eventDataRuleName": {"type": "keyword"}, "eventDataRunspaceId": {"type": "keyword"}, "eventDataSamAccountName": {"type": "keyword"}, "eventDataScriptBlockId": {"type": "keyword"}, "eventDataServiceAccount": {"type": "keyword"}, "eventDataServiceFileName": {"type": "keyword"}, "eventDataServiceName": {"type": "keyword"}, "eventDataServiceSid": {"type": "keyword"}, "eventDataServiceType": {"type": "keyword"}, "eventDataShareLocalPath": {"type": "keyword"}, "eventDataShareName": {"type": "keyword"}, "eventDataShutdownType": {"type": "keyword"}, "eventDataSourceAddress": {"type": "keyword"}, "eventDataSourceHandleId": {"type": "keyword"}, "eventDataSourcePort": {"type": "keyword"}, "eventDataStartType": {"type": "keyword"}, "eventDataStatus": {"type": "keyword"}, "eventDataStatusDescription": {"type": "keyword"}, "eventDataSubjectLogonId": {"type": "keyword"}, "eventDataSubjectUserSid": {"type": "keyword"}, "eventDataTargetHandleId": {"type": "keyword"}, "eventDataTargetLogonGuid": {"type": "keyword"}, "eventDataTargetLogonId": {"type": "keyword"}, "eventDataTargetName": {"type": "keyword"}, "eventDataTargetServerName": {"type": "keyword"}, "eventDataTargetSid": {"type": "keyword"}, "eventDataTargetUserSid": {"type": "keyword"}, "eventDataTaskName": {"type": "keyword"}, "eventDataTemplateName": {"type": "keyword"}, "eventDataTicketEncryptionType": {"type": "keyword"}, "eventDataTicketOptions": {"type": "keyword"}, "eventDataTokenElevationType": {"type": "keyword"}, "eventDataTransactionId": {"type": "keyword"}, "eventDataUserPrincipalName": {"type": "keyword"}, "eventDataVendorIds": {"type": "keyword"}, "eventID": {"type": "keyword"}, "eventType": {"type": "keyword"}, "eventtype": {"type": "keyword"}, "host": {"type": "object", "dynamic": false, "properties": {"os": {"type": "object", "dynamic": false, "properties": {"version": {"type": "keyword"}}}}}, "hostOsVersion": {"type": "keyword"}, "insightDetails": {"type": "object", "dynamic": false, "properties": {"eventName": {"type": "keyword"}}}, "keywords": {"type": "keyword"}, "managementEvent": {"type": "keyword"}, "message": {"type": "keyword"}, "msg": {"type": "keyword"}, "opcode": {"type": "keyword"}, "osType": {"type": "keyword"}, "policyid": {"type": "keyword"}, "priority": {"type": "keyword"}, "process": {"type": "keyword"}, "product": {"type": "keyword"}, "providerGuid": {"type": "keyword"}, "providerName": {"type": "keyword"}, "readOnly": {"type": "keyword"}, "recipientAccountId": {"type": "keyword"}, "recordId": {"type": "keyword"}, "requestID": {"type": "keyword"}, "requestParameters": {"type": "object", "dynamic": false, "properties": {"Host": {"type": "keyword"}, "bucketName": {"type": "keyword"}, "key": {"type": "keyword"}, "logGroupName": {"type": "keyword"}, "policyArn": {"type": "keyword"}, "roleArn": {"type": "keyword"}, "roleName": {"type": "keyword"}, "roleSessionName": {"type": "keyword"}, "userName": {"type": "keyword"}}}, "resourceId": {"type": "keyword"}, "responseElements": {"type": "object", "dynamic": false, "properties": {"ConsoleLogin": {"type": "keyword"}, "assumedRoleUser": {"type": "object", "dynamic": false, "properties": {"arn": {"type": "keyword"}}}, "credentials": {"type": "object", "dynamic": false, "properties": {"accessKeyId": {"type": "keyword"}, "expiration": {"type": "keyword"}}}, "x": {"type": "keyword"}}}, "src_host": {"type": "keyword"}, "subtype": {"type": "keyword"}, "tags": {"type": "keyword"}, "task": {"type": "keyword"}, "threadId": {"type": "keyword"}, "tlsDetails": {"type": "object", "dynamic": false, "properties": {"tlsVersion": {"type": "keyword"}}}, "userAgent": {"type": "keyword"}, "userDataBackupPath": {"type": "keyword"}, "userDomain": {"type": "keyword"}, "userIdentity": {"type": "object", "dynamic": false, "properties": {"accessKeyId": {"type": "keyword"}, "invokedBy": {"type": "keyword"}, "principalId": {"type": "keyword"}, "sessionContext": {"type": "object", "dynamic": false, "properties": {"attributes": {"type": "object", "dynamic": false, "properties": {"mfaAuthenticated": {"type": "keyword"}}}, "sessionIssuer": {"type": "object", "dynamic": false, "properties": {"userName": {"type": "keyword"}}}}}}}}}}`

// alertIndexMappings pins the nested event bags inside alert documents so they
// use flat_object (same conflict-proof behaviour as the log index). Without
// this, lastEvent.event and events[N].event are dynamically mapped and two
// vendor events with the same field but different types will collide.
// The alert's own top-level fields (severity, status, tags, etc.) keep their
// natural dynamic types — only the embedded event bags need pinning.
const alertIndexMappings = `
{
  "lastEvent": {"type":"object","properties":{
    "event": {"type":"flat_object"},
    "controls": {"type":"keyword"},
    "utm": {"type": "object", "dynamic": false, "properties": {"ActorContextId": {"type": "keyword"}, "AppPoolName": {"type": "keyword"}, "CorrelationId": {"type": "keyword"}, "DetectionMethod": {"type": "keyword"}, "EventData": {"type": "keyword"}, "ExternalAccess": {"type": "keyword"}, "ItemType": {"type": "keyword"}, "LogonError": {"type": "keyword"}, "ObjectId": {"type": "keyword"}, "OrganizationId": {"type": "keyword"}, "OrganizationName": {"type": "keyword"}, "OriginatingServer": {"type": "keyword"}, "P2Sender": {"type": "keyword"}, "Platform": {"type": "keyword"}, "Recipients": {"type": "keyword"}, "RequestId": {"type": "keyword"}, "SourceFileExtension": {"type": "keyword"}, "Subject": {"type": "keyword"}, "TargetUserOrGroupName": {"type": "keyword"}, "TargetUserOrGroupType": {"type": "keyword"}, "UserType": {"type": "keyword"}, "Verdict": {"type": "keyword"}, "Workload": {"type": "keyword"}, "accessType": {"type": "keyword"}, "activityId": {"type": "keyword"}, "additionalEventData": {"type": "object", "dynamic": false, "properties": {"AuthenticationMethod": {"type": "keyword"}, "CipherSuite": {"type": "keyword"}}}, "agentVersion": {"type": "keyword"}, "appAccessContextClientAppId": {"type": "keyword"}, "awsRegion": {"type": "keyword"}, "beatVersion": {"type": "keyword"}, "category": {"type": "keyword"}, "ciscoMsg": {"type": "keyword"}, "deviceIps": {"type": "keyword"}, "eventCategory": {"type": "keyword"}, "eventCode": {"type": "keyword"}, "eventDataAccessList": {"type": "keyword"}, "eventDataAccessMask": {"type": "keyword"}, "eventDataAlgorithmName": {"type": "keyword"}, "eventDataApplication": {"type": "keyword"}, "eventDataBootAppStatus": {"type": "keyword"}, "eventDataCA": {"type": "keyword"}, "eventDataCallerProcessId": {"type": "keyword"}, "eventDataCallerProcessName": {"type": "keyword"}, "eventDataCategoryId": {"type": "keyword"}, "eventDataClassId": {"type": "keyword"}, "eventDataClassName": {"type": "keyword"}, "eventDataClientProcessId": {"type": "keyword"}, "eventDataContextInfo": {"type": "keyword"}, "eventDataDCName": {"type": "keyword"}, "eventDataDeviceDescription": {"type": "keyword"}, "eventDataDeviceId": {"type": "keyword"}, "eventDataDisplayName": {"type": "keyword"}, "eventDataErrorCode": {"type": "keyword"}, "eventDataExceptionCode": {"type": "keyword"}, "eventDataFQDN": {"type": "keyword"}, "eventDataFailureReasonDescription": {"type": "keyword"}, "eventDataFilePath": {"type": "keyword"}, "eventDataFilterRTID": {"type": "keyword"}, "eventDataHandleId": {"type": "keyword"}, "eventDataHiveName": {"type": "keyword"}, "eventDataImagePath": {"type": "keyword"}, "eventDataIpPort": {"type": "keyword"}, "eventDataKeyFilePath": {"type": "keyword"}, "eventDataKeyName": {"type": "keyword"}, "eventDataKeyType": {"type": "keyword"}, "eventDataLastBootGood": {"type": "keyword"}, "eventDataLastShutdownGood": {"type": "keyword"}, "eventDataLayerName": {"type": "keyword"}, "eventDataLevel": {"type": "keyword"}, "eventDataLogonGuid": {"type": "keyword"}, "eventDataLogonType": {"type": "keyword"}, "eventDataMemberName": {"type": "keyword"}, "eventDataMemberSid": {"type": "keyword"}, "eventDataMessage": {"type": "keyword"}, "eventDataNewSd": {"type": "keyword"}, "eventDataNewSize": {"type": "keyword"}, "eventDataNewState": {"type": "keyword"}, "eventDataNewTime": {"type": "keyword"}, "eventDataNewValueType": {"type": "keyword"}, "eventDataObjectServer": {"type": "keyword"}, "eventDataObjectType": {"type": "keyword"}, "eventDataObjectValueName": {"type": "keyword"}, "eventDataOldSd": {"type": "keyword"}, "eventDataOldValueType": {"type": "keyword"}, "eventDataOperation": {"type": "keyword"}, "eventDataOperationType": {"type": "keyword"}, "eventDataOriginalSize": {"type": "keyword"}, "eventDataOutcome": {"type": "keyword"}, "eventDataParam1": {"type": "keyword"}, "eventDataParam2": {"type": "keyword"}, "eventDataParentProcessId": {"type": "keyword"}, "eventDataParentProcessName": {"type": "keyword"}, "eventDataPasswordLastSet": {"type": "keyword"}, "eventDataPreAuthType": {"type": "keyword"}, "eventDataPreviousTime": {"type": "keyword"}, "eventDataPrivilegeList": {"type": "keyword"}, "eventDataProcessID": {"type": "keyword"}, "eventDataProcessId": {"type": "keyword"}, "eventDataProfileChanged": {"type": "keyword"}, "eventDataProtocol": {"type": "keyword"}, "eventDataProviderName": {"type": "keyword"}, "eventDataReadOperation": {"type": "keyword"}, "eventDataRelativeTargetName": {"type": "keyword"}, "eventDataRequestId": {"type": "keyword"}, "eventDataReturnCode": {"type": "keyword"}, "eventDataRuleId": {"type": "keyword"}, "eventDataRuleName": {"type": "keyword"}, "eventDataRunspaceId": {"type": "keyword"}, "eventDataSamAccountName": {"type": "keyword"}, "eventDataScriptBlockId": {"type": "keyword"}, "eventDataServiceAccount": {"type": "keyword"}, "eventDataServiceFileName": {"type": "keyword"}, "eventDataServiceName": {"type": "keyword"}, "eventDataServiceSid": {"type": "keyword"}, "eventDataServiceType": {"type": "keyword"}, "eventDataShareLocalPath": {"type": "keyword"}, "eventDataShareName": {"type": "keyword"}, "eventDataShutdownType": {"type": "keyword"}, "eventDataSourceAddress": {"type": "keyword"}, "eventDataSourceHandleId": {"type": "keyword"}, "eventDataSourcePort": {"type": "keyword"}, "eventDataStartType": {"type": "keyword"}, "eventDataStatus": {"type": "keyword"}, "eventDataStatusDescription": {"type": "keyword"}, "eventDataSubjectLogonId": {"type": "keyword"}, "eventDataSubjectUserSid": {"type": "keyword"}, "eventDataTargetHandleId": {"type": "keyword"}, "eventDataTargetLogonGuid": {"type": "keyword"}, "eventDataTargetLogonId": {"type": "keyword"}, "eventDataTargetName": {"type": "keyword"}, "eventDataTargetServerName": {"type": "keyword"}, "eventDataTargetSid": {"type": "keyword"}, "eventDataTargetUserSid": {"type": "keyword"}, "eventDataTaskName": {"type": "keyword"}, "eventDataTemplateName": {"type": "keyword"}, "eventDataTicketEncryptionType": {"type": "keyword"}, "eventDataTicketOptions": {"type": "keyword"}, "eventDataTokenElevationType": {"type": "keyword"}, "eventDataTransactionId": {"type": "keyword"}, "eventDataUserPrincipalName": {"type": "keyword"}, "eventDataVendorIds": {"type": "keyword"}, "eventID": {"type": "keyword"}, "eventType": {"type": "keyword"}, "eventtype": {"type": "keyword"}, "host": {"type": "object", "dynamic": false, "properties": {"os": {"type": "object", "dynamic": false, "properties": {"version": {"type": "keyword"}}}}}, "hostOsVersion": {"type": "keyword"}, "insightDetails": {"type": "object", "dynamic": false, "properties": {"eventName": {"type": "keyword"}}}, "keywords": {"type": "keyword"}, "managementEvent": {"type": "keyword"}, "message": {"type": "keyword"}, "msg": {"type": "keyword"}, "opcode": {"type": "keyword"}, "osType": {"type": "keyword"}, "policyid": {"type": "keyword"}, "priority": {"type": "keyword"}, "process": {"type": "keyword"}, "product": {"type": "keyword"}, "providerGuid": {"type": "keyword"}, "providerName": {"type": "keyword"}, "readOnly": {"type": "keyword"}, "recipientAccountId": {"type": "keyword"}, "recordId": {"type": "keyword"}, "requestID": {"type": "keyword"}, "requestParameters": {"type": "object", "dynamic": false, "properties": {"Host": {"type": "keyword"}, "bucketName": {"type": "keyword"}, "key": {"type": "keyword"}, "logGroupName": {"type": "keyword"}, "policyArn": {"type": "keyword"}, "roleArn": {"type": "keyword"}, "roleName": {"type": "keyword"}, "roleSessionName": {"type": "keyword"}, "userName": {"type": "keyword"}}}, "resourceId": {"type": "keyword"}, "responseElements": {"type": "object", "dynamic": false, "properties": {"ConsoleLogin": {"type": "keyword"}, "assumedRoleUser": {"type": "object", "dynamic": false, "properties": {"arn": {"type": "keyword"}}}, "credentials": {"type": "object", "dynamic": false, "properties": {"accessKeyId": {"type": "keyword"}, "expiration": {"type": "keyword"}}}, "x": {"type": "keyword"}}}, "src_host": {"type": "keyword"}, "subtype": {"type": "keyword"}, "tags": {"type": "keyword"}, "task": {"type": "keyword"}, "threadId": {"type": "keyword"}, "tlsDetails": {"type": "object", "dynamic": false, "properties": {"tlsVersion": {"type": "keyword"}}}, "userAgent": {"type": "keyword"}, "userDataBackupPath": {"type": "keyword"}, "userDomain": {"type": "keyword"}, "userIdentity": {"type": "object", "dynamic": false, "properties": {"accessKeyId": {"type": "keyword"}, "invokedBy": {"type": "keyword"}, "principalId": {"type": "keyword"}, "sessionContext": {"type": "object", "dynamic": false, "properties": {"attributes": {"type": "object", "dynamic": false, "properties": {"mfaAuthenticated": {"type": "keyword"}}}, "sessionIssuer": {"type": "object", "dynamic": false, "properties": {"userName": {"type": "keyword"}}}}}}}}}
  }},
  "events": {"type":"object","properties":{
    "event": {"type":"flat_object"},
    "controls": {"type":"keyword"},
    "utm": {"type": "object", "dynamic": false, "properties": {"ActorContextId": {"type": "keyword"}, "AppPoolName": {"type": "keyword"}, "CorrelationId": {"type": "keyword"}, "DetectionMethod": {"type": "keyword"}, "EventData": {"type": "keyword"}, "ExternalAccess": {"type": "keyword"}, "ItemType": {"type": "keyword"}, "LogonError": {"type": "keyword"}, "ObjectId": {"type": "keyword"}, "OrganizationId": {"type": "keyword"}, "OrganizationName": {"type": "keyword"}, "OriginatingServer": {"type": "keyword"}, "P2Sender": {"type": "keyword"}, "Platform": {"type": "keyword"}, "Recipients": {"type": "keyword"}, "RequestId": {"type": "keyword"}, "SourceFileExtension": {"type": "keyword"}, "Subject": {"type": "keyword"}, "TargetUserOrGroupName": {"type": "keyword"}, "TargetUserOrGroupType": {"type": "keyword"}, "UserType": {"type": "keyword"}, "Verdict": {"type": "keyword"}, "Workload": {"type": "keyword"}, "accessType": {"type": "keyword"}, "activityId": {"type": "keyword"}, "additionalEventData": {"type": "object", "dynamic": false, "properties": {"AuthenticationMethod": {"type": "keyword"}, "CipherSuite": {"type": "keyword"}}}, "agentVersion": {"type": "keyword"}, "appAccessContextClientAppId": {"type": "keyword"}, "awsRegion": {"type": "keyword"}, "beatVersion": {"type": "keyword"}, "category": {"type": "keyword"}, "ciscoMsg": {"type": "keyword"}, "deviceIps": {"type": "keyword"}, "eventCategory": {"type": "keyword"}, "eventCode": {"type": "keyword"}, "eventDataAccessList": {"type": "keyword"}, "eventDataAccessMask": {"type": "keyword"}, "eventDataAlgorithmName": {"type": "keyword"}, "eventDataApplication": {"type": "keyword"}, "eventDataBootAppStatus": {"type": "keyword"}, "eventDataCA": {"type": "keyword"}, "eventDataCallerProcessId": {"type": "keyword"}, "eventDataCallerProcessName": {"type": "keyword"}, "eventDataCategoryId": {"type": "keyword"}, "eventDataClassId": {"type": "keyword"}, "eventDataClassName": {"type": "keyword"}, "eventDataClientProcessId": {"type": "keyword"}, "eventDataContextInfo": {"type": "keyword"}, "eventDataDCName": {"type": "keyword"}, "eventDataDeviceDescription": {"type": "keyword"}, "eventDataDeviceId": {"type": "keyword"}, "eventDataDisplayName": {"type": "keyword"}, "eventDataErrorCode": {"type": "keyword"}, "eventDataExceptionCode": {"type": "keyword"}, "eventDataFQDN": {"type": "keyword"}, "eventDataFailureReasonDescription": {"type": "keyword"}, "eventDataFilePath": {"type": "keyword"}, "eventDataFilterRTID": {"type": "keyword"}, "eventDataHandleId": {"type": "keyword"}, "eventDataHiveName": {"type": "keyword"}, "eventDataImagePath": {"type": "keyword"}, "eventDataIpPort": {"type": "keyword"}, "eventDataKeyFilePath": {"type": "keyword"}, "eventDataKeyName": {"type": "keyword"}, "eventDataKeyType": {"type": "keyword"}, "eventDataLastBootGood": {"type": "keyword"}, "eventDataLastShutdownGood": {"type": "keyword"}, "eventDataLayerName": {"type": "keyword"}, "eventDataLevel": {"type": "keyword"}, "eventDataLogonGuid": {"type": "keyword"}, "eventDataLogonType": {"type": "keyword"}, "eventDataMemberName": {"type": "keyword"}, "eventDataMemberSid": {"type": "keyword"}, "eventDataMessage": {"type": "keyword"}, "eventDataNewSd": {"type": "keyword"}, "eventDataNewSize": {"type": "keyword"}, "eventDataNewState": {"type": "keyword"}, "eventDataNewTime": {"type": "keyword"}, "eventDataNewValueType": {"type": "keyword"}, "eventDataObjectServer": {"type": "keyword"}, "eventDataObjectType": {"type": "keyword"}, "eventDataObjectValueName": {"type": "keyword"}, "eventDataOldSd": {"type": "keyword"}, "eventDataOldValueType": {"type": "keyword"}, "eventDataOperation": {"type": "keyword"}, "eventDataOperationType": {"type": "keyword"}, "eventDataOriginalSize": {"type": "keyword"}, "eventDataOutcome": {"type": "keyword"}, "eventDataParam1": {"type": "keyword"}, "eventDataParam2": {"type": "keyword"}, "eventDataParentProcessId": {"type": "keyword"}, "eventDataParentProcessName": {"type": "keyword"}, "eventDataPasswordLastSet": {"type": "keyword"}, "eventDataPreAuthType": {"type": "keyword"}, "eventDataPreviousTime": {"type": "keyword"}, "eventDataPrivilegeList": {"type": "keyword"}, "eventDataProcessID": {"type": "keyword"}, "eventDataProcessId": {"type": "keyword"}, "eventDataProfileChanged": {"type": "keyword"}, "eventDataProtocol": {"type": "keyword"}, "eventDataProviderName": {"type": "keyword"}, "eventDataReadOperation": {"type": "keyword"}, "eventDataRelativeTargetName": {"type": "keyword"}, "eventDataRequestId": {"type": "keyword"}, "eventDataReturnCode": {"type": "keyword"}, "eventDataRuleId": {"type": "keyword"}, "eventDataRuleName": {"type": "keyword"}, "eventDataRunspaceId": {"type": "keyword"}, "eventDataSamAccountName": {"type": "keyword"}, "eventDataScriptBlockId": {"type": "keyword"}, "eventDataServiceAccount": {"type": "keyword"}, "eventDataServiceFileName": {"type": "keyword"}, "eventDataServiceName": {"type": "keyword"}, "eventDataServiceSid": {"type": "keyword"}, "eventDataServiceType": {"type": "keyword"}, "eventDataShareLocalPath": {"type": "keyword"}, "eventDataShareName": {"type": "keyword"}, "eventDataShutdownType": {"type": "keyword"}, "eventDataSourceAddress": {"type": "keyword"}, "eventDataSourceHandleId": {"type": "keyword"}, "eventDataSourcePort": {"type": "keyword"}, "eventDataStartType": {"type": "keyword"}, "eventDataStatus": {"type": "keyword"}, "eventDataStatusDescription": {"type": "keyword"}, "eventDataSubjectLogonId": {"type": "keyword"}, "eventDataSubjectUserSid": {"type": "keyword"}, "eventDataTargetHandleId": {"type": "keyword"}, "eventDataTargetLogonGuid": {"type": "keyword"}, "eventDataTargetLogonId": {"type": "keyword"}, "eventDataTargetName": {"type": "keyword"}, "eventDataTargetServerName": {"type": "keyword"}, "eventDataTargetSid": {"type": "keyword"}, "eventDataTargetUserSid": {"type": "keyword"}, "eventDataTaskName": {"type": "keyword"}, "eventDataTemplateName": {"type": "keyword"}, "eventDataTicketEncryptionType": {"type": "keyword"}, "eventDataTicketOptions": {"type": "keyword"}, "eventDataTokenElevationType": {"type": "keyword"}, "eventDataTransactionId": {"type": "keyword"}, "eventDataUserPrincipalName": {"type": "keyword"}, "eventDataVendorIds": {"type": "keyword"}, "eventID": {"type": "keyword"}, "eventType": {"type": "keyword"}, "eventtype": {"type": "keyword"}, "host": {"type": "object", "dynamic": false, "properties": {"os": {"type": "object", "dynamic": false, "properties": {"version": {"type": "keyword"}}}}}, "hostOsVersion": {"type": "keyword"}, "insightDetails": {"type": "object", "dynamic": false, "properties": {"eventName": {"type": "keyword"}}}, "keywords": {"type": "keyword"}, "managementEvent": {"type": "keyword"}, "message": {"type": "keyword"}, "msg": {"type": "keyword"}, "opcode": {"type": "keyword"}, "osType": {"type": "keyword"}, "policyid": {"type": "keyword"}, "priority": {"type": "keyword"}, "process": {"type": "keyword"}, "product": {"type": "keyword"}, "providerGuid": {"type": "keyword"}, "providerName": {"type": "keyword"}, "readOnly": {"type": "keyword"}, "recipientAccountId": {"type": "keyword"}, "recordId": {"type": "keyword"}, "requestID": {"type": "keyword"}, "requestParameters": {"type": "object", "dynamic": false, "properties": {"Host": {"type": "keyword"}, "bucketName": {"type": "keyword"}, "key": {"type": "keyword"}, "logGroupName": {"type": "keyword"}, "policyArn": {"type": "keyword"}, "roleArn": {"type": "keyword"}, "roleName": {"type": "keyword"}, "roleSessionName": {"type": "keyword"}, "userName": {"type": "keyword"}}}, "resourceId": {"type": "keyword"}, "responseElements": {"type": "object", "dynamic": false, "properties": {"ConsoleLogin": {"type": "keyword"}, "assumedRoleUser": {"type": "object", "dynamic": false, "properties": {"arn": {"type": "keyword"}}}, "credentials": {"type": "object", "dynamic": false, "properties": {"accessKeyId": {"type": "keyword"}, "expiration": {"type": "keyword"}}}, "x": {"type": "keyword"}}}, "src_host": {"type": "keyword"}, "subtype": {"type": "keyword"}, "tags": {"type": "keyword"}, "task": {"type": "keyword"}, "threadId": {"type": "keyword"}, "tlsDetails": {"type": "object", "dynamic": false, "properties": {"tlsVersion": {"type": "keyword"}}}, "userAgent": {"type": "keyword"}, "userDataBackupPath": {"type": "keyword"}, "userDomain": {"type": "keyword"}, "userIdentity": {"type": "object", "dynamic": false, "properties": {"accessKeyId": {"type": "keyword"}, "invokedBy": {"type": "keyword"}, "principalId": {"type": "keyword"}, "sessionContext": {"type": "object", "dynamic": false, "properties": {"attributes": {"type": "object", "dynamic": false, "properties": {"mfaAuthenticated": {"type": "keyword"}}}, "sessionIssuer": {"type": "object", "dynamic": false, "properties": {"userName": {"type": "keyword"}}}}}}}}}
  }}
}`

func getOpenSearchContainerID() (string, error) {
	containerIDs, err := utils.RunCmdWithOutput("docker", "ps", "-q", "-f", "name=utmstack_node1")
	if err != nil {
		return "", fmt.Errorf("error getting opensearch container: %v", err)
	}
	if len(containerIDs) == 0 {
		return "", fmt.Errorf("opensearch container not found")
	}
	return containerIDs[0], nil
}

func execCurl(containerID string, method, url, data string) error {
	cnf := config.GetConfig()
	args := []string{"exec", containerID, "curl", "-s", "-k", "-u", "admin:" + cnf.OpenSearchPassword, "-X", method}
	if data != "" {
		args = append(args, "-H", "Content-Type: application/json", "-d", data)
	}
	args = append(args, url)

	_, err := utils.RunCmdWithOutput("docker", args...)
	return err
}

func InitOpenSearch() error {
	containerID, err := getOpenSearchContainerID()
	if err != nil {
		return err
	}

	// Wait for OpenSearch to be ready
	for intent := 0; intent <= 10; intent++ {
		time.Sleep(1 * time.Minute)

		err := execCurl(containerID, "GET", "https://localhost:9200/_cluster/health?wait_for_status=green&timeout=50s", "")
		if err != nil {
			if intent >= 10 {
				return err
			}
		} else {
			break
		}
	}

	// Create snapshot repository
	snapshotData := `{"type":"fs","settings":{"location":"/usr/share/opensearch/.utm_geoip/","compress":true}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_snapshot/.utm_geoip", snapshotData); err != nil {
		return err
	}

	// Create index template
	templateData := `{"index_patterns":["v11-alert-*","v11-log-*",".utm-*",".utmstack-*"],"template":{"settings":{"index.number_of_shards":1,"index.number_of_replicas":0,"index.mapping.total_fields.limit":50000}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_indexes", templateData); err != nil {
		return err
	}

	// v11 settings-only template (covers log + alert) — unchanged from v11.
	logTemplateData := `{"index_patterns":["v11-log-*","v11-alert-*"],"template":{"settings":{"index.max_shards":30000}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_log_indexes", logTemplateData); err != nil {
		return err
	}

	// Event-bag mappings template: scoped to v11-log-* ONLY (alert docs carry
	// severity as an integer — they must not inherit these log-side types).
	// No settings here; settings come from utmstack_log_indexes above.
	// Priority 10 so it composes with (not collides with) the settings-only
	// utmstack_indexes template, which also matches v11-log-* at priority 0.
	eventMappingsData := `{"index_patterns":["v11-log-*"],"priority":10,"template":{"mappings":{"properties":` + logIndexMappings + `}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_log_event_mappings", eventMappingsData); err != nil {
		return err
	}

	// Alert-document mappings template: alerts embed full Event copies at
	// lastEvent and events[*] — each carrying an `event` bag. Without this,
	// those nested bags are dynamically mapped, and two alerts whose events
	// disagree on a bag field's type (int vs string) will collide, failing
	// alert ingestion. Pin them as flat_object (conflict-proof). The alert's
	// own top-level fields (severity int, status, tags…) stay dynamic.
	alertMappingsData := `{"index_patterns":["v11-alert-*"],"priority":10,"template":{"mappings":{"properties":` + alertIndexMappings + `}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_alert_event_mappings", alertMappingsData); err != nil {
		return err
	}


	// Restore geoip snapshot
	restoreData := `{"indices":".utm-geoip","include_global_state":false}`
	if err := execCurl(containerID, "POST", "https://localhost:9200/_snapshot/.utm_geoip/.utm_geoip/_restore", restoreData); err != nil {
		return err
	}

	return nil
}

func UpdateOpenSearch() error{

	containerID, err := getOpenSearchContainerID()
	if err != nil {
		return err
	}

	// (Re)create the v11 settings-only log template (covers log + alert).
	logTemplateData := `{"index_patterns":["v11-log-*","v11-alert-*"],"template":{"settings":{"index.max_shards":30000}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_log_indexes", logTemplateData); err != nil {
		return err
	}

	// (Re)create the event-bag mappings template (v11-log-* only).
	// Priority 10 so it composes with (not collides with) the settings-only
	// utmstack_indexes template, which also matches v11-log-* at priority 0.
	eventMappingsData := `{"index_patterns":["v11-log-*"],"priority":10,"template":{"mappings":{"properties":` + logIndexMappings + `}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_log_event_mappings", eventMappingsData); err != nil {
		return err
	}

	// (Re)create the alert-document event-bag mappings template (v11-alert-*).
	alertMappingsData := `{"index_patterns":["v11-alert-*"],"priority":10,"template":{"mappings":{"properties":` + alertIndexMappings + `}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_alert_event_mappings", alertMappingsData); err != nil {
		return err
	}

	// Add ONLY the new bag mappings to already-existing v11-log-* indices
	// (non-destructive: event/controls are new keys - no type clash with the
	// legacy dynamic log mapping). Typed top-levels are NOT retro-applied: they
	// can conflict with existing dynamic mappings and would 400 the upgrade.
	if err := execCurl(containerID, "PUT", "https://localhost:9200/v11-log-*/_mapping?allow_no_indices=true",
		`{"properties":` + newBagMappings + `}`); err != nil {
		return err
	}

	// Add the nested event-bag mappings to already-existing v11-alert-* indices.
	// lastEvent.event / events[*].event are new keys post-rename (previously
	// lastEvent.log), so this is additive - no conflict with existing mappings.
	if err := execCurl(containerID, "PUT", "https://localhost:9200/v11-alert-*/_mapping?allow_no_indices=true",
		`{"properties":` + alertIndexMappings + `}`); err != nil {
		return err
	}

	// updated already existing index (update case)
	if err := execCurl(containerID, "PUT", "https://localhost:9200/v11-log-*,v11-alert-*/_settings?allow_no_indices=true", `{"index.max_shards":30000}`); err != nil {
		return err
	}

	if err := execCurl(containerID, "PUT", "https://localhost:9200/v11-alert-*,v11-log-*,.utm-*,.utmstack-*/_settings?allow_no_indices=true", `{"index.mapping.total_fields.limit":50000}`); err != nil {
		return err
	}
	return nil

}

