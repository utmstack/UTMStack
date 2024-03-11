package com.park.utmstack.service.compliance;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.User;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.chart_builder.UtmDashboard_;
import com.park.utmstack.domain.compliance.UtmComplianceReportConfig_;
import com.park.utmstack.domain.compliance.UtmComplianceReportSchedule;
import com.park.utmstack.domain.compliance.UtmComplianceReportSchedule_;
import com.park.utmstack.repository.compliance.UtmComplianceReportScheduleRepository;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.compliance.UtmComplianceReportScheduleCriteria;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.scheduling.support.CronExpression;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import tech.jhipster.service.QueryService;

import javax.persistence.criteria.JoinType;
import java.time.Clock;
import java.time.Instant;
import java.time.ZoneOffset;
import java.util.List;
import java.util.Objects;
import java.util.Optional;

/**
 * Service Implementation for managing {@link UtmComplianceReportSchedule}.
 */
@Service
@Transactional
public class UtmComplianceReportScheduleService extends QueryService<UtmComplianceReportSchedule> {

    private final Logger log = LoggerFactory.getLogger(UtmComplianceReportScheduleService.class);
    private static final String CLASSNAME = "UtmComplianceReportScheduleService";
    private final ApplicationEventService applicationEventService;

    private final UtmComplianceReportScheduleRepository utmComplianceReportScheduleRepository;
    private final UserService userService;
    private final ComplianceMailService complianceMailService;

    public UtmComplianceReportScheduleService(ApplicationEventService applicationEventService,
                                              UtmComplianceReportScheduleRepository utmComplianceReportScheduleRepository,
                                              UserService userService,
                                              ComplianceMailService complianceMailService) {
        this.applicationEventService = applicationEventService;
        this.utmComplianceReportScheduleRepository = utmComplianceReportScheduleRepository;
        this.userService = userService;
        this.complianceMailService = complianceMailService;
    }

    /**
     * Save a utmComplianceReportSchedule.
     *
     * @param utmComplianceReportSchedule the entity to save.
     * @return the persisted entity.
     */
    public UtmComplianceReportSchedule save(UtmComplianceReportSchedule utmComplianceReportSchedule) throws Exception {
        log.debug("Request to save UtmComplianceReportSchedule : {}", utmComplianceReportSchedule);
        final String ctx = CLASSNAME + ".save";
        User user = userService.getCurrentUserLogin();
        utmComplianceReportSchedule.setUserId(user.getId());

        try {
            CronExpression.parse(utmComplianceReportSchedule.getScheduleString());
        } catch (IllegalArgumentException arg) {
            throw new BadRequestAlertException(ctx + ": " + "Invalid value of field -> scheduleString", "utmComplianceReportSchedule", "invalidvalue");
        }
        if (utmComplianceReportSchedule.getId() == null) { // When inserting set time to now
            utmComplianceReportSchedule.setLastExecutionTime(Instant.now(Clock.systemUTC()));
        } else { // When updating, use db Instant to avoid update
            utmComplianceReportSchedule.setLastExecutionTime(findOne(utmComplianceReportSchedule.getId()).get().getLastExecutionTime());
        }
        return utmComplianceReportScheduleRepository.save(utmComplianceReportSchedule);
    }

    /**
     * Get all the utmComplianceReportSchedules.
     *
     * @return the list of entities.
     */
    @Transactional(readOnly = true)
    public List<UtmComplianceReportSchedule> findAll() {
        log.debug("Request to get all UtmComplianceReportSchedules");
        return utmComplianceReportScheduleRepository.findAll();
    }

    /**
     * Get all the utmComplianceReportSchedules for the connected user.
     *
     * @return the list of entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmComplianceReportSchedule> findAllOfCurrentUser(UtmComplianceReportScheduleCriteria criteria, Pageable pageable) {
        log.debug("Request to get all UtmComplianceReportSchedules");
        log.debug("find by criteria : {}, page: {}", criteria, pageable);
        final Specification<UtmComplianceReportSchedule> specification = createSpecification(criteria);
        return utmComplianceReportScheduleRepository.findAll(specification, pageable);
    }

    /**
     * Get one utmComplianceReportSchedule by id.
     *
     * @param id the id of the entity.
     * @return the entity.
     */
    @Transactional(readOnly = true)
    public Optional<UtmComplianceReportSchedule> findOne(Long id) {
        log.debug("Request to get UtmComplianceReportSchedule : {}", id);
        return utmComplianceReportScheduleRepository.findById(id);
    }

    /**
     * Get one utmComplianceReportSchedule by its field values and current connected user.
     *
     * @param reportSchedule the UtmComplianceReportSchedule to search for.
     * @return the entity if exists.
     */
    @Transactional(readOnly = true)
    public Optional<UtmComplianceReportSchedule> findByComplianceReportValues(UtmComplianceReportSchedule reportSchedule) {
        log.debug("Request to get UtmComplianceReportSchedule : {}", reportSchedule);
        User user = userService.getCurrentUserLogin();
        return utmComplianceReportScheduleRepository.findFirstByUserIdAndComplianceIdAndScheduleString(user.getId(),
            reportSchedule.getComplianceId(),reportSchedule.getScheduleString());
    }

    /**
     * Delete the utmComplianceReportSchedule by id.
     *
     * @param id the id of the entity.
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmComplianceReportSchedule : {}", id);
        utmComplianceReportScheduleRepository.deleteById(id);
    }


    /**
     * Scheduled method to execute the compliance report pdf generation and email delivery
     * */
    @Scheduled(fixedDelay = 5000, initialDelay = 30000)
    public void scheduleComplianceReport() {
        final String ctx = CLASSNAME + ".scheduleComplianceReport";

        List<UtmComplianceReportSchedule> schedulesList = findAll();

        schedulesList.forEach(current -> {
            Optional<User> user = userService.getUserWithAuthorities(current.getUserId());
            try {
                Instant currentDate = Instant.now(Clock.systemUTC());
                Instant next = getNext(current.getScheduleString(), current.getLastExecutionTime(), currentDate);
                if (isTimeToExecute(next, currentDate)) {
                    // Set the next execution time (Base time seed)
                    current.setLastExecutionTime(next);
                    utmComplianceReportScheduleRepository.save(current);
                    complianceMailService.sendComplianceByMail(Constants.FRONT_BASE_URL + current.getUrlWithParams(),
                        user.get().getEmail());
                }

            } catch (Exception e) {
                String msg = ctx + ": " + e.getLocalizedMessage();
                log.error(msg);
                applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            }
        });
    }

    /***
     * Method to know if is time to execute the task
     */
    private boolean isTimeToExecute(Instant next, Instant currentDate) {
        return currentDate.atZone(ZoneOffset.UTC).toInstant().isAfter(next);
    }

    /**
     * Method to know the next valid Instant to execute the task, even if the system was shut down for a while
     * */
    private Instant getNext(String cronExpresion, Instant lastExecution, Instant currentDate) {
        CronExpression parse = CronExpression.parse(cronExpresion);
        Instant possibleNext = Objects.requireNonNull(parse.next(lastExecution.atZone(ZoneOffset.UTC))).toInstant();
        Long diffBetweenLastAndNext = possibleNext.getEpochSecond() - lastExecution.atZone(ZoneOffset.UTC).toInstant().getEpochSecond();
        Long currentSecs = currentDate.atZone(ZoneOffset.UTC).toInstant().getEpochSecond();

        if (possibleNext.plusSeconds(diffBetweenLastAndNext).getEpochSecond() >= currentSecs) {
            return possibleNext;
        } else {
            // Then is a delay between the current date and last execution, so we have to move to the
            // near next execution to avoid extra executions, because the general scheduler that calls these methods,
            // is every 5 seconds
            Long diffBetweenCurrentAndPossibleNext = currentSecs - possibleNext.getEpochSecond();
            Integer rate = Long.valueOf(diffBetweenCurrentAndPossibleNext/diffBetweenLastAndNext).intValue();
            Instant resultNext = lastExecution.atZone(ZoneOffset.UTC).toInstant().plusSeconds(diffBetweenLastAndNext * rate);
            return resultNext.atZone(ZoneOffset.UTC).toInstant();
        }

    }

    private Specification<UtmComplianceReportSchedule> createSpecification(UtmComplianceReportScheduleCriteria criteria) {

        User user = userService.getCurrentUserLogin();
        Specification<UtmComplianceReportSchedule> specification = Specification.where((root, query, criteriaBuilder) ->
                                                                                        criteriaBuilder.equal(root.get("userId"), user.getId()));
        if (criteria != null) {
            if (criteria.getName() != null) {
                specification = specification.and(buildSpecification(criteria.getName(),
                        root -> root.join(UtmComplianceReportSchedule_.compliance, JoinType.INNER).join(UtmComplianceReportConfig_.associatedDashboard, JoinType.INNER).get(UtmDashboard_.name)));
            }
        }
        return specification;
    }


}
