package com.park.utmstack.service;

import com.park.utmstack.domain.UtmSpaceNotificationControl;
import com.park.utmstack.repository.UtmSpaceNotificationControlRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

/**
 * Service Implementation for managing UtmAlertLast.
 */
@Service
@Transactional
public class UtmSpaceNotificationControlService {

    private static final String CLASSNAME = "UtmSpaceNotificationControlService";
    private final UtmSpaceNotificationControlRepository spaceNotificationControlRepository;

    public UtmSpaceNotificationControlService(UtmSpaceNotificationControlRepository spaceNotificationControlRepository) {
        this.spaceNotificationControlRepository = spaceNotificationControlRepository;
    }

    /**
     * Save a utmAlertLast.
     *
     * @param utmAlertLast the entity to save
     * @return the persisted entity
     */
    public UtmSpaceNotificationControl save(UtmSpaceNotificationControl utmAlertLast) throws Exception {
        final String ctx = CLASSNAME + ".save";
        try {
            return spaceNotificationControlRepository.save(utmAlertLast);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    @Transactional
    public Optional<UtmSpaceNotificationControl> findById(Long id) throws Exception {
        final String ctx = CLASSNAME + ".getById";
        try {
            return spaceNotificationControlRepository.findById(id);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
