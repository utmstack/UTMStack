package com.park.utmstack.service.impl.chart_builder;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.repository.chart_builder.UtmVisualizationRepository;
import com.park.utmstack.service.chart_builder.UtmVisualizationService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmVisualization.
 */
@Service
@Transactional
public class UtmVisualizationServiceImpl implements UtmVisualizationService {
    private static final String CLASSNAME = "UtmVisualizationServiceImpl";

    private final Logger log = LoggerFactory.getLogger(UtmVisualizationServiceImpl.class);

    private final UtmVisualizationRepository utmVisualizationRepository;

    public UtmVisualizationServiceImpl(UtmVisualizationRepository utmVisualizationRepository) {
        this.utmVisualizationRepository = utmVisualizationRepository;
    }

    @EventListener(ApplicationReadyEvent.class)
    public void init() {
        final String ctx = CLASSNAME + ".init";
        try {
            List<UtmVisualization> visualizations = utmVisualizationRepository.findAllByChTypeIn(List.of("AREA_CHART", "VERTICAL_BAR_CHART", "PIE_CHART", "HORIZONTAL_BAR_CHART", "LINE_CHART"));
            for (UtmVisualization visualization : visualizations) {
                if (!StringUtils.hasText(visualization.getChartConfig()))
                    continue;

                ObjectMapper om = new ObjectMapper();
                JsonNode jsonNode = om.readTree(visualization.getChartConfig());

                switch (visualization.getChartType()) {
                    case HORIZONTAL_BAR_CHART:
                    case AREA_CHART:
                    case VERTICAL_BAR_CHART:
                    case LINE_CHART:
                        if (jsonNode.has("legend") && jsonNode.get("legend").has("data"))
                            ((ObjectNode) jsonNode.get("legend")).putArray("data");
                        if (jsonNode.has("xAxis") && jsonNode.get("xAxis").has("data"))
                            ((ObjectNode) jsonNode.get("xAxis")).putArray("data");
                        if (jsonNode.has("yAxis") && jsonNode.get("yAxis").has("data"))
                            ((ObjectNode) jsonNode.get("yAxis")).putArray("data");
                        if (jsonNode.has("series"))
                            ((ObjectNode) jsonNode).putArray("series");
                        break;
                    case PIE_CHART:
                        if (jsonNode.has("legend") && jsonNode.get("legend").has("data"))
                            ((ObjectNode) jsonNode.get("legend")).putArray("data");
                        break;
                }
                visualization.setChartConfig(om.writeValueAsString(jsonNode));
            }
            utmVisualizationRepository.saveAllAndFlush(visualizations);
        } catch (Exception e) {
            log.error(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Save a utmVisualization.
     *
     * @param utmVisualization the entity to save
     * @return the persisted entity
     */
    @Override
    public UtmVisualization save(UtmVisualization utmVisualization) {
        log.debug("Request to save UtmVisualization : {}", utmVisualization);
        return utmVisualizationRepository.save(utmVisualization);
    }

    @Override
    public void saveAll(Iterable<UtmVisualization> visualizations) {
        log.debug("Request to save batch of visualizations : {}", visualizations);
        utmVisualizationRepository.saveAll(visualizations);
    }

    /**
     * Get all the utmVisualizations.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Override
    @Transactional(readOnly = true)
    public Page<UtmVisualization> findAll(Pageable pageable) {
        log.debug("Request to get all UtmVisualizations");
        return utmVisualizationRepository.findAll(pageable);
    }


    /**
     * Get one utmVisualization by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Override
    @Transactional(readOnly = true)
    public Optional<UtmVisualization> findOne(Long id) {
        log.debug("Request to get UtmVisualization : {}", id);
        return utmVisualizationRepository.findById(id);
    }

    /**
     * Delete the utmVisualization by id.
     *
     * @param id the id of the entity
     */
    @Override
    public void delete(Long id) {
        log.debug("Request to delete UtmVisualization : {}", id);
        utmVisualizationRepository.deleteById(id);
    }

    @Override
    public void deleteByIdIn(List<Long> ids) {
        utmVisualizationRepository.deleteByIdIn(ids);
    }

    @Override
    public void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> ids) {
        utmVisualizationRepository.deleteAllBySystemOwnerIsTrueAndIdNotIn(ids);
    }

    @Override
    public Optional<UtmVisualization> findByName(String name) {
        return utmVisualizationRepository.findByName(name);
    }

    @Override
    public Long getSystemSequenceNextValue() {
        long value = 1;
        Optional<UtmVisualization> opt = utmVisualizationRepository.findFirstBySystemOwnerIsTrueOrderByIdDesc();
        if (opt.isPresent())
            value = opt.get().getId() + 1;
        return value;
    }
}
