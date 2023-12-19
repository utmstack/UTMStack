package com.park.utmstack.service.impl.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmDashboard;
import com.park.utmstack.domain.chart_builder.UtmDashboardVisualization;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.repository.chart_builder.UtmDashboardRepository;
import com.park.utmstack.repository.chart_builder.UtmDashboardVisualizationRepository;
import com.park.utmstack.repository.chart_builder.UtmVisualizationRepository;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.UtmStackService;
import com.park.utmstack.service.chart_builder.UtmDashboardService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;

import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.*;

/**
 * Service Implementation for managing UtmDashboard.
 */
@Service
@Transactional
public class UtmDashboardServiceImpl implements UtmDashboardService {

    private static final String CLASS_NAME = "UtmDashboardServiceImpl";
    private final Logger log = LoggerFactory.getLogger(UtmDashboardServiceImpl.class);

    private final UtmDashboardRepository dashboardRepository;
    private final UtmVisualizationRepository visualizationRepository;
    private final UtmDashboardVisualizationRepository dashboardVisualizationRepository;
    private final UtmStackService utmStackService;

    public UtmDashboardServiceImpl(UtmDashboardRepository dashboardRepository,
                                   UtmVisualizationRepository visualizationRepository,
                                   UtmDashboardVisualizationRepository dashboardVisualizationRepository,
                                   UtmStackService utmStackService) {
        this.dashboardRepository = dashboardRepository;
        this.visualizationRepository = visualizationRepository;
        this.dashboardVisualizationRepository = dashboardVisualizationRepository;
        this.utmStackService = utmStackService;
    }

    /**
     * Save a utmDashboard.
     *
     * @param utmDashboard the entity to save
     * @return the persisted entity
     */
    @Override
    public UtmDashboard save(UtmDashboard utmDashboard) {
        log.debug("Request to save UtmDashboard : {}", utmDashboard);
        return dashboardRepository.save(utmDashboard);
    }

    @Override
    public void saveAll(List<UtmDashboard> utmDashboard) {
        dashboardRepository.saveAll(utmDashboard);
    }

    /**
     * Get all the utmDashboards.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Override
    @Transactional(readOnly = true)
    public Page<UtmDashboard> findAll(Pageable pageable) {
        log.debug("Request to get all UtmDashboards");
        return dashboardRepository.findAll(pageable);
    }


    /**
     * Get one utmDashboard by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Override
    @Transactional(readOnly = true)
    public Optional<UtmDashboard> findOne(Long id) {
        log.debug("Request to get UtmDashboard : {}", id);
        return dashboardRepository.findById(id);
    }

    /**
     * Delete the utmDashboard by id.
     *
     * @param id the id of the entity
     */
    @Override
    public void delete(Long id) {
        log.debug("Request to delete UtmDashboard : {}", id);
        dashboardRepository.deleteById(id);
    }

    @Override
    public void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> ids) {
        dashboardRepository.deleteAllBySystemOwnerIsTrueAndIdNotIn(ids);
    }

    @Override
    @Transactional
    public void importDashboards(List<UtmDashboardVisualization> dashboards, Boolean override) throws Exception {
        final String ctx = CLASS_NAME + ".importDashboards";

        try {
            if (CollectionUtils.isEmpty(dashboards))
                return;

            Map<Long, Long> dashboardIds = new HashMap<>();
            Map<Long, Long> visualizationIds = new HashMap<>();

            UtmDashboard utmDashboard;
            UtmVisualization utmVisualization;
            boolean inDevelop = utmStackService.isInDevelop();

            for (UtmDashboardVisualization dashboard : dashboards) {
                dashboard.setId(null);

                Objects.requireNonNull(dashboard.getDashboard(), "A dashboard info is missing");
                Objects.requireNonNull(dashboard.getVisualization(), "A visualization info is missing");

                // Inserting visualizations
                if (!visualizationIds.containsKey(dashboard.getIdVisualization())) {
                    utmVisualization = dashboard.getVisualization();
                    Optional<UtmVisualization> dbVisualization = visualizationRepository.findByName(utmVisualization.getName());

                    if (dbVisualization.isPresent()) {
                        if (override) {
                            utmVisualization.setId(dbVisualization.get().getId());
                            utmVisualization.setModifiedDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                            utmVisualization.setSystemOwner(inDevelop);
                            utmVisualization = visualizationRepository.save(utmVisualization);
                        } else {
                            utmVisualization = dbVisualization.get();
                        }
                    } else {
                        utmVisualization.setId(inDevelop ? getSystemSequenceNextValue() : null);
                        utmVisualization.setSystemOwner(inDevelop);
                        utmVisualization.setCreatedDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                        utmVisualization.setUserCreated(SecurityUtils.getCurrentUserLogin().orElse("system"));
                        utmVisualization = visualizationRepository.save(utmVisualization);
                    }
                    visualizationIds.put(dashboard.getIdVisualization(), utmVisualization.getId());
                }

                // Inserting dashboards
                if (!dashboardIds.containsKey(dashboard.getDashboard().getId())) {
                    utmDashboard = dashboard.getDashboard();
                    Optional<UtmDashboard> dbDashboard = dashboardRepository.findByName(utmDashboard.getName());

                    if (dbDashboard.isPresent()) {
                        if (override) {
                            utmDashboard.setId(dbDashboard.get().getId());
                            utmDashboard.setModifiedDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                            utmDashboard = dashboardRepository.save(utmDashboard);
                        } else {
                            utmDashboard = dbDashboard.get();
                        }
                    } else {
                        utmDashboard.setId(inDevelop ? getSystemSequenceNextValue() : null);
                        utmDashboard.setSystemOwner(inDevelop);
                        utmDashboard.setCreatedDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                        utmDashboard.setUserCreated(SecurityUtils.getCurrentUserLogin().orElse("system"));
                        utmDashboard = dashboardRepository.save(utmDashboard);
                    }
                    dashboardIds.put(dashboard.getIdDashboard(), utmDashboard.getId());
                }

                Long idDashboard = dashboardIds.get(dashboard.getIdDashboard());
                Long idVisualization = visualizationIds.get(dashboard.getIdVisualization());
                Optional<UtmDashboardVisualization> opt = dashboardVisualizationRepository.findByIdDashboardAndIdVisualization(idDashboard, idVisualization);

                if (!opt.isPresent()) {
                    dashboard.setIdDashboard(dashboardIds.get(dashboard.getIdDashboard()));
                    dashboard.setIdVisualization(visualizationIds.get(dashboard.getIdVisualization()));
                    dashboardVisualizationRepository.save(dashboard);
                }
            }
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    @Override
    public Long getSystemSequenceNextValue() {
        long value = 1;
        Optional<UtmDashboard> opt = dashboardRepository.findFirstBySystemOwnerIsTrueOrderByIdDesc();
        if (opt.isPresent())
            value = opt.get().getId() + 1;
        return value;
    }
}
