package com.park.utmstack.service.app_info;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.park.utmstack.service.dto.app_info.AppInfoDto;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.io.File;
import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.concurrent.atomic.AtomicLong;

import static com.park.utmstack.config.Constants.APP_VERSION_FILE;

@Service
@Slf4j
public class AppInfoService {

    private static final String CLASSNAME = "AppInfoService";
    private static final String DEFAULT_CM_HEALTH_URL = "https://cm.utmstack.com/api/v1/health";
    private static final long CONNECTIVITY_CACHE_TTL_MS = Duration.ofSeconds(60).toMillis();
    private static final Duration PROBE_TIMEOUT = Duration.ofSeconds(3);

    private final String cmHealthUrl;
    private final HttpClient httpClient;
    private final AtomicLong lastCheckedAtMillis = new AtomicLong(0L);
    private volatile boolean lastOnline = false;

    public AppInfoService(
            @Value("${utmstack.cm.health-url:" + DEFAULT_CM_HEALTH_URL + "}") String cmHealthUrl) {
        this.cmHealthUrl = cmHealthUrl;
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(PROBE_TIMEOUT)
                .build();
    }

    public AppInfoDto loadVersionInfo() {
        final String ctx = "loadVersionInfo";
        try {
            ObjectMapper mapper = new ObjectMapper();
            return mapper.readValue(new File(APP_VERSION_FILE), AppInfoDto.class);
        } catch (IOException e) {
            log.error("{}: An error occurred while reading the version file: {}", CLASSNAME + "." + ctx, e.getMessage(), e);
            throw new RuntimeException("An error occurred while reading the version file");
        }
    }

    /**
     * Checks whether the SIEM server can reach Customer Manager's health endpoint,
     * caching the result for 60 seconds so open UI tabs do not hammer Customer Manager.
     */
    public boolean isCustomerManagerReachable() {
        long now = System.currentTimeMillis();
        long last = lastCheckedAtMillis.get();
        if (last > 0 && (now - last) < CONNECTIVITY_CACHE_TTL_MS) {
            return lastOnline;
        }
        boolean reachable = probeCustomerManager();
        lastOnline = reachable;
        lastCheckedAtMillis.set(now);
        return reachable;
    }

    private boolean probeCustomerManager() {
        try {
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(cmHealthUrl))
                    .timeout(PROBE_TIMEOUT)
                    .GET()
                    .build();
            HttpResponse<Void> response = httpClient.send(request, HttpResponse.BodyHandlers.discarding());
            int status = response.statusCode();
            return status >= 200 && status < 400;
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            log.debug("{}.probeCustomerManager: interrupted probing {}: {}", CLASSNAME, cmHealthUrl, e.getMessage());
            return false;
        } catch (Exception e) {
            log.debug("{}.probeCustomerManager: Customer Manager unreachable at {}: {}", CLASSNAME, cmHealthUrl, e.getMessage());
            return false;
        }
    }
}

