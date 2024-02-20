package com.park.utmstack.web.rest.logstash_pipeline;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashInput;
import com.park.utmstack.domain.logstash_pipeline.enums.InputPlugin;
import com.park.utmstack.domain.logstash_pipeline.factory.InputFactory;
import com.park.utmstack.domain.logstash_pipeline.types.InputConfiguration;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.logstash_pipeline.UtmLogstashInputConfigurationService;
import com.park.utmstack.service.logstash_pipeline.UtmLogstashInputService;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing {@link UtmLogstashInput}.
 */
@RestController
@RequestMapping("/api")
public class UtmLogstashInputResource {

    private static final String CLASSNAME = "UtmLogstashInputResource";

    private final Logger log = LoggerFactory.getLogger(UtmLogstashInputResource.class);

    private static final String ENTITY_NAME = "utmLogstashInput";

    private final UtmLogstashInputService utmLogstashInputService;
    private final ApplicationEventService applicationEventService;
    private final UtmLogstashInputConfigurationService utmLogstashInputConfigurationService;
    private final InputFactory inputFactory;
    private final List<String> allowedTypes = Arrays.asList(InputPlugin.TCP.name(),InputPlugin.HTTP.name());

    public UtmLogstashInputResource(
        UtmLogstashInputService utmLogstashInputService,
        ApplicationEventService applicationEventService,
        UtmLogstashInputConfigurationService utmLogstashInputConfigurationService, InputFactory inputFactory) {
        this.utmLogstashInputService = utmLogstashInputService;
        this.applicationEventService = applicationEventService;
        this.utmLogstashInputConfigurationService = utmLogstashInputConfigurationService;
        this.inputFactory = inputFactory;
    }

    /**
     * {@code PUT  /logstash-inputs} : Updates an existing utmLogstashInput.
     *
     * @param utmLogstashInput the utmLogstashInput to update.
     * @return the {@link ResponseEntity} with status {@code 200 (OK)} and with body the updated utmLogstashInput,
     * or with status {@code 400 (Bad Request)} if the utmLogstashInput is not valid,
     * or with status {@code 500 (Internal Server Error)} if the utmLogstashInput couldn't be updated.
     */
    @PutMapping("/logstash-inputs")
    public ResponseEntity<Void> update(@Valid @RequestBody UtmLogstashInput utmLogstashInput) {
        final String ctx = CLASSNAME + ".update";
        try {
            if (!utmLogstashInputService.findOne(utmLogstashInput.getId()).isPresent()) {
                throw new BadRequestAlertException("Input not found", "inputManagement", "not found");
            }
            utmLogstashInputService.update(utmLogstashInput);
            return ResponseEntity.noContent().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code GET  /logstash-inputs/types} : get allowed input types to create.
     *
     * @return the List of types allowed with status {@code 200 (OK)}
     */
    @GetMapping("/logstash-inputs/types")
    public ResponseEntity<List<String>> getAllowedInputTypes() {
        final String ctx = CLASSNAME + ".getAllowedInputTypes";
        try {
            return ResponseEntity.ok().body(allowedTypes);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code DELETE  /logstash-inputs/:id} : delete the "id" utmLogstashInput.
     *
     * @param id the id of the utmLogstashInput to delete.
     * @return the {@link ResponseEntity} with status {@code 204 (NO_CONTENT)}.
     */
    @DeleteMapping("/logstash-inputs/{id}")
    public ResponseEntity<Void> deleteUtmLogstashInput(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteLogstashInput";
        try {
            Optional<UtmLogstashInput> toDelete = utmLogstashInputService.findOne(id);
            if (toDelete.isPresent() && toDelete.get().getSystemOwner()) {
                throw new BadRequestAlertException("You can't delete system inputs", "inputManagement", "system input");
            }
            utmLogstashInputService.delete(id);
            return ResponseEntity.ok().headers(com.park.utmstack.web.rest.util.HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code GET  /logstash-inputs/available-value} : get available configuration values by type.
     *
     * @param type the type of configuration to search available values for.
     * @return the List of available configuration values by type, with status {@code 200 (OK)}.
     */
    @GetMapping("/logstash-inputs/available-value")
    public ResponseEntity<List<String>> getAvailableConfigValueByType(@RequestParam String type) {
        final String ctx = CLASSNAME + ".getAvailableConfigValueByType";
        try {
            if(!allowedTypes.stream().filter(allowed->allowed.compareToIgnoreCase(type)==0).findFirst().isPresent()){
                throw new Exception("Input type not supported");
            }
            List<String> configs = utmLogstashInputConfigurationService.getAvailableConfigsByType(type);
            if (configs.isEmpty()) {
                throw new Exception("No configuration value available for this type");
            }
            return ResponseEntity.ok().body(configs);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code GET  /logstash-inputs/config} : get a default input configuration structure by a given type.
     *
     * @param type the type used to construct the input configuration.
     * @return the  {@link ResponseEntity<InputConfiguration>}, with status {@code 200 (OK)}.
     */
    @GetMapping("/logstash-inputs/config")
    public ResponseEntity<InputConfiguration> getInputDefinitionByType(@RequestParam String type) {
        final String ctx = CLASSNAME + ".getInputDefinitionByType";
        try {
            if(!allowedTypes.stream().filter(allowed->allowed.compareToIgnoreCase(type)==0).findFirst().isPresent()){
                throw new Exception("Input type not supported");
            }
            InputConfiguration configs = inputFactory.getInputConfig(InputPlugin.getPluginEnum(type)).getConfiguration();
            return ResponseEntity.ok().body(configs);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}
