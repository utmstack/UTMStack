package com.park.utmstack.service.network_scan;

import com.park.utmstack.domain.network_scan.UtmAssetTypes;
import com.park.utmstack.repository.network_scan.UtmAssetTypesRepository;
import com.park.utmstack.service.application_events.ApplicationEventService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/**
 * Service Implementation for managing UtmAssetTags.
 */
@Service
@Transactional
public class UtmAssetTypesService {

    private final Logger log = LoggerFactory.getLogger(UtmAssetTypesService.class);
    private static final String CLASSNAME = "UtmAssetTypesService";

    private final UtmAssetTypesRepository utmAssetTagsRepository;

    public UtmAssetTypesService(UtmAssetTypesRepository utmAssetTagsRepository,
                                ApplicationEventService eventService) {
        this.utmAssetTagsRepository = utmAssetTagsRepository;
    }

    /**
     * Get all the utm asset types.
     *
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmAssetTypes> findAll(Pageable pageable) throws Exception {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return utmAssetTagsRepository.findAll(pageable);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
