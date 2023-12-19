package com.park.utmstack.service.incident;

import com.park.utmstack.domain.incident.UtmIncidentHistory;
import com.park.utmstack.domain.incident.enums.IncidentHistoryActionEnum;
import com.park.utmstack.repository.incident.UtmIncidentHistoryRepository;
import com.park.utmstack.service.UserService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Date;
import java.util.Optional;

/**
 * Service Implementation for managing UtmIncidentHistory.
 */
@Service
@Transactional
public class UtmIncidentHistoryService {
    private final String CLASSNAME = "UtmIncidentHistoryService";
    private final Logger log = LoggerFactory.getLogger(UtmIncidentHistoryService.class);

    private final UtmIncidentHistoryRepository utmIncidentHistoryRepository;

    private final UserService userService;

    public UtmIncidentHistoryService(UtmIncidentHistoryRepository utmIncidentHistoryRepository, UserService userService) {
        this.utmIncidentHistoryRepository = utmIncidentHistoryRepository;
        this.userService = userService;
    }

    /**
     * Save a utmIncidentHistory.
     *
     * @param utmIncidentHistory the entity to save
     * @return the persisted entity
     */
    public UtmIncidentHistory save(UtmIncidentHistory utmIncidentHistory) {
        log.debug("Request to save UtmIncidentHistory : {}", utmIncidentHistory);
        return utmIncidentHistoryRepository.save(utmIncidentHistory);
    }


    public void createHistory(IncidentHistoryActionEnum incidentHistoryActionEnum, Long incidentId, String action, String actionDetail) {
        final String ctx = CLASSNAME + ".createHistory";
        try {
            String userLogin = userService.getCurrentUserLogin().getLogin();
            UtmIncidentHistory utmIncidentHistory = new UtmIncidentHistory();
            utmIncidentHistory.setActionDate(new Date().toInstant());
            utmIncidentHistory.actionCreatedBy(userLogin);
            utmIncidentHistory.setActionType(incidentHistoryActionEnum);
            utmIncidentHistory.setActionDetail(actionDetail);
            utmIncidentHistory.setAction(action);
            utmIncidentHistory.setIncidentId(incidentId);
            utmIncidentHistoryRepository.save(utmIncidentHistory);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }

    }

    /**
     * Get all the utmIncidentHistories.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentHistory> findAll(Pageable pageable) {
        log.debug("Request to get all UtmIncidentHistories");
        return utmIncidentHistoryRepository.findAll(pageable);
    }


    /**
     * Get one utmIncidentHistory by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmIncidentHistory> findOne(Long id) {
        log.debug("Request to get UtmIncidentHistory : {}", id);
        return utmIncidentHistoryRepository.findById(id);
    }

    /**
     * Delete the utmIncidentHistory by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmIncidentHistory : {}", id);
        utmIncidentHistoryRepository.deleteById(id);
    }
}
