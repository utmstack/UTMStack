package com.park.utmstack.service;

import com.park.utmstack.domain.UtmImages;
import com.park.utmstack.domain.shared_types.enums.ImageShortName;
import com.park.utmstack.repository.UtmImagesRepository;
import com.park.utmstack.util.ImageValidatorUtil;
import com.park.utmstack.util.enums.ImageComponents;
import com.park.utmstack.util.exceptions.UtmImageValidationException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.io.InputStream;
import java.util.List;
import java.util.Map;
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
    public UtmImages save(UtmImages utmImages) throws UtmImageValidationException {
        log.debug("Request to save UtmImages : {}", utmImages);
        try {
            Map<ImageComponents,Object> imageComponents = ImageValidatorUtil.imageComponents(utmImages.getUserImg());
            if(imageComponents.isEmpty()) {
                throw new NullPointerException("Map is empty");
            }
            String mimeType = (String) imageComponents.get(ImageComponents.IMG_MIME_TYPE);
            String extension = (String) imageComponents.get(ImageComponents.IMG_FILE_EXTENSION);
            InputStream imageData = (InputStream) imageComponents.get(ImageComponents.IMG_DATA);

            if (!ImageValidatorUtil.isMimeTypeAllowed(mimeType)) {
                throw new UtmImageValidationException("Could not update UtmImages: Invalid image MIME_TYPE (" + mimeType + ") only (" + String.join(",", ImageValidatorUtil.ALLOWED_MIME_TYPES) + ") MIME_TYPES are allowed");
            }
            if (!ImageValidatorUtil.isExtensionAllowed(extension)) {
                throw new UtmImageValidationException("Could not update UtmImages: Invalid image extension (" + extension + ") only (" + String.join(",", ImageValidatorUtil.ALLOWED_EXTENSIONS) + ") extensions are allowed");
            }
            if (!ImageValidatorUtil.isRealImage(imageData)) {
                throw new UtmImageValidationException("Could not update UtmImages: Not a valid image");
            }
            return utmImagesRepository.save(utmImages);
        } catch (UtmImageValidationException e) {
            throw new UtmImageValidationException(e.getMessage());
        } catch (Exception e) {
            throw new UtmImageValidationException("Could not insert the image: The image is corrupted, is not a valid image or maybe is a SVG (SVG is not allowed)");
        }
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
