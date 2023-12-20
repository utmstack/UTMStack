package com.park.utmstack.service;

import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.shared_types.AlertType;
import com.park.utmstack.domain.shared_types.static_dashboard.CardType;
import com.park.utmstack.util.exceptions.DashboardOverviewException;
import com.park.utmstack.util.exceptions.ElasticsearchIndexDocumentUpdateException;
import com.park.utmstack.util.exceptions.UtmElasticsearchException;

import java.io.IOException;
import java.util.List;

/**
 * Service Interface for managing Alerts.
 */
public interface UtmAlertService {

    void updateStatus(List<String> alertIds, int status, String statusObservation) throws UtmElasticsearchException,
        IOException, ElasticsearchIndexDocumentUpdateException;

    void updateTags(List<String> alertIds, List<String> tags, Boolean createRule) throws ElasticsearchIndexDocumentUpdateException;

    List<CardType> countAlertsByStatus(List<FilterType> filters) throws DashboardOverviewException;

    void updateNotes(String alertId, String message) throws ElasticsearchIndexDocumentUpdateException;

    void convertToIncident(List<String> eventIds, String incidentName, Integer incidentId, String incidentSource) throws ElasticsearchIndexDocumentUpdateException;

    List<AlertType> getAlertsByIds(List<String> ids) throws UtmElasticsearchException;
}
