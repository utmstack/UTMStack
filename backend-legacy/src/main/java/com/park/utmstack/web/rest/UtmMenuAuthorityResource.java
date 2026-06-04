package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmMenuAuthority;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.UtmMenuAuthorityService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.net.URI;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmMenuAuthority.
 */
@RestController
@RequestMapping("/api")
public class UtmMenuAuthorityResource {

    private final Logger log = LoggerFactory.getLogger(UtmMenuAuthorityResource.class);

    private static final String ENTITY_NAME = "utmMenuAuthority";
    private static final String CLASSNAME = "UtmHostTechnologyResource";

    private final UtmMenuAuthorityService utmMenuAuthorityService;
    private final ApplicationEventService applicationEventService;

    public UtmMenuAuthorityResource(UtmMenuAuthorityService utmMenuAuthorityService,
                                    ApplicationEventService applicationEventService) {
        this.utmMenuAuthorityService = utmMenuAuthorityService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-menu-authorities : Create a new utmMenuAuthority.
     *
     * @param utmMenuAuthority the utmMenuAuthority to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmMenuAuthority, or with status 400 (Bad Request) if the utmMenuAuthority has already an ID
     */
    @PostMapping("/utm-menu-authorities")
    public ResponseEntity<UtmMenuAuthority> createUtmMenuAuthority(
        @Valid @RequestBody UtmMenuAuthority utmMenuAuthority) {
        final String ctx = CLASSNAME + ".createUtmMenuAuthority";
        try {
            if (utmMenuAuthority.getId() != null)
                throw new BadRequestAlertException("A new utmMenuAuthority cannot already have an ID", ENTITY_NAME, "idexists");

            UtmMenuAuthority result = utmMenuAuthorityService.save(utmMenuAuthority);
            return ResponseEntity.created(new URI("/api/utm-menu-authorities/" + result.getId())).headers(
                HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString())).body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * PUT  /utm-menu-authorities : Updates an existing utmMenuAuthority.
     *
     * @param utmMenuAuthority the utmMenuAuthority to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmMenuAuthority,
     * or with status 400 (Bad Request) if the utmMenuAuthority is not valid,
     * or with status 500 (Internal Server Error) if the utmMenuAuthority couldn't be updated
     */
    @PutMapping("/utm-menu-authorities")
    public ResponseEntity<UtmMenuAuthority> updateUtmMenuAuthority(
        @Valid @RequestBody UtmMenuAuthority utmMenuAuthority) {
        final String ctx = CLASSNAME + ".updateUtmMenuAuthority";
        try {
            if (utmMenuAuthority.getId() == null)
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");

            UtmMenuAuthority result = utmMenuAuthorityService.save(utmMenuAuthority);
            return ResponseEntity.ok().headers(
                HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmMenuAuthority.getId().toString())).body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-menu-authorities : get all the utmMenuAuthorities.
     *
     * @return the ResponseEntity with status 200 (OK) and the list of utmMenuAuthorities in body
     */
    @GetMapping("/utm-menu-authorities")
    public List<UtmMenuAuthority> getAllUtmMenuAuthorities() {
        final String ctx = CLASSNAME + ".getAllUtmMenuAuthorities";
        try {
            return utmMenuAuthorityService.findAll();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    /**
     * GET  /utm-menu-authorities/:id : get the "id" utmMenuAuthority.
     *
     * @param id the id of the utmMenuAuthority to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmMenuAuthority, or with status 404 (Not Found)
     */
    @GetMapping("/utm-menu-authorities/{id}")
    public ResponseEntity<UtmMenuAuthority> getUtmMenuAuthority(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getUtmMenuAuthority";
        try {
            Optional<UtmMenuAuthority> utmMenuAuthority = utmMenuAuthorityService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmMenuAuthority);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-menu-authorities/:id : delete the "id" utmMenuAuthority.
     *
     * @param id the id of the utmMenuAuthority to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-menu-authorities/{id}")
    public ResponseEntity<Void> deleteUtmMenuAuthority(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteUtmMenuAuthority";
        try {
            utmMenuAuthorityService.delete(id);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }
}
