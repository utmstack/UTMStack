package com.park.utmstack.service;

import com.park.utmstack.domain.UtmAlertSocaiProcessingRequest;
import com.park.utmstack.repository.UtmAlertSocaiProcessingRequestRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
@Transactional
public class UtmAlertSocaiProcessingRequestService {

    private static final String CLASSNAME = "UtmAlertSocaiProcessingRequestService";
    private final Logger log = LoggerFactory.getLogger(UtmAlertSocaiProcessingRequestService.class);

    private final UtmAlertSocaiProcessingRequestRepository alertSocaiProcessingRequestRepository;

    public UtmAlertSocaiProcessingRequestService(UtmAlertSocaiProcessingRequestRepository alertSocaiProcessingRequestRepository) {
        this.alertSocaiProcessingRequestRepository = alertSocaiProcessingRequestRepository;
    }

    public void saveAll(List<UtmAlertSocaiProcessingRequest> requests) {
        final String ctx = CLASSNAME + ".saveAll";
        try {
            alertSocaiProcessingRequestRepository.saveAllAndFlush(requests);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }

    @Transactional(readOnly = true)
    public Page<UtmAlertSocaiProcessingRequest> findAll(Pageable pageable) {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return alertSocaiProcessingRequestRepository.findAll(pageable);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }

    public void delete(List<String> ids) {
        final String ctx = CLASSNAME + ".delete";
        try {
            alertSocaiProcessingRequestRepository.deleteAllById(ids);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }
}
