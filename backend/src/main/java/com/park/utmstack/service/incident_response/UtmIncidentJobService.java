package com.park.utmstack.service.incident_response;

import com.park.utmstack.domain.incident.enums.IncidentHistoryActionEnum;
import com.park.utmstack.domain.incident_response.UtmIncidentAction;
import com.park.utmstack.domain.incident_response.UtmIncidentJob;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.repository.incident_response.UtmIncidentActionRepository;
import com.park.utmstack.repository.incident_response.UtmIncidentJobRepository;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.incident.UtmIncidentHistoryService;
import com.park.utmstack.service.network_scan.UtmNetworkScanService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.Assert;
import org.springframework.util.StringUtils;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmIncidentJob.
 */
@Service
@Transactional
public class UtmIncidentJobService {
    private static final String CLASSNAME = "UtmIncidentJobService";

    private final Logger log = LoggerFactory.getLogger(UtmIncidentJobService.class);

    private final UtmIncidentActionRepository utmIncidentActionRepository;
    private final UtmIncidentJobRepository incidentJobRepository;
    private final UtmIncidentActionCommandService incidentActionCommandService;
    private final UtmNetworkScanService networkScanService;

    private final UtmIncidentHistoryService utmIncidentHistoryService;

    public UtmIncidentJobService(UtmIncidentActionRepository utmIncidentActionRepository, UtmIncidentJobRepository incidentJobRepository,
                                 UtmIncidentActionCommandService incidentActionCommandService,
                                 UtmNetworkScanService networkScanService, UtmIncidentHistoryService utmIncidentHistoryService) {
        this.utmIncidentActionRepository = utmIncidentActionRepository;
        this.incidentJobRepository = incidentJobRepository;
        this.incidentActionCommandService = incidentActionCommandService;
        this.networkScanService = networkScanService;
        this.utmIncidentHistoryService = utmIncidentHistoryService;
    }

    /**
     * Save a incidentJob.
     *
     * @param incidentJob the entity to save
     * @return the persisted entity
     */
    public UtmIncidentJob save(UtmIncidentJob incidentJob) throws Exception {
        final String ctx = CLASSNAME + ".save";
        try {
            UtmNetworkScan asset = networkScanService.findByNameOrIp(incidentJob.getAgent())
                .orElseThrow(() -> new RuntimeException(String.format("Asset %1$s not found", incidentJob.getAgent())));
            incidentJob.setCreatedDate(Instant.now());
            incidentJob.setCreatedUser(SecurityUtils.getCurrentUserLogin().orElse("system"));
            incidentActionCommandService.getCommand(asset.getAssetOsPlatform(), incidentJob.getActionId())
                .ifPresent(cmd -> {
                    String params = incidentJob.getParams();
                    incidentJob.setParams(StringUtils.hasText(params) ? cmd + " " + params : cmd);
                });

            UtmIncidentJob job = incidentJobRepository.save(incidentJob);

            if (job.getOriginType().equalsIgnoreCase("INCIDENT")) {

                UtmIncidentAction action = utmIncidentActionRepository.findById(incidentJob.getActionId()).orElse(null);

                String msg = String.format("Command executed over agent %s", job.getAgent());
                if (action != null) {
                    msg += String.format(", command %s", action.getActionCommand() + " " + job.getParams());
                }

                utmIncidentHistoryService.createHistory(IncidentHistoryActionEnum.INCIDENT_COMMAND_EXECUTED,
                    incidentJob.getOriginId().longValue(), "Incident command executed", msg);
            }
            return job;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public void saveAll(List<UtmIncidentJob> jobs) {
        incidentJobRepository.saveAll(jobs);
    }

    /**
     * Get all the utmIncidentJobs.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentJob> findAll(Pageable pageable) {
        log.debug("Request to get all UtmIncidentJobs");
        return incidentJobRepository.findAll(pageable);
    }


    /**
     * Get one utmIncidentJob by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmIncidentJob> findOne(Long id) {
        log.debug("Request to get UtmIncidentJob : {}", id);
        return incidentJobRepository.findById(id);
    }

    /**
     * Delete the utmIncidentJob by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmIncidentJob : {}", id);
        incidentJobRepository.deleteById(id);
    }

    public List<UtmIncidentJob> findAllByAgent(String host) throws Exception {
        final String ctx = CLASSNAME + ".findAllByAgent";
        try {
            Assert.hasText(host, "Parameter [host] is null or empty");
            return incidentJobRepository.findAllByAgent(host);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
