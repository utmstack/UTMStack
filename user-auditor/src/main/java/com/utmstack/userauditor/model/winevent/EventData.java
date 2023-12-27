package com.utmstack.userauditor.model.winevent;

import com.fasterxml.jackson.annotation.JsonProperty;

public class EventData {
    @JsonProperty("AuthenticationPackageName")
    public String authenticationPackageName;

    @JsonProperty("AllowedToDelegateTo")
    public String allowedToDelegateTo;

    @JsonProperty("AccountExpires")
    public String accountExpires;

    @JsonProperty("ElevatedToken")
    public String elevatedToken;

    @JsonProperty("ImpersonationLevel")
    public String impersonationLevel;

    @JsonProperty("IpAddress")
    public String ipAddress;

    @JsonProperty("IpPort")
    public String ipPort;

    @JsonProperty("KeyLength")
    public String keyLength;

    @JsonProperty("LmPackageName")
    public String lmPackageName;

    @JsonProperty("LogonGuid")
    public String logonGuid;

    @JsonProperty("LogonProcessName")
    public String logonProcessName;

    @JsonProperty("LogonType")
    public String logonType;

    @JsonProperty("ProcessId")
    public String processId;

    @JsonProperty("ProcessName")
    public String processName;

    @JsonProperty("PrivilegeList")
    public String privilegeList;

    @JsonProperty("RestrictedAdminMode")
    public String restrictedAdminMode;

    @JsonProperty("SubjectDomainName")
    public String subjectDomainName;

    @JsonProperty("SubjectLogonId")
    public String subjectLogonId;

    @JsonProperty("SubjectUserName")
    public String subjectUserName;

    @JsonProperty("SubjectUserSid")
    public String subjectUserSid;

    @JsonProperty("TargetDomainName")
    public String targetDomainName;

    @JsonProperty("TargetLinkedLogonId")
    public String targetLinkedLogonId;

    @JsonProperty("TargetSid")
    public String targetSid;

    @JsonProperty("TargetLogonId")
    public String targetLogonId;

    @JsonProperty("TargetOutboundDomainName")
    public String targetOutboundDomainName;

    @JsonProperty("TargetOutboundUserName")
    public String targetOutboundUserName;

    @JsonProperty("TargetUserName")
    public String targetUserName;

    @JsonProperty("TargetUserSid")
    public String targetUserSid;

    @JsonProperty("TransmittedServices")
    public String transmittedServices;

    @JsonProperty("VirtualAccount")
    public String virtualAccount;

    @JsonProperty("WorkstationName")
    public String workstationName;

    @JsonProperty("DisplayName")
    public String displayName;

    @JsonProperty("HomeDirectory")
    public String homeDirectory;

    @JsonProperty("HomePath")
    public String homePath;

    @JsonProperty("LogonHours")
    public String logonHours;

    @JsonProperty("NewUacValue")
    public String newUacValue;

    @JsonProperty("OldUacValue")
    public String oldUacValue;

    @JsonProperty("PasswordLastSet")
    public String passwordLastSet;

    @JsonProperty("PrimaryGroupId")
    public String primaryGroupId;

    @JsonProperty("ProfilePath")
    public String profilePath;

    @JsonProperty("SamAccountName")
    public String samAccountName;

    @JsonProperty("ScriptPath")
    public String scriptPath;

    @JsonProperty("SidHistory")
    public String sidHistory;

    @JsonProperty("UserAccountControl")
    public String userAccountControl;

    @JsonProperty("UserParameters")
    public String userParameters;

    @JsonProperty("UserPrincipalName")
    public String userPrincipalName;

    @JsonProperty("UserWorkstations")
    public String userWorkstations;

    @JsonProperty("RemoteCredentialGuard")
    public String remoteCredentialGuard;
}
