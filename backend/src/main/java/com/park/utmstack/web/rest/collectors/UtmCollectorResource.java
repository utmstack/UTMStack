package com.park.utmstack.web.rest.collectors;

import com.park.utmstack.aop.logging.AuditEvent;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.network_scan.AssetGroupFilter;
import com.park.utmstack.domain.network_scan.NetworkScanFilter;
import com.park.utmstack.service.application_modules.UtmModuleGroupService;
import com.park.utmstack.service.collectors.CollectorService;
import com.park.utmstack.service.collectors.UtmCollectorService;
import com.park.utmstack.service.dto.collectors.CollectorActionEnum;
import com.park.utmstack.service.dto.collectors.dto.CollectorConfigDTO;
import com.park.utmstack.service.dto.collectors.dto.CollectorDTO;
import com.park.utmstack.service.dto.collectors.CollectorModuleEnum;
import com.park.utmstack.service.dto.collectors.dto.ListCollectorsResponseDTO;
import com.park.utmstack.service.dto.network_scan.AssetGroupDTO;
import com.park.utmstack.service.dto.network_scan.UpdateGroupDTO;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import lombok.RequiredArgsConstructor;
import org.springdoc.api.annotations.ParameterObject;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;


/**
 * REST controller for managing {@link UtmCollectorResource}.
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/api/collectors")
public class UtmCollectorResource {

    private final UtmModuleGroupService moduleGroupService;
    private final UtmCollectorService utmCollectorService;
    private final CollectorService collectorService;

    @AuditEvent(
            attemptType = ApplicationEventType.CONFIG_UPDATE_ATTEMPT,
            successType = ApplicationEventType.CONFIG_UPDATE_SUCCESS,
            attemptMessage = "Attempt to upsert collector configuration initiated",
            successMessage = "Collector configuration upserted successfully"
    )
    @PostMapping("/config")
    public ResponseEntity<Void> upsertCollectorConfig(@Valid @RequestBody CollectorConfigDTO collectorConfig,
                                                      @RequestParam(name = "action", defaultValue = "CREATE") CollectorActionEnum action) {

        collectorService.upsertCollectorConfig(collectorConfig);
        return ResponseEntity.noContent().build();
    }

    @GetMapping
    public ResponseEntity<ListCollectorsResponseDTO> listCollectorsByModule(@RequestParam(required = false, defaultValue = "0") Integer pageNumber,
                                                                            @RequestParam(required = false, defaultValue = "10") Integer pageSize,
                                                                            @RequestParam(required = false) String hostname,
                                                                            @RequestParam(required = false) CollectorModuleEnum module,
                                                                            @RequestParam(required = false) String sortBy) {


        ListCollectorsResponseDTO response = collectorService.listCollector(hostname, pageNumber, pageSize, sortBy, module);
        HttpHeaders headers = new HttpHeaders();
        headers.add("X-Total-Count", Long.toString(response.getTotal()));
        return ResponseEntity.ok().headers(headers).body(response);
    }

    @GetMapping("/{collectorId}/module-groups")
    public ResponseEntity<List<UtmModuleGroup>> getModuleGroups(@PathVariable String collectorId) {

        return ResponseEntity.ok(moduleGroupService.findAllByCollectorId(collectorId));

    }

    @PutMapping("/asset-groups")
    public ResponseEntity<Void> updateGroup(@Valid @RequestBody UpdateGroupDTO body) {

        utmCollectorService.updateGroup(body.getAssetsIds(), body.getAssetGroupId());

        return ResponseEntity.ok().build();
    }


    @GetMapping("/asset-groups")
    public ResponseEntity<List<AssetGroupDTO>> searchGroupsByFilter(AssetGroupFilter filter, Pageable pageable) {


        Page<AssetGroupDTO> page = collectorService.searchGroupsByFilter(filter, pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/utm-asset-groups/searchGroupsByFilter");
        return ResponseEntity.ok().headers(headers).body(page.getContent());

    }

    @GetMapping("/search-by-filters")
    public ResponseEntity<List<CollectorDTO>> searchByFilters(@ParameterObject NetworkScanFilter filters,
                                                              @ParameterObject Pageable pageable) {

        Page<CollectorDTO> page = this.utmCollectorService.searchByFilters(filters, pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/search-by-filters");
        return ResponseEntity.ok().headers(headers).body(page.getContent());

    }

    @AuditEvent(
            attemptType = ApplicationEventType.COLLECTOR_DELETE_ATTEMPT,
            successType = ApplicationEventType.COLLECTOR_DELETE_SUCCESS,
            attemptMessage = "Attempt to delete collector initiated",
            successMessage = "Collector deleted successfully"
    )
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteCollector(@PathVariable Long id) {
        collectorService.deleteCollector(id);
        return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert("UtmCollector", id.toString())).build();
    }

}
