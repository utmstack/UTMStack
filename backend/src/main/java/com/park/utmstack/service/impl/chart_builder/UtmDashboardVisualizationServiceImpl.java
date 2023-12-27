package com.park.utmstack.service.impl.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmDashboardVisualization;
import com.park.utmstack.repository.chart_builder.UtmDashboardVisualizationRepository;
import com.park.utmstack.service.chart_builder.UtmDashboardVisualizationService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * Service Implementation for managing UtmDashboardVisualization.
 */
@Service
@Transactional
public class UtmDashboardVisualizationServiceImpl implements UtmDashboardVisualizationService {

    private final Logger log = LoggerFactory.getLogger(UtmDashboardVisualizationServiceImpl.class);

    private final UtmDashboardVisualizationRepository utmDashboardVisualizationRepository;

    public UtmDashboardVisualizationServiceImpl(UtmDashboardVisualizationRepository utmDashboardVisualizationRepository) {
        this.utmDashboardVisualizationRepository = utmDashboardVisualizationRepository;
    }

    /**
     * Save a utmDashboardVisualization.
     *
     * @param utmDashboardVisualization the entity to save
     * @return the persisted entity
     */
    @Override
    public UtmDashboardVisualization save(UtmDashboardVisualization utmDashboardVisualization) {
        log.debug("Request to save UtmDashboardVisualization : {}", utmDashboardVisualization);
        return utmDashboardVisualizationRepository.save(utmDashboardVisualization);
    }

    @Override
    public void saveAll(List<UtmDashboardVisualization> relations) {
        relations = relations.stream().filter(rel -> !utmDashboardVisualizationRepository.findByIdDashboardAndIdVisualization(rel.getIdDashboard(), rel.getIdVisualization()).isPresent()).collect(Collectors.toList());
        utmDashboardVisualizationRepository.saveAll(relations);
    }

    /**
     * Get all the utmDashboardVisualizations.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Override
    @Transactional(readOnly = true)
    public Page<UtmDashboardVisualization> findAll(Pageable pageable) {
        log.debug("Request to get all UtmDashboardVisualizations");
        return utmDashboardVisualizationRepository.findAll(pageable);
    }


    /**
     * Get one utmDashboardVisualization by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Override
    @Transactional(readOnly = true)
    public Optional<UtmDashboardVisualization> findOne(Long id) {
        log.debug("Request to get UtmDashboardVisualization : {}", id);
        return utmDashboardVisualizationRepository.findById(id);
    }

    /**
     * Delete the utmDashboardVisualization by id.
     *
     * @param id the id of the entity
     */
    @Override
    public void delete(Long id) {
        log.debug("Request to delete UtmDashboardVisualization : {}", id);
        utmDashboardVisualizationRepository.deleteById(id);
    }

    @Override
    public Optional<List<UtmDashboardVisualization>> findAllByIdDashboard(Long idDashboard) {
        return utmDashboardVisualizationRepository.findAllByIdDashboard(idDashboard);
    }
}
