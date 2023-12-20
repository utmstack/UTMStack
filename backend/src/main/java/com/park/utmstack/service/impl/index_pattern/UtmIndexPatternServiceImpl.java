package com.park.utmstack.service.impl.index_pattern;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import com.park.utmstack.domain.index_pattern.enums.SystemIndexPattern;
import com.park.utmstack.repository.index_pattern.UtmIndexPatternRepository;
import com.park.utmstack.service.index_pattern.UtmIndexPatternService;
import com.park.utmstack.util.events.IndexPatternsReadyEvent;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.context.event.EventListener;
import org.springframework.core.annotation.Order;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;

import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * Service Implementation for managing UtmIndexPattern.
 */
@Service
@Transactional
public class UtmIndexPatternServiceImpl implements UtmIndexPatternService {

    private final Logger log = LoggerFactory.getLogger(UtmIndexPatternServiceImpl.class);
    private static final String CLASSNAME = "UtmIndexPatternServiceImpl";

    private final UtmIndexPatternRepository indexPatternRepository;
    private final ApplicationEventPublisher publisher;

    public UtmIndexPatternServiceImpl(UtmIndexPatternRepository utmIndexPatternRepository, ApplicationEventPublisher publisher) {
        this.indexPatternRepository = utmIndexPatternRepository;
        this.publisher = publisher;
    }

    @Order()
    @EventListener(ApplicationReadyEvent.class)
    public void init() {
        final String ctx = CLASSNAME + ".init";

        try {
            Map<Long, String> patterns = indexPatternRepository.findAll(Sort.by("id").ascending()).stream()
                .collect(Collectors.toMap(UtmIndexPattern::getId, UtmIndexPattern::getPattern));

            if (CollectionUtils.isEmpty(patterns))
                throw new Exception("There is no system index patterns configured");

            Constants.SYS_INDEX_PATTERN.put(SystemIndexPattern.LOGS, patterns.get(1L));
            Constants.SYS_INDEX_PATTERN.put(SystemIndexPattern.ALERTS, patterns.get(2L));
            Constants.SYS_INDEX_PATTERN.put(SystemIndexPattern.LOGS_WINDOWS, patterns.get(8L));

            publisher.publishEvent(new IndexPatternsReadyEvent(this));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }


    /**
     * Save a utmIndexPattern.
     *
     * @param utmIndexPattern the entity to save
     * @return the persisted entity
     */
    @Override
    public UtmIndexPattern save(UtmIndexPattern utmIndexPattern) {
        log.debug("Request to save UtmIndexPattern : {}", utmIndexPattern);
        return indexPatternRepository.save(utmIndexPattern);
    }

    @Override
    public void saveAll(List<UtmIndexPattern> patterns) {
        indexPatternRepository.saveAll(patterns);
    }

    /**
     * Get all the utmIndexPatterns.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Override
    @Transactional(readOnly = true)
    public Page<UtmIndexPattern> findAll(Pageable pageable) {
        log.debug("Request to get all UtmIndexPatterns");
        return indexPatternRepository.findAll(pageable);
    }

    @Override
    @Transactional(readOnly = true)
    public List<UtmIndexPattern> findAll() {
        return indexPatternRepository.findAll();
    }

    /**
     * Get one utmIndexPattern by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Override
    @Transactional(readOnly = true)
    public Optional<UtmIndexPattern> findOne(Long id) {
        log.debug("Request to get UtmIndexPattern : {}", id);
        return indexPatternRepository.findById(id);
    }

    /**
     * Delete the utmIndexPattern by id.
     *
     * @param id the id of the entity
     */
    @Override
    public void delete(Long id) {
        log.debug("Request to delete UtmIndexPattern : {}", id);
        indexPatternRepository.deleteById(id);
    }

    @Override
    public void deleteAllByPatternSystemIsTrueAndIdNotIn(List<Long> ids) {
        indexPatternRepository.deleteAllByPatternSystemIsTrueAndIdNotIn(ids);
    }

    public Long getSystemSequenceNextValue() {
        long value = 1;
        Optional<UtmIndexPattern> opt = indexPatternRepository.findFirstByPatternSystemIsTrueOrderByIdDesc();
        if (opt.isPresent())
            value = opt.get().getId() + 1;
        return value;
    }

    @Override
    public List<UtmIndexPattern> findAllByPatternModule(String nameShort) {
        return indexPatternRepository.findAllByPatternModule(nameShort);
    }
}
