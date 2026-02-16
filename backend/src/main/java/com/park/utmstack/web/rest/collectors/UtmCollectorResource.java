package com.park.utmstack.web.rest.collectors;

import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.network_scan.AssetGroupFilter;
import com.park.utmstack.domain.network_scan.NetworkScanFilter;
import com.park.utmstack.service.application_modules.UtmModuleGroupService;
import com.park.utmstack.service.collectors.CollectorService;
import com.park.utmstack.service.collectors.UtmCollectorService;
import com.park.utmstack.service.dto.collectors.CollectorActionEnum;
import com.park.utmstack.service.dto.collectors.CollectorHostnames;
import com.park.utmstack.service.dto.collectors.dto.CollectorConfigDTO;
import com.park.utmstack.service.dto.collectors.dto.CollectorDTO;
import com.park.utmstack.service.dto.collectors.CollectorModuleEnum;
import com.park.utmstack.service.dto.collectors.dto.ListCollectorsResponseDTO;
import com.park.utmstack.service.dto.network_scan.AssetGroupDTO;
import com.park.utmstack.service.dto.network_scan.UpdateGroupDTO;
import com.park.utmstack.service.grpc.ListRequest;
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
@RequestMapping("/api")
public class UtmCollectorResource {

    private final UtmModuleGroupService moduleGroupService;
    private final UtmCollectorService utmCollectorService;
    private final CollectorService collectorService;

    @PostMapping("/collector-config")
    public ResponseEntity<Void> upsertCollectorConfig(@Valid @RequestBody CollectorConfigDTO collectorConfig,
                                                      @RequestParam(name = "action", defaultValue = "CREATE") CollectorActionEnum action) {

        collectorService.upsertCollectorConfig(collectorConfig);
        return ResponseEntity.noContent().build();
    }


    @GetMapping("/collectors-list")
    public ResponseEntity<ListCollectorsResponseDTO> listCollectorsByModule(@RequestParam(required = false) Integer pageNumber,
                                                                            @RequestParam(required = false) Integer pageSize,
                                                                            @RequestParam(required = false) CollectorModuleEnum module,
                                                                            @RequestParam(required = false) String sortBy) {


        ListCollectorsResponseDTO response = collectorService.listCollector(null, module);
        HttpHeaders headers = new HttpHeaders();
        headers.add("X-Total-Count", Long.toString(response.getTotal()));
        return ResponseEntity.ok().headers(headers).body(response);
    }


    @GetMapping("/collector-hostnames")
    public ResponseEntity<CollectorHostnames> listCollectorHostNames(@RequestParam(required = false) Integer pageNumber,
                                                                     @RequestParam(required = false) Integer pageSize,
                                                                     @RequestParam(required = false) CollectorModuleEnum module,
                                                                     @RequestParam(required = false) String sortBy) {

        ListRequest request = ListRequest.newBuilder()
                .setPageNumber(pageNumber != null ? pageNumber : 0)
                .setPageSize(pageSize != null ? pageSize : 1000000)
                .setSearchQuery(module != null ? "module.Is=" + module : "")
                .setSortBy(sortBy != null ? sortBy : "")
                .build();

        return ResponseEntity.ok().body(collectorService.listCollectorHostnames(request));

    }

    /**
     * {@code GET  /collector-by-hostname-and-module} : Get collector's list according to the request params.
     *
     * @param hostname the host name to search for.
     * @param module   the collector module to search for
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, status {@code 400 (Bad request)} if the internal key is not set,
     * or with status {@code 502 (Bad Gateway)} if the agent manager returns an error.
     */
    @GetMapping("/collector-by-hostname-and-module")
    public ResponseEntity<ListCollectorsResponseDTO> listCollectorByHostNameAndModule(@RequestParam String hostname,
                                                                                      @RequestParam CollectorModuleEnum module) {

        return ResponseEntity.ok().body(collectorService.listCollector(hostname, module));
    }

    @GetMapping("/groups-by-collectors/{collectorId}")
    public ResponseEntity<List<UtmModuleGroup>> getModuleGroups(@PathVariable String collectorId) {

        return ResponseEntity.ok(moduleGroupService.findAllByCollectorId(collectorId));

    }

    @PutMapping("/updateGroup")
    public ResponseEntity<Void> updateGroup(@Valid @RequestBody UpdateGroupDTO body) {

        utmCollectorService.updateGroup(body.getAssetsIds(), body.getAssetGroupId());

        return ResponseEntity.ok().build();
    }


    @GetMapping("/searchGroupsByFilter")
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

    @DeleteMapping("/collectors/{id}")
    public ResponseEntity<Void> deleteCollector(@PathVariable Long id) {
        collectorService.deleteCollector(id);
        return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert("UtmCollector", id.toString())).build();
    }

}
