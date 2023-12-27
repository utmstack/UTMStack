package com.park.utmstack.web.rest.index_policy;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.index_policy.*;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.index_policy.IndexPolicyService;
import com.park.utmstack.util.UtilResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.Objects;

@RestController
@RequestMapping("/api/index-policy")
public class IndexPolicyResource {
    private static final String CLASSNAME = "IndexPolicyResource";

    private final Logger log = LoggerFactory.getLogger(IndexPolicyResource.class);

    private final IndexPolicyService indexPolicyService;
    private final ApplicationEventService applicationEventService;

    public IndexPolicyResource(IndexPolicyService indexPolicyService,
                               ApplicationEventService applicationEventService) {
        this.indexPolicyService = indexPolicyService;
        this.applicationEventService = applicationEventService;
    }

    @GetMapping("/policy")
    public ResponseEntity<PolicySettings> getPolicy() {
        final String ctx = CLASSNAME + ".getPolicy";
        try {
            IndexPolicy policy = indexPolicyService.getPolicy().orElse(null);

            if (Objects.isNull(policy))
                return ResponseEntity.ok().build();

            State openState = policy.getPolicy().getStates()
                .stream().filter(s -> s.getName().equals(Constants.STATE_OPEN)).findAny().orElse(null);

            Transition transition = Objects.requireNonNull(openState).getTransitions().get(0);

            PolicySettings response = new PolicySettings();
            response.setSnapshotActive(transition.getStateName().equals(Constants.STATE_SAFE_DELETE));
            response.setDeleteAfter(transition.getConditions().getMinIndexAge());

            return ResponseEntity.ok(response);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @PutMapping("/policy")
    public ResponseEntity<UpdateManagedIndexPolicyResponse> updateIndexPolicy(@Valid @RequestBody PolicySettings settings) {
        final String ctx = CLASSNAME + ".updateIndexPolicy";
        try {
            indexPolicyService.updatePolicy(settings);
            UpdateManagedIndexPolicyResponse rs = indexPolicyService.updateManagedIndexPolicy(settings.isSnapshotActive());
            return ResponseEntity.ok(Objects.requireNonNull(rs));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}
