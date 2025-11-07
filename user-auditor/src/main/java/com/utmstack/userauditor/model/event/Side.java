package com.utmstack.userauditor.model.event;

import lombok.Data;

@Data
public class Side {
    private long bytesSent;
    private long bytesReceived;
    private long packagesSent;
    private long packagesReceived;
    private int connections;
    private double usedCpuPercent;
    private double usedMemPercent;
    private int totalCpuUnits;
    private long totalMem;
    private String ip;
    private String host;
    private String user;
    private String group;
    private int port;
    private String domain;
    private String fqdn;
    private String mac;
    private String process;
    private Geolocation geolocation;
    private String file;
    private String path;
    private String hash;
    private String url;
    private String email;
    // getters and setters
}

