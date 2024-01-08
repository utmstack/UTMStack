package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmConfigurationSection;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.UtmConfigurationSectionQueryService;
import com.park.utmstack.service.UtmConfigurationSectionService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmConfigurationSectionCriteria;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springdoc.api.annotations.ParameterObject;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmConfigurationSection.
 */
@RestController
@RequestMapping("/api")
public class UtmConfigurationSectionResource {
    private static final String CLASSNAME = "UtmConfigurationSectionResource";

    private final Logger log = LoggerFactory.getLogger(UtmConfigurationSectionResource.class);

    private static final String ENTITY_NAME = "utmConfigurationSection";

    private final UtmConfigurationSectionService configurationSectionService;
    private final UtmConfigurationSectionQueryService configurationSectionQueryService;
    private final ApplicationEventService applicationEventService;

    public UtmConfigurationSectionResource(UtmConfigurationSectionService configurationSectionService,
                                           UtmConfigurationSectionQueryService configurationSectionQueryService,
                                           ApplicationEventService applicationEventService) {
        this.configurationSectionService = configurationSectionService;
        this.configurationSectionQueryService = configurationSectionQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-configuration-sections : Create a new utmConfigurationSection.
     *
     * @param utmConfigurationSection the utmConfigurationSection to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmConfigurationSection, or with status 400 (Bad Request) if the utmConfigurationSection has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-configuration-sections")
    public ResponseEntity<UtmConfigurationSection> createUtmConfigurationSection(@RequestBody UtmConfigurationSection utmConfigurationSection) throws URISyntaxException {
        log.debug("REST request to save UtmConfigurationSection : {}", utmConfigurationSection);
        if (utmConfigurationSection.getId() != null) {
            throw new BadRequestAlertException("A new utmConfigurationSection cannot already have an ID", ENTITY_NAME, "idexists");
        }
        UtmConfigurationSection result = configurationSectionService.save(utmConfigurationSection);
        return ResponseEntity.created(new URI("/api/utm-configuration-sections/" + result.getId()))
                .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
                .body(result);
    }

    /**
     * PUT  /utm-configuration-sections : Updates an existing utmConfigurationSection.
     *
     * @param utmConfigurationSection the utmConfigurationSection to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmConfigurationSection,
     * or with status 400 (Bad Request) if the utmConfigurationSection is not valid,
     * or with status 500 (Internal Server Error) if the utmConfigurationSection couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-configuration-sections")
    public ResponseEntity<UtmConfigurationSection> updateUtmConfigurationSection(@RequestBody UtmConfigurationSection utmConfigurationSection) throws URISyntaxException {
        log.debug("REST request to update UtmConfigurationSection : {}", utmConfigurationSection);
        if (utmConfigurationSection.getId() == null) {
            throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
        }
        UtmConfigurationSection result = configurationSectionService.save(utmConfigurationSection);
        return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmConfigurationSection.getId().toString()))
                .body(result);
    }

    /**
     * GET  /utm-configuration-sections : get all the utmConfigurationSections.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmConfigurationSections in body
     */
    @GetMapping("/utm-configuration-sections")
    public ResponseEntity<List<UtmConfigurationSection>> getConfigurationSections(@ParameterObject UtmConfigurationSectionCriteria criteria,
                                                                                  @ParameterObject Pageable pageable) {
        final String ctx = CLASSNAME + ".getConfigurationSections";
        try {
            Page<UtmConfigurationSection> page = configurationSectionQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-configuration-sections");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            final String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildInternalServerErrorResponse(msg);
        }
    }

    /**
     * GET  /utm-configuration-sections/count : count all the utmConfigurationSections.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-configuration-sections/count")
    public ResponseEntity<Long> countUtmConfigurationSections(UtmConfigurationSectionCriteria criteria) {
        log.debug("REST request to count UtmConfigurationSections by criteria: {}", criteria);
        return ResponseEntity.ok().body(configurationSectionQueryService.countByCriteria(criteria));
    }

    /**
     * GET  /utm-configuration-sections/:id : get the "id" utmConfigurationSection.
     *
     * @param id the id of the utmConfigurationSection to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmConfigurationSection, or with status 404 (Not Found)
     */
    @GetMapping("/utm-configuration-sections/{id}")
    public ResponseEntity<UtmConfigurationSection> getUtmConfigurationSection(@PathVariable Long id) {
        log.debug("REST request to get UtmConfigurationSection : {}", id);
        Optional<UtmConfigurationSection> utmConfigurationSection = configurationSectionService.findOne(id);
        return ResponseUtil.wrapOrNotFound(utmConfigurationSection);
    }

    /**
     * DELETE  /utm-configuration-sections/:id : delete the "id" utmConfigurationSection.
     *
     * @param id the id of the utmConfigurationSection to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-configuration-sections/{id}")
    public ResponseEntity<Void> deleteUtmConfigurationSection(@PathVariable Long id) {
        log.debug("REST request to delete UtmConfigurationSection : {}", id);
        configurationSectionService.delete(id);
        return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
    }
}
