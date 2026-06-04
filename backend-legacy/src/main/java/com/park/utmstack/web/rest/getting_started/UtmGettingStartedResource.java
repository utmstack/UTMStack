package com.park.utmstack.web.rest.getting_started;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.getting_started.GettingStartedStepEnum;
import com.park.utmstack.domain.getting_started.UtmGettingStarted;
import com.park.utmstack.service.UtmStackService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.getting_started.GettingStartedService;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.vm.GettingStartedInit;
import io.swagger.v3.oas.annotations.parameters.RequestBody;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;

/**
 * REST controller for managing UtmIndexPattern.
 */
@RestController
@RequestMapping("/api")
public class UtmGettingStartedResource {

    private final Logger log = LoggerFactory.getLogger(UtmGettingStartedResource.class);

    private static final String ENTITY_NAME = "utmGettingStarted";
    private static final String CLASSNAME = "UtmGettingStartedResource";

    private final GettingStartedService gettingStartedService;
    private final ApplicationEventService applicationEventService;

    public UtmGettingStartedResource(GettingStartedService gettingStartedService,
                                     ApplicationEventService applicationEventService,
                                     UtmStackService utmStackService) {
        this.gettingStartedService = gettingStartedService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-getting-started : Initialize getting started steps.
     *
     * @param init if the instance is running in the SaaS or not
     * @return the ResponseEntity with status 201 (Created)
     */
    @PostMapping("/utm-getting-started/init")
    public ResponseEntity<String> initSteps(@Valid @RequestBody GettingStartedInit init) {
        final String ctx = CLASSNAME + ".initSteps";

        try {
            gettingStartedService.initializeSteps(init.isInSaas());

            return ResponseEntity.ok("initialized");
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    /**
     * POST  /utm-getting-started : Initialize getting started steps.
     *
     * @param step if the instance is running in the SaaS or not
     * @return the ResponseEntity with status 200 (completed)
     */
    @GetMapping("/utm-getting-started/complete")
    public ResponseEntity<String> completeStep(@RequestParam String step) {
        final String ctx = CLASSNAME + ".completeStep";

        try {
            gettingStartedService.completeStep(GettingStartedStepEnum.valueOf(step));

            return ResponseEntity.ok("complete");
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }


    /**
     * GET  /utm-getting-started : get all the utmIndexPatterns.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK)
     */
    @GetMapping("/utm-getting-started")
    public ResponseEntity<List<UtmGettingStarted>> getSteps(Pageable pageable) {
        final String ctx = CLASSNAME + ".getSteps";
        try {
            Page<UtmGettingStarted> page = gettingStartedService.getSteps(pageable);
            return ResponseEntity.ok().body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }


}
