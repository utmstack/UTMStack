package com.park.utmstack.service;

import com.park.utmstack.domain.UtmImages;
import com.park.utmstack.domain.shared_types.enums.ImageShortName;
import com.park.utmstack.repository.UtmImagesRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmImages.
 */
@Service
@Transactional
public class UtmImagesService {

    private static final String CLASSNAME = "UtmImagesService";

    private final Logger log = LoggerFactory.getLogger(UtmImagesService.class);

    private final UtmImagesRepository utmImagesRepository;

    public UtmImagesService(UtmImagesRepository utmImagesRepository) {
        this.utmImagesRepository = utmImagesRepository;
    }

    /**
     * Save a utmImages.
     *
     * @param utmImages the entity to save
     * @return the persisted entity
     */
    public UtmImages save(UtmImages utmImages) {
        log.debug("Request to save UtmImages : {}", utmImages);
        return utmImagesRepository.save(utmImages);
    }

    /**
     * Get all the utmImages.
     *
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public List<UtmImages> findAll() {
        return utmImagesRepository.findAll(Sort.by("shortName").ascending());
    }


    /**
     * Get one utmImages by id.
     *
     * @param shortName the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmImages> findOne(ImageShortName shortName) {
        log.debug("Request to get UtmImages : {}", shortName);
        return utmImagesRepository.findById(shortName);
    }

    public void reset() {
        final String ctx = CLASSNAME + ".reset";
        try {
            utmImagesRepository.reset();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
