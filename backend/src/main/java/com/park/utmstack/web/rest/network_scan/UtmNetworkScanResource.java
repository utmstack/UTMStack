package com.park.utmstack.web.rest.network_scan;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.network_scan.NetworkScanFilter;
import com.park.utmstack.domain.network_scan.Property;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.enums.PropertyFilter;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.network_scan.NetworkScanDTO;
import com.park.utmstack.service.dto.network_scan.UtmNetworkScanCriteria;
import com.park.utmstack.service.network_scan.UtmNetworkScanQueryService;
import com.park.utmstack.service.network_scan.UtmNetworkScanService;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springdoc.api.annotations.ParameterObject;
import org.springframework.core.io.ByteArrayResource;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import javax.validation.constraints.NotEmpty;
import java.io.ByteArrayOutputStream;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmNetworkScan.
 */
@RestController
@RequestMapping("/api")
public class UtmNetworkScanResource {

    private final Logger log = LoggerFactory.getLogger(UtmNetworkScanResource.class);

    private static final String ENTITY_NAME = "utmNetworkScan";
    private static final String CLASSNAME = "UtmNetworkScanResource";

    private final UtmNetworkScanService networkScanService;
    private final ApplicationEventService eventService;
    private final UtmNetworkScanQueryService utmNetworkScanQueryService;

    public UtmNetworkScanResource(UtmNetworkScanService networkScanService,
                                  ApplicationEventService eventService,
                                  UtmNetworkScanQueryService utmNetworkScanQueryService) {
        this.networkScanService = networkScanService;
        this.eventService = eventService;
        this.utmNetworkScanQueryService = utmNetworkScanQueryService;
    }

    @PostMapping("/utm-network-scans/saveOrUpdateCustomAsset")
    public ResponseEntity<Void> saveOrUpdateCustomAsset(@Valid @RequestBody NetworkScanDTO asset) {
        final String ctx = CLASSNAME + ".saveAsset";

        try {
            networkScanService.saveOrUpdateCustomAsset(asset);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * PUT  /utm-network-scans : Updates an existing utmNetworkScan.
     *
     * @param utmNetworkScan the utmNetworkScan to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmNetworkScan,
     * or with status 400 (Bad Request) if the utmNetworkScan is not valid,
     * or with status 500 (Internal Server Error) if the utmNetworkScan couldn't be updated
     */
    @PutMapping("/utm-network-scans")
    public ResponseEntity<UtmNetworkScan> updateUtmNetworkScan(@Valid @RequestBody UtmNetworkScan utmNetworkScan) {
        final String ctx = CLASSNAME + ".updateUtmNetworkScan";
        if (utmNetworkScan.getId() == null)
            throw new BadRequestAlertException(ctx + ": Invalid id", ENTITY_NAME, "idnull");
        try {
            UtmNetworkScan result = networkScanService.save(utmNetworkScan);
            return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmNetworkScan.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PutMapping("/utm-network-scans/updateType")
    public ResponseEntity<Void> updateType(@Valid @RequestBody UpdateTypeRequestBody body) {
        final String ctx = CLASSNAME + ".updateType";
        try {
            networkScanService.updateType(body.getAssetsIds(), body.getAssetTypeId());
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PutMapping("/utm-network-scans/updateGroup")
    public ResponseEntity<Void> updateGroup(@Valid @RequestBody UpdateGroupRequestBody body) {
        final String ctx = CLASSNAME + ".updateGroup";
        try {
            networkScanService.updateGroup(body.getAssetsIds(), body.getAssetGroupId());
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-network-scans : get all the utmNetworkScans.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmNetworkScans in body
     */
    @GetMapping("/utm-network-scans")
    public ResponseEntity<List<UtmNetworkScan>> getAllUtmNetworkScans(UtmNetworkScanCriteria criteria, Pageable pageable) {
        log.debug("REST request to get UtmNetworkScans by criteria: {}", criteria);
        Page<UtmNetworkScan> page = utmNetworkScanQueryService.findByCriteria(criteria, pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-network-scans");
        return ResponseEntity.ok().headers(headers).body(page.getContent());
    }

    /**
     * GET  /utm-network-scans/count : count all the utmNetworkScans.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-network-scans/count")
    public ResponseEntity<Long> countUtmNetworkScans(UtmNetworkScanCriteria criteria) {
        log.debug("REST request to count UtmNetworkScans by criteria: {}", criteria);
        return ResponseEntity.ok().body(utmNetworkScanQueryService.countByCriteria(criteria));
    }

    /**
     * GET  /utm-network-scans/:id : get the "id" utmNetworkScan.
     *
     * @param id the id of the utmNetworkScan to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmNetworkScan, or with status 404 (Not Found)
     */
    @GetMapping("/utm-network-scans/{id}")
    public ResponseEntity<Optional<NetworkScanDTO>> getUtmNetworkScan(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getUtmNetworkScan";
        try {
            return ResponseEntity.ok(networkScanService.searchDetails(id));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/utm-network-scans/search-by-filters")
    public ResponseEntity<List<NetworkScanDTO>> searchByFilters(@ParameterObject NetworkScanFilter filters,
                                                                @ParameterObject Pageable pageable) {
        final String ctx = CLASSNAME + ".searchByFilters";
        try {
            Page<NetworkScanDTO> page = networkScanService.searchByFilters(filters, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-network-scans/search-by-filters");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/utm-network-scans/searchPropertyValues")
    public ResponseEntity<List<?>> searchPropertyValues(@RequestParam Property prop,
                                                        @RequestParam(required = false) String value,
                                                        @RequestParam Boolean forGroups,
                                                        Pageable pageable) {
        final String ctx = CLASSNAME + ".searchPropertyValues";
        try {
            return ResponseEntity.ok(networkScanService.searchPropertyValues(prop, value, forGroups, pageable));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/utm-network-scans/countNewAssets")
    public ResponseEntity<Integer> countNewAssets() {
        final String ctx = CLASSNAME + ".countNewAssets";
        try {
            Integer integer = networkScanService.countNewAssets();
            return ResponseEntity.ok(integer);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/utm-network-scans/getNetworkScanReport")
    public ResponseEntity<ByteArrayResource> getNetworkScanReport(NetworkScanFilter f, Pageable p) {
        final String ctx = CLASSNAME + ".getNetworkScanReport";
        try {
            ByteArrayOutputStream baos = networkScanService.getNetworkScanReport(f, p);

            if (baos != null)
                return ResponseEntity.ok().header(HttpHeaders.CONTENT_DISPOSITION,
                        "attachment;filename=report.pdf").contentType(
                        MediaType.APPLICATION_PDF).contentLength(baos.size())
                    .body(new ByteArrayResource(baos.toByteArray()));
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @DeleteMapping("/utm-network-scans/deleteCustomAsset/{id}")
    public ResponseEntity<Void> deleteCustomAsset(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteCustomAsset";
        try {
            networkScanService.deleteCustomAsset(id);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/utm-network-scans/can-run-command")
    public ResponseEntity<Boolean> canRunCommand(@RequestParam String assetName) {
        final String ctx = CLASSNAME + ".canRunCommand";
        try {
            return ResponseEntity.ok(networkScanService.canRunCommand(assetName));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/utm-network-scans/agent-os-platform")
    public ResponseEntity<List<String>> getAgentsOsPlatform() {
        final String ctx = CLASSNAME + ".getAgentsOsPlatform";
        try {
            return ResponseEntity.ok(networkScanService.getAgentsOsPlatform());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }


    public static class UpdateTypeRequestBody {
        @NotEmpty
        private List<Long> assetsIds;

        private Long assetTypeId;

        public List<Long> getAssetsIds() {
            return assetsIds;
        }

        public void setAssetsIds(List<Long> assetsIds) {
            this.assetsIds = assetsIds;
        }

        public Long getAssetTypeId() {
            return assetTypeId;
        }

        public void setAssetTypeId(Long assetTypeId) {
            this.assetTypeId = assetTypeId;
        }
    }

    public static class UpdateGroupRequestBody {
        @NotEmpty
        private List<Long> assetsIds;

        private Long assetGroupId;

        public List<Long> getAssetsIds() {
            return assetsIds;
        }

        public void setAssetsIds(List<Long> assetsIds) {
            this.assetsIds = assetsIds;
        }

        public Long getAssetGroupId() {
            return assetGroupId;
        }

        public void setAssetGroupId(Long assetGroupId) {
            this.assetGroupId = assetGroupId;
        }
    }
}
