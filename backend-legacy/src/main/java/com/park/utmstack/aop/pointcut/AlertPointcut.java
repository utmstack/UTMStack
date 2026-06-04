package com.park.utmstack.aop.pointcut;

import org.opensearch.client.opensearch._types.query_dsl.Query;
import org.springframework.stereotype.Component;

import java.time.Instant;

@Component
public class AlertPointcut {

    public void automaticAlertStatusChangePointcut(Query query, Integer status, String statusObservation, String indexPattern) {
        // Method is empty as this is just a Pointcut, the implementations are in the advices.
    }

    public void automaticAlertTagsChangePointcut(Query query, String tags, String ruleName, String indexPattern) {
        // Method is empty as this is just a Pointcut, the implementations are in the advices.
    }

    public void convertToIncidentPointcut(Query query, String incidentName, Integer incidentId, Instant incidentCreationDate, String incidentCreatedBy, String incidentSource, String indexPattern) {
        // Method is empty as this is just a Pointcut, the implementations are in the advices.
    }
}
