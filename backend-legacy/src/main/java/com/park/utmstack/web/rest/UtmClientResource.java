package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmClient;
import com.park.utmstack.service.UtmClientService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import tech.jhipster.web.util.ResponseUtil;

import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmClient.
 */
@RestController
@RequestMapping("/api")
public class UtmClientResource {

    private final Logger log = LoggerFactory.getLogger(UtmClientResource.class);

    private static final String ENTITY_NAME = "utmClient";

    private final UtmClientService utmClientService;

    public UtmClientResource(UtmClientService utmClientService) {
        this.utmClientService = utmClientService;
    }

    /**
     * GET  /utm-clients : get all the utmClients.
     *
     * @return the ResponseEntity with status 200 (OK) and the list of utmClients in body
     */
    @GetMapping("/utm-clients")
    public List<UtmClient> getAllUtmClients() {
        log.debug("REST request to get all UtmClients");
        return utmClientService.findAll();
    }

    /**
     * GET  /utm-clients/:id : get the "id" utmClient.
     *
     * @param id the id of the utmClient to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmClient, or with status 404 (Not Found)
     */
    @GetMapping("/utm-clients/{id}")
    public ResponseEntity<UtmClient> getUtmClient(@PathVariable Long id) {
        log.debug("REST request to get UtmClient : {}", id);
        Optional<UtmClient> utmClient = utmClientService.findOne(id);
        return ResponseUtil.wrapOrNotFound(utmClient);
    }
}
