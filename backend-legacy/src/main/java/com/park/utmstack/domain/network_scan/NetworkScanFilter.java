package com.park.utmstack.domain.network_scan;

import com.park.utmstack.domain.network_scan.enums.AssetRegisteredMode;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import lombok.Getter;
import lombok.Setter;

import java.time.Instant;
import java.util.List;

@Setter
@Getter
public class NetworkScanFilter {
    private String assetIpMacName;
    private List<String> os;
    private List<Boolean> alive;
    private List<AssetStatus> status;
    private List<String> type;
    private List<String> alias;
    private List<String> probe;
    private List<Integer> openPorts;
    private List<String> groups;
    private Instant discoveredInitDate;
    private Instant discoveredEndDate;
    private AssetRegisteredMode registeredMode;
    private List<Boolean> agent;
    private List<String> osPlatform;
    private List<String> dataTypes;
}
