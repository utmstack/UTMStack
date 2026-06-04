package com.park.utmstack.service.dto.chart_builder;

import tech.jhipster.service.filter.IntegerFilter;
import tech.jhipster.service.filter.LongFilter;

import java.io.Serializable;

public class UtmDashboardVisualizationCriteria implements Serializable {
    private LongFilter id;
    private LongFilter idVisualization;
    private LongFilter idDashboard;
    private IntegerFilter order;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getIdVisualization() {
        return idVisualization;
    }

    public void setIdVisualization(LongFilter idVisualization) {
        this.idVisualization = idVisualization;
    }

    public LongFilter getIdDashboard() {
        return idDashboard;
    }

    public void setIdDashboard(LongFilter idDashboard) {
        this.idDashboard = idDashboard;
    }

    public IntegerFilter getOrder() {
        return order;
    }

    public void setOrder(IntegerFilter order) {
        this.order = order;
    }
}
