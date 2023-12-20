package com.park.utmstack.web.rest.incident;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident.UtmIncidentNote;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.incident.NewIncidentNoteDTO;
import com.park.utmstack.service.dto.incident.UtmIncidentNoteCriteria;
import com.park.utmstack.service.incident.UtmIncidentNoteQueryService;
import com.park.utmstack.service.incident.UtmIncidentNoteService;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;

/**
 * REST controller for managing UtmIncidentNote.
 */
@RestController
@RequestMapping("/api")
public class UtmIncidentNoteResource {
    private final String CLASS_NAME = "UtmIncidentNoteResource";
    private final Logger log = LoggerFactory.getLogger(UtmIncidentNoteResource.class);

    private static final String ENTITY_NAME = "utmIncidentNote";

    private final UtmIncidentNoteService utmIncidentNoteService;

    private final UtmIncidentNoteQueryService utmIncidentNoteQueryService;

    private final ApplicationEventService applicationEventService;

    public UtmIncidentNoteResource(UtmIncidentNoteService utmIncidentNoteService,
                                   UtmIncidentNoteQueryService utmIncidentNoteQueryService,
                                   ApplicationEventService applicationEventService) {
        this.utmIncidentNoteService = utmIncidentNoteService;
        this.utmIncidentNoteQueryService = utmIncidentNoteQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-incident-notes : Create a new utmIncidentNote.
     *
     * @param newIncidentNoteDTO the utmIncidentNote to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmIncidentNote, or with status 400 (Bad Request) if the utmIncidentNote has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-incident-notes")
    public ResponseEntity<UtmIncidentNote> createUtmIncidentNote(@Valid @RequestBody NewIncidentNoteDTO newIncidentNoteDTO) throws URISyntaxException {
        final String ctx = ".createUtmIncident";
        try {
            log.debug("REST request to save UtmIncidentNote : {}", newIncidentNoteDTO);
            if (newIncidentNoteDTO.getIncidentId() == null) {
                throw new BadRequestAlertException("A new utmIncidentNote need an incident ID", ENTITY_NAME, "idexists");
            }
            UtmIncidentNote result = utmIncidentNoteService.save(newIncidentNoteDTO, false);
            return ResponseEntity.created(new URI("/api/utm-incident-notes/" + result.getId()))
                .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-incident-notes : get all the utmIncidentNotes.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmIncidentNotes in body
     */
    @GetMapping("/utm-incident-notes")
    public ResponseEntity<List<UtmIncidentNote>> getAllUtmIncidentNotes(UtmIncidentNoteCriteria criteria, Pageable pageable) {
        final String ctx = ".createUtmIncident";
        try {
            log.debug("REST request to get UtmIncidentNotes by criteria: {}", criteria);
            Page<UtmIncidentNote> page = utmIncidentNoteQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-incident-notes");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }

    }
}
