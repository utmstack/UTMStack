package com.park.utmstack.web.rest;

import com.park.utmstack.domain.Authority;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.AuthorityService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;

@RestController
@RequestMapping("/api")
public class AuthorityResource {
    private static final String CLASSNAME = "AuthorityResource";
    private final Logger log = LoggerFactory.getLogger(AuthorityResource.class);

    private static final String ENTITY_NAME = "jhi_authority";

    private final AuthorityService authorityService;
    private final ApplicationEventService applicationEventService;

    public AuthorityResource(AuthorityService authorityService,
                             ApplicationEventService applicationEventService) {
        this.authorityService = authorityService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /authority : Create a new authority.
     *
     * @param authority the authority to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmRule, or with status 400 (Bad Request) if the utmRule has already an ID
     */
    @PostMapping("/authority")
    public ResponseEntity<Authority> createAuthority(@Valid @RequestBody Authority authority) {
        final String ctx = CLASSNAME + ".createAuthority";
        try {
            Authority result = authorityService.save(authority);
            return ResponseEntity.created(new URI("/api/authority/" + result.getName())).headers(
                HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getName())).body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * PUT  /authority : Updates an existing authority.
     *
     * @param authority the authority to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmRule,
     * or with status 400 (Bad Request) if the utmRule is not valid,
     * or with status 500 (Internal Server Error) if the utmRule couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/authority")
    public ResponseEntity<Authority> updateAuthority(@Valid @RequestBody Authority authority) throws Exception {
        final String ctx = CLASSNAME + ".updateAuthority";

        try {
            if (authority.getName() == null)
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "name null");
            Authority result = authorityService.save(authority);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, result.getName())).body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /authority : get all the utmRules.
     *
     * @return the ResponseEntity with status 200 (OK) and the list of utmRules in body
     */
    @GetMapping("/authority")
    public ResponseEntity<List<Authority>> getAllAuthority() {
        final String ctx = CLASSNAME + ".getAllAuthority";
        try {
            List<Authority> authorities = authorityService.findAll();
            return ResponseEntity.ok(authorities);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * DELETE  /authority/:name : delete the "name" authority.
     *
     * @param name the id of the utmRule to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/authority/{name}")
    public ResponseEntity<Void> deleteAuthority(@PathVariable String name) {
        final String ctx = CLASSNAME + ".deleteAuthority";
        try {
            authorityService.delete(name);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, name)).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}
