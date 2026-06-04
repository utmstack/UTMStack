package com.park.utmstack.service.network_scan;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.network_scan.NetworkScan;
import com.park.utmstack.service.ApplicationPropertyService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.web_clients.rest_template.RestTemplateService;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.List;

@Service
public class NetworkScanService {
    private static final String CLASSNAME = "NetworkScanService";

    private final RestTemplateService restTemplate;
    private final ApplicationPropertyService propertyService;
    private final ApplicationEventService eventService;

    public NetworkScanService(RestTemplateService restTemplate,
                              ApplicationPropertyService propertyService,
                              ApplicationEventService eventService) {
        this.restTemplate = restTemplate;
        this.propertyService = propertyService;
        this.eventService = eventService;
    }

    /**
     * @param iface
     * @param ranges
     * @return
     * @throws Exception
     */
    public List<NetworkScan> getNetworkScanResults(String apiUrl, String iface, String[] ranges) throws Exception {
        final String ctx = CLASSNAME + ".getNetworkScanResults";
        try {
            List<NetworkScan> result = new ArrayList<>();
            ObjectMapper om = new ObjectMapper();

            for (String range : ranges) {
                String url = String.format("%1$s/api/v1/scanner/scan?interface=%2$s&network=%3$s", apiUrl, iface, range);
                ResponseEntity<String> rs = restTemplate.get(url, String.class);

                if (StringUtils.hasText(rs.getBody())) {
                    List<NetworkScan> scanResponse = om.readValue(rs.getBody(), new TypeReference<>() {
                    });

                    if (!CollectionUtils.isEmpty(scanResponse))
                        result.addAll(scanResponse);
                }
            }
            return result;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public boolean ping(String apiBaseUrl) {
        final String ctx = CLASSNAME + ".ping";
        try {
            if (apiBaseUrl.endsWith("/"))
                apiBaseUrl = apiBaseUrl.substring(0, apiBaseUrl.length() - 1);
            final String URL = apiBaseUrl + "/api/v1/checks/ping";
            ResponseEntity<String> rs = restTemplate.get(URL, String.class);
            return rs != null && rs.getStatusCodeValue() == 200;
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage() + ". Localized message: " + e.getLocalizedMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return false;
        }
    }

    public boolean checkInterface(String iface, Long probeId) throws Exception {
        final String ctx = CLASSNAME + ".checkInterface";
        try {
            Assert.hasText(iface, "iface parameter is nul or empty");

            String baseUrl = getNetworkScanApiBaseUrl(probeId);
            final String URL = baseUrl + "/api/v1/checks/check_interface?interface=" + iface;
            ResponseEntity<String> rs = restTemplate.get(URL, String.class);

            ObjectMapper om = new ObjectMapper();
            return om.readValue(rs.getBody(), Boolean.class);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage() + ". Localized message: " + e.getLocalizedMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return false;
        }
    }

    private String getNetworkScanApiBaseUrl(Long probeId) throws Exception {
        final String ctx = CLASSNAME + ".getNetworkScanApiBaseUrl";
        try {
            return Constants.CFG.get(Constants.PROP_NETWORK_SCAN_API_URL + "." + probeId);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
