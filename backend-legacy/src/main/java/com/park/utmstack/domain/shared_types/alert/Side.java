package com.park.utmstack.domain.shared_types.alert;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.park.utmstack.domain.shared_types.Geolocation;
import lombok.Data;
import java.util.List;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
@JsonIgnoreProperties(ignoreUnknown = true)
public class Side {

    // Network traffic attributes
    private Double bytesSent;
    private Double bytesReceived;
    private Long packagesSent;
    private Long packagesReceived;

    // Network identification attributes
    private String ip;
    private String host;
    private String user;
    private String group;
    private Integer port;
    private String domain;
    private String mac;
    private Geolocation geolocation;
    private String url;
    private String cidr;

    // Certificate and fingerprint attributes
    private String certificateFingerprint;
    private String ja3Fingerprint;
    private String jarmFingerprint;
    private String sshBanner;
    private String sshFingerprint;

    // Web attributes
    private String cookie;
    private String jabberId;

    // Email attributes
    private String email;
    private String dkim;
    private String dkimSignature;
    private String emailAddress;
    private String emailBody;
    private String emailDisplayName;
    private String emailSubject;
    private String emailThreadIndex;
    private String emailXMailer;

    // WHOIS attributes
    private String whoisRegistrant;
    private String whoisRegistrar;

    // Process-related attributes
    private String process;
    private String processState;
    private String command;
    private String windowsScheduledTask;
    private String windowsServiceDisplayName;
    private String windowsServiceName;

    // File-related attributes
    private String file;
    private String path;
    private String filename;
    private String sizeInBytes;
    private String mimeType;

    // Hash-related attributes
    private String hash;
    private String authentihash;
    private String cdhash;
    private String md5;
    private String sha1;
    private String sha224;
    private String sha256;
    private String sha384;
    private String sha3224;
    private String sha3256;
    private String sha3384;
    private String sha3512;
    private String sha512;
    private String sha512224;
    private String sha512256;
    private String hex;
    private String base64;

    // System-related attributes
    private String operatingSystem;
    private String chromeExtension;
    private String mobileAppId;

    // Vulnerability-related attributes
    private String cpe;
    private String cve;

    // Malware-related attributes
    private String malware;
    private String malwareFamily;
    private String malwareType;

    // Key-related attributes
    private String pgpPrivateKey;
    private String pgpPublicKey;

    // Resources attributes
    private Long connections;
    private Integer usedCpuPercent;
    private Integer usedMemPercent;
    private Integer totalCpuUnits;
    private Long totalMem;
    private List<DiskInfo> disks;
}

