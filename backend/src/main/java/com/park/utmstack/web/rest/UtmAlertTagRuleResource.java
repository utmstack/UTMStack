package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmAlertTag;
import com.park.utmstack.domain.UtmAlertTagRule;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.UtmAlertTagRuleService;
import com.park.utmstack.service.UtmAlertTagService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import com.park.utmstack.web.rest.vm.AlertTagRuleFilterVM;
import com.park.utmstack.web.rest.vm.AlertTagRuleVM;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.util.CollectionUtils;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.net.URISyntaxException;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.Optional;

/**
 * REST controller for managing UtmTagRule.
 */
@RestController
@RequestMapping("/api")
public class UtmAlertTagRuleResource {

    private static final String CLASSNAME = "UtmTagRuleResource";
    private final Logger log = LoggerFactory.getLogger(UtmAlertTagRuleResource.class);

    private final UtmAlertTagRuleService tagRuleService;
    private final ApplicationEventService applicationEventService;
    private final UtmAlertTagService alertTagService;

    public UtmAlertTagRuleResource(UtmAlertTagRuleService tagRuleService,
                                   ApplicationEventService applicationEventService,
                                   UtmAlertTagService alertTagService) {
        this.tagRuleService = tagRuleService;
        this.applicationEventService = applicationEventService;
        this.alertTagService = alertTagService;
    }

    @PostMapping("/alert-tag-rules")
    public ResponseEntity<AlertTagRuleVM> createAlertTagRule(@Valid @RequestBody AlertTagRuleVM ruleVM) {
        final String ctx = CLASSNAME + ".createUtmAlertTagRule";
        try {
            if (!Objects.isNull(ruleVM.getId()))
                throw new Exception("A new alert tag rule can't have an ID");
            ruleVM.setActive(true);
            ruleVM.setDeleted(false);

            UtmAlertTagRule alertTagRule = tagRuleService.save(new UtmAlertTagRule(ruleVM));

            ruleVM.setId(alertTagRule.getId());
            ruleVM.setCreatedBy(alertTagRule.getCreatedBy());
            ruleVM.setCreatedDate(alertTagRule.getCreatedDate());
            ruleVM.setLastModifiedBy(alertTagRule.getLastModifiedBy());
            ruleVM.setLastModifiedDate(alertTagRule.getLastModifiedDate());

            return ResponseEntity.ok(ruleVM);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @PutMapping("/alert-tag-rules")
    public ResponseEntity<AlertTagRuleVM> updateAlertTagRule(@Valid @RequestBody AlertTagRuleVM ruleVM) throws URISyntaxException {
        final String ctx = CLASSNAME + ".updateAlertTagRule";
        try {
            if (Objects.isNull(ruleVM.getId()))
                throw new Exception("ID can't be null");

            UtmAlertTagRule alertTagRule = tagRuleService.save(new UtmAlertTagRule(ruleVM));

            ruleVM.setLastModifiedBy(alertTagRule.getLastModifiedBy());
            ruleVM.setLastModifiedDate(alertTagRule.getLastModifiedDate());

            return ResponseEntity.ok(ruleVM);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @GetMapping("/alert-tag-rules")
    public ResponseEntity<List<AlertTagRuleVM>> getAlertTagRulesByFilter(AlertTagRuleFilterVM filters, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAlertTagRulesByFilter";
        try {
            Page<UtmAlertTagRule> page = tagRuleService.findByFilter(filters, pageable);

            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/alert-tag-rules/filter");
            List<AlertTagRuleVM> result = new ArrayList<>();

            for (UtmAlertTagRule content : page) {
                List<UtmAlertTag> tags = alertTagService.findAllByIdIn(content.getAppliedTagsAsListOfLong());
                result.add(new AlertTagRuleVM(content, tags));
            }
            return ResponseEntity.ok().headers(headers).body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @GetMapping("/alert-tag-rules/get-by-ids")
    public ResponseEntity<List<AlertTagRuleVM>> getAlertTagRulesByIds(@RequestParam List<Long> ids) {
        final String ctx = CLASSNAME + ".getAlertTagRulesByIds";
        try {
            List<UtmAlertTagRule> rules = tagRuleService.findByIdIn(ids);

            if (CollectionUtils.isEmpty(rules))
                return ResponseEntity.ok().build();

            List<AlertTagRuleVM> result = new ArrayList<>();
            for (UtmAlertTagRule rule : rules) {
                List<UtmAlertTag> tags = alertTagService.findAllByIdIn(rule.getAppliedTagsAsListOfLong());
                result.add(new AlertTagRuleVM(rule, tags));
            }

            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }


    @GetMapping("/alert-tag-rules/{id}")
    public ResponseEntity<AlertTagRuleVM> getAlertTagRule(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getAlertTagRule";
        try {
            Optional<UtmAlertTagRule> opt = tagRuleService.findOne(id);
            if (opt.isPresent()) {
                List<UtmAlertTag> tags = alertTagService.findAllByIdIn(opt.get().getAppliedTagsAsListOfLong());
                return ResponseEntity.ok(new AlertTagRuleVM(opt.get(), tags));
            }
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @DeleteMapping("/alert-tag-rules/{id}")
    public ResponseEntity<Void> deleteAlertTagRule(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteAlertTagRule";
        tagRuleService.delete(id);
        return ResponseEntity.ok().build();
    }
}
