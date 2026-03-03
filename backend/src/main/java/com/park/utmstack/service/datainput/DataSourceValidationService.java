package com.park.utmstack.service.datainput;

import com.park.utmstack.domain.datainput_ingestion.UtmDataInputStatus;
import com.park.utmstack.domain.UtmServerModule;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.repository.datainput_ingestion.UtmDataInputStatusRepository;
import com.park.utmstack.service.UtmServerModuleService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;

import java.util.Arrays;
import java.util.List;
import java.util.concurrent.TimeUnit;


@Service
@RequiredArgsConstructor
@Transactional
public class DataSourceValidationService {

    private static final String CLASSNAME = "DataSourceValidationService";
    private final Logger log = LoggerFactory.getLogger(DataSourceValidationService.class);

    private final UtmDataInputStatusRepository dataInputStatusRepository;
    private final UtmServerModuleService serverModuleService;
    private final ApplicationEventService applicationEventService;

    private static final long TIMEOUT_SECONDS = 3600L;


    @Scheduled(fixedDelay = 900000, initialDelay = 60000)
    public void validateCriticalDataSources() {
        final String ctx = CLASSNAME + ".validateCriticalDataSources";
        try {
            log.debug("{}: Starting validation of critical data sources", ctx);

            // Map source types to server module names
            validateDataSource("aws", "aws");
            validateDataSource("o365", "office365");
            validateDataSource("hids", "transporter");

            log.debug("{}: Completed validation of critical data sources", ctx);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg, e);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    /**
     * Validates a single data source type and marks modules for restart if needed.
     *
     * @param dataType     the data type to validate (e.g., "aws", "o365")
     * @param serverModule the corresponding server module name (e.g., "aws", "office365")
     */
    private void validateDataSource(String dataType, String serverModule) {
        final String ctx = CLASSNAME + ".validateDataSource";
        try {
            log.debug("{}: Validating data source type: {}", ctx, dataType);

            // Fetch all entries for this data type
            List<UtmDataInputStatus> dataInputs = dataInputStatusRepository.findAllByDataTypeIn(List.of(dataType));

            if (CollectionUtils.isEmpty(dataInputs)) {
                log.debug("{}: No data inputs found for type: {}", ctx, dataType);
                return;
            }

            // Check if any input has sent data within the timeout window
            long currentTimeInSeconds = TimeUnit.MILLISECONDS.toSeconds(System.currentTimeMillis());
            boolean hasRecentData = dataInputs.stream()
                    .anyMatch(input -> (currentTimeInSeconds - input.getTimestamp()) < TIMEOUT_SECONDS);

            if (hasRecentData) {
                log.debug("{}: Data source {} is healthy and sending logs", ctx, dataType);
                return;
            }

            // If no recent data, mark the module for restart
            log.warn("{}: No recent data from {} sources, marking module {} for restart", ctx, dataType, serverModule);
            markModuleForRestart(serverModule);

        } catch (Exception e) {
            log.error("{}: Error validating data source {} - {}", ctx, dataType, e.getMessage(), e);
            applicationEventService.createEvent(
                    ctx + ": Error validating data source " + dataType,
                    ApplicationEventType.ERROR);
        }
    }

    /**
     * Marks all instances of a server module as needing restart.
     *
     * @param serverModule the module name to mark for restart
     */
    private void markModuleForRestart(String serverModule) {
        final String ctx = CLASSNAME + ".markModuleForRestart";
        try {
            List<UtmServerModule> modules = serverModuleService.findAllByModuleName(serverModule);

            if (CollectionUtils.isEmpty(modules)) {
                log.warn("{}: No modules found with name: {}", ctx, serverModule);
                return;
            }

            modules.forEach(module -> module.setNeedsRestart(true));
            serverModuleService.saveAll(modules);

            log.info("{}: Marked {} module instance(s) for restart", ctx, modules.size());
            applicationEventService.createEvent(
                    "Data source module '" + serverModule + "' marked for restart due to inactivity",
                    ApplicationEventType.WARNING);

        } catch (Exception e) {
            log.error("{}: Error marking module {} for restart - {}", ctx, serverModule, e.getMessage(), e);
            applicationEventService.createEvent(
                    ctx + ": Error marking module for restart",
                    ApplicationEventType.ERROR);
        }
    }
}

