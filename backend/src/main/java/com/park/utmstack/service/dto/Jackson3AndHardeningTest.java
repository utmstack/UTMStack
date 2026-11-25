package com.park.utmstack.service.dto;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.UtmPorts;
import com.park.utmstack.repository.notification.UtmNotificationRepository;
import com.park.utmstack.service.api_key.ApiKeyService;
import com.park.utmstack.service.dto.logstash_pipeline.UtmLogstashPipelineDTO;
import com.park.utmstack.service.dto.network_scan.NetworkScanDTO;
import com.park.utmstack.service.dto.notification.NotificationDTO;
import com.park.utmstack.web.rest.vm.AlertTagRuleVM;
import com.park.utmstack.web.rest.vm.UtmCorrelationRulesVM;
import com.park.utmstack.web.rest.vm.UtmLogstashPipelineVM;
import org.junit.jupiter.api.Test;
import org.springframework.data.jpa.repository.Query;
import tools.jackson.databind.DeserializationFeature;
import tools.jackson.databind.json.JsonMapper;

import java.lang.reflect.Method;
import java.time.LocalDate;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Regression tests for Spring Boot 4 / Jackson 3 constructor discovery,
 * API key daily index naming, and notification single-bound date filtering.
 */
class Jackson3AndHardeningTest {

    private final JsonMapper jackson3 = JsonMapper.builder()
        .disable(DeserializationFeature.FAIL_ON_NULL_FOR_PRIMITIVES)
        .build();
    private final ObjectMapper jackson2 = new ObjectMapper();

    @Test
    void multiConstructorDtosDeserializeCleanlyInJackson3AndJackson2() throws Exception {
        String alertTagJson = "{\"name\":\"Tag Rule\",\"description\":\"Desc\",\"active\":true,\"conditions\":[],\"tags\":[]}";
        AlertTagRuleVM tagRule3 = jackson3.readValue(alertTagJson, AlertTagRuleVM.class);
        assertEquals("Tag Rule", tagRule3.getName());
        assertEquals("Tag Rule", jackson2.readValue(alertTagJson, AlertTagRuleVM.class).getName());

        String assetJson = "{\"assetName\":\"host-1\",\"assetIp\":\"198.51.100.10\",\"assetAlive\":true,\"isAgent\":false}";
        NetworkScanDTO asset3 = jackson3.readValue(assetJson, NetworkScanDTO.class);
        assertEquals("host-1", asset3.getAssetName());
        assertEquals("198.51.100.10", jackson2.readValue(assetJson, NetworkScanDTO.class).getAssetIp());

        String pipelineVmJson = "{\"pipeline\":{\"pipelineName\":\"pipe-1\"},\"filters\":[]}";
        UtmLogstashPipelineVM pipeVm3 = jackson3.readValue(pipelineVmJson, UtmLogstashPipelineVM.class);
        assertNotNull(pipeVm3.getPipelineDTO());
        assertEquals("pipe-1", pipeVm3.getPipelineDTO().getPipelineName());

        String corrVmJson = "{\"rule\":null,\"dataTypeRelations\":[]}";
        assertNotNull(jackson3.readValue(corrVmJson, UtmCorrelationRulesVM.class));

        String portJson = "{\"port\":443,\"tcp\":\"open\"}";
        UtmPorts port3 = jackson3.readValue(portJson, UtmPorts.class);
        assertEquals(443, port3.getPort());

        String groupJson = "{\"groupName\":\"Tenant A\",\"moduleId\":34}";
        UtmModuleGroup group3 = jackson3.readValue(groupJson, UtmModuleGroup.class);
        assertEquals("Tenant A", group3.getGroupName());

        String scanJson = "{\"assetName\":\"host-2\",\"assetIp\":\"203.0.113.5\"}";
        assertEquals("host-2", jackson3.readValue(scanJson, UtmNetworkScan.class).getAssetName());

        String pipeDtoJson = "{\"pipelineName\":\"pipe-2\"}";
        assertEquals("pipe-2", jackson3.readValue(pipeDtoJson, UtmLogstashPipelineDTO.class).getPipelineName());

        String filterJson = "{\"field\":\"category\",\"operator\":\"IS\",\"value\":\"Malware\"}";
        assertEquals("category", jackson3.readValue(filterJson, FilterType.class).getField());

        String notifJson = "{\"source\":\"EMAIL_SETTING\",\"type\":\"INFO\",\"message\":\"Test message\"}";
        NotificationDTO notif3 = jackson3.readValue(notifJson, NotificationDTO.class);
        assertFalse(notif3.isRead());
    }

    @Test
    void apiAccessLogsIndexIsConcreteDailyIndexMatchingWildcardPattern() {
        String idx = ApiKeyService.apiAccessLogsIndexForDate(LocalDate.of(2026, 11, 24));
        assertEquals("v11-api-access-logs-2026-11-24", idx);
        assertFalse(idx.contains("*"), "Index name passed to OpenSearch index() must never contain wildcard *");
    }

    @Test
    void notificationDateQuerySupportsSingleSidedBounds() throws Exception {
        Method m = UtmNotificationRepository.class.getMethod(
            "searchByFilters",
            com.park.utmstack.domain.notification.NotificationSource.class,
            com.park.utmstack.domain.notification.NotificationType.class,
            com.park.utmstack.domain.notification.NotificationStatus.class,
            java.time.LocalDateTime.class,
            java.time.LocalDateTime.class,
            org.springframework.data.domain.Pageable.class
        );
        Query q = m.getAnnotation(Query.class);
        assertNotNull(q);
        assertTrue(q.value().contains("n.createdAt >= :from"), "Query must allow open-ended from bound: " + q.value());
        assertTrue(q.value().contains("n.createdAt <= :to"), "Query must allow open-ended to bound: " + q.value());
        assertFalse(q.value().toUpperCase().contains("BETWEEN :FROM AND :TO"), "BETWEEN fails when only one bound is supplied");
    }
}
