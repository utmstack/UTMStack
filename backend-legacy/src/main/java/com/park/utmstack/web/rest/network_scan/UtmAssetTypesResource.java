package com.park.utmstack.web.rest.network_scan;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.network_scan.UtmAssetTypes;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.network_scan.UtmAssetTypesService;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

/**
 * REST controller for managing UtmAssetTags.
 */
@RestController
@RequestMapping("/api")
public class UtmAssetTypesResource {

    private final Logger log = LoggerFactory.getLogger(UtmAssetTypesResource.class);

    private static final String CLASSNAME = "UtmAssetTypesResource";
    private static final String ENTITY_NAME = "UtmAssetTypes";

    private final UtmAssetTypesService utmAssetTagsService;
    private final ApplicationEventService eventService;

    public UtmAssetTypesResource(UtmAssetTypesService utmAssetTagsService,
                                 ApplicationEventService eventService) {
        this.utmAssetTagsService = utmAssetTagsService;
        this.eventService = eventService;
    }

    /**
     * GET  /utm-asset-tags : get all the utmAssetTags.
     *
     * @return the ResponseEntity with status 200 (OK) and the list of utmAssetTags in body
     */
    @GetMapping("/utm-asset-types/all")
    public ResponseEntity<List<UtmAssetTypes>> getAssetTypes(Pageable pageable) {
        final String ctx = CLASSNAME + ".getAssetTypes";
        try {
            Page<UtmAssetTypes> page = utmAssetTagsService.findAll(pageable);
            return ResponseEntity.ok(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }
}
