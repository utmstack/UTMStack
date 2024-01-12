package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmImages;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.shared_types.enums.ImageShortName;
import com.park.utmstack.repository.UtmImagesRepository;
import com.park.utmstack.service.UtmImagesService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmImages.
 */
@RestController
@RequestMapping("/api")
public class UtmImagesResource {

    private final Logger log = LoggerFactory.getLogger(UtmImagesResource.class);

    private static final String CLASSNAME = "UtmImagesResource";

    private final UtmImagesService utmImagesService;
    private final UtmImagesRepository imagesRepository;
    private final ApplicationEventService applicationEventService;

    public UtmImagesResource(UtmImagesService utmImagesService,
                             UtmImagesRepository imagesRepository,
                             ApplicationEventService applicationEventService) {
        this.utmImagesService = utmImagesService;
        this.imagesRepository = imagesRepository;
        this.applicationEventService = applicationEventService;
    }


    @PutMapping("/images")
    public ResponseEntity<UtmImages> updateImage(@Valid @RequestBody UtmImages image) {
        final String ctx = CLASSNAME + ".updateImage";
        try {
            Optional<UtmImages> imageOpt = imagesRepository.findById(image.getShortName());

            if (imageOpt.isEmpty())
                return UtilResponse.buildBadRequestResponse("Image short name not recognized: " + image.getShortName());

            UtmImages img = imageOpt.get();
            img.setUserImg(image.getUserImg());
            return ResponseEntity.ok(utmImagesService.save(img));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildInternalServerErrorResponse(msg);
        }
    }

    @GetMapping("/images/all")
    public ResponseEntity<List<UtmImages>> getAllImages() {
        final String ctx = CLASSNAME + ".getAllImages";
        try {
            return ResponseEntity.ok(utmImagesService.findAll());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @GetMapping("/images/{shortName}")
    public ResponseEntity<UtmImages> getImage(@PathVariable ImageShortName shortName) {
        final String ctx = CLASSNAME + ".getImage";
        try {
            Optional<UtmImages> utmImages = utmImagesService.findOne(shortName);
            return ResponseUtil.wrapOrNotFound(utmImages);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @GetMapping("images/reset")
    public ResponseEntity<Void> reset() {
        final String ctx = CLASSNAME + ".reset";
        try {
            utmImagesService.reset();
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }
}
