package com.park.utmstack.web.rest.overview;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.domain.shared_types.EventsByObjectsInTimeType;
import com.park.utmstack.domain.shared_types.static_dashboard.BarType;
import com.park.utmstack.domain.shared_types.static_dashboard.CardType;
import com.park.utmstack.domain.shared_types.static_dashboard.PieType;
import com.park.utmstack.domain.shared_types.static_dashboard.TableType;
import com.park.utmstack.service.UtmAlertService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.overview.OverviewService;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.opensearch.client.opensearch._types.aggregations.CalendarInterval;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

@RestController
@RequestMapping("/api/overview")
public class OverviewResource {
    private final Logger log = LoggerFactory.getLogger(OverviewResource.class);
    private static final String CLASS_NAME = "OverviewResource";

    private final OverviewService overviewService;
    private final UtmAlertService alertService;
    private final ApplicationEventService applicationEventService;

    public OverviewResource(OverviewService overviewService,
                            UtmAlertService alertService,
                            ApplicationEventService applicationEventService) {
        this.overviewService = overviewService;
        this.alertService = alertService;
        this.applicationEventService = applicationEventService;
    }

    //=================================================================================================
    //= ALERTS RESOURCES
    //=================================================================================================
    @GetMapping("/count-alerts-today-and-last-week")
    public ResponseEntity<List<CardType>> countAlertsTodayAndLastWeek() {
        final String ctx = CLASS_NAME + ".countAlertsTodayAndLastWeek";
        try {
            return ResponseEntity.ok(overviewService.countAlertsTodayAndLastWeek());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/count-alerts-by-status")
    public ResponseEntity<List<CardType>> countAlertsByStatus(@RequestParam String from, @RequestParam String to) {
        final String ctx = CLASS_NAME + ".countAlertsByStatus";
        try {
            FilterType timestampFilter = new FilterType(Constants.timestamp, OperatorType.IS_BETWEEN, Arrays.asList(from, to));
            FilterType statusFilter = new FilterType(Constants.alertStatus, OperatorType.IS_NOT, 1);
            List<FilterType> filters = new ArrayList<>();
            filters.add(timestampFilter);
            filters.add(statusFilter);
            return ResponseEntity.ok(alertService.countAlertsByStatus(filters));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/top-alerts")
    public ResponseEntity<TableType> topAlerts(@RequestParam String from, @RequestParam String to,
                                               @RequestParam Integer top) {
        final String ctx = CLASS_NAME + ".topAlerts";
        try {
            return ResponseEntity.ok(overviewService.topAlerts(from, to, top));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/count-alerts-by-severity")
    public ResponseEntity<PieType> countAlertsBySeverity(@RequestParam String from, @RequestParam String to,
                                                         @RequestParam Integer top) {
        final String ctx = CLASS_NAME + ".countAlertsBySeverity";
        try {
            return ResponseEntity.ok(overviewService.countAlertsBySeverity(from, to, top));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/top-alerts-by-category")
    public ResponseEntity<BarType> topAlertsByCategory(@RequestParam String from, @RequestParam String to,
                                                       @RequestParam Integer top) {
        final String ctx = CLASS_NAME + ".topAlertsByCategory";
        try {
            return ResponseEntity.ok(overviewService.topAlertsByCategory(from, to, top));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    //=================================================================================================
    //= EVENT RESOURCES
    //=================================================================================================
    @GetMapping("/count-events-by-type")
    public ResponseEntity<PieType> countEventsByType(@RequestParam String from, @RequestParam String to,
                                                     @RequestParam Integer top) {
        final String ctx = CLASS_NAME + ".countEventsByType";
        try {
            return ResponseEntity.ok(overviewService.countEventsByType(from, to, top));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/events-in-time")
    public ResponseEntity<EventsByObjectsInTimeType> eventsInTime(@RequestParam String from, @RequestParam String to,
                                                                  @RequestParam CalendarInterval interval) {
        final String ctx = CLASS_NAME + ".eventsInTime";
        try {
            return ResponseEntity.ok(overviewService.eventsInTime(from, to, interval));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/top-windows-events")
    public ResponseEntity<TableType> topWindowsEvents(@RequestParam String from, @RequestParam String to,
                                                      @RequestParam Integer top) {
        final String ctx = CLASS_NAME + ".topWindowsEvents";
        try {
            return ResponseEntity.ok(overviewService.topWindowsEvents(from, to, top));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}
