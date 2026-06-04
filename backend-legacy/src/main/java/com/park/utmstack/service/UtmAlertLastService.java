package com.park.utmstack.service;

import com.park.utmstack.domain.UtmAlertLast;
import com.park.utmstack.repository.UtmAlertLastRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;
import java.time.ZoneOffset;
import java.util.Optional;

@Service
@Transactional
public class UtmAlertLastService {
    private static final String CLASSNAME = "UtmAlertLastService";
    private final Logger log = LoggerFactory.getLogger(UtmAlertLastService.class);

    private final UtmAlertLastRepository lastAlertRepository;

    public UtmAlertLastService(UtmAlertLastRepository lastAlertRepository) {
        this.lastAlertRepository = lastAlertRepository;
    }

    @EventListener(ApplicationReadyEvent.class)
    public void init() {
        final String ctx = CLASSNAME + ".init";
        try {
            Optional<UtmAlertLast> opt = lastAlertRepository.findById(1L);
            if (opt.isEmpty())
                lastAlertRepository.save(new UtmAlertLast(Instant.now().atZone(ZoneOffset.UTC).toInstant()));
        } catch (Exception e) {
            log.error(ctx + ": Fail to set the initial date for processing alerts");
        }
    }
}
