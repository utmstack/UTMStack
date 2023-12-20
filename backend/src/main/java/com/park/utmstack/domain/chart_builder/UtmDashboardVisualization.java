package com.park.utmstack.domain.chart_builder;


import com.park.utmstack.domain.chart_builder.types.query.FilterType;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.io.Serializable;
import java.util.List;

/**
 * A UtmDashboardVisualization.
 */
@Entity
@Table(name = "utm_dashboard_visualization")
public class UtmDashboardVisualization implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "id_visualization", nullable = false)
    private Long idVisualization;

    @NotNull
    @Column(name = "id_dashboard", nullable = false)
    private Long idDashboard;

    @NotNull
    @Column(name = "dv_order", nullable = false)
    private Integer order;

    @NotNull
    @Column(name = "dv_width", nullable = false)
    private Float width;

    @NotNull
    @Column(name = "dv_height", nullable = false)
    private Float height;

    @NotNull
    @Column(name = "dv_top", nullable = false)
    private Float top;

    @NotNull
    @Column(name = "dv_left", nullable = false)
    private Float left;

    @Column(name = "dv_show_time_filter")
    private Boolean showTimeFilter;

    @Column(name = "dv_default_time_range")
    private String defaultTimeRange;

    @Column(name = "dv_grid_info")
    private String gridInfo;

    @ManyToOne(fetch = FetchType.EAGER)
    @JoinColumn(name = "id_visualization", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmVisualization visualization;

    @ManyToOne(fetch = FetchType.EAGER)
    @JoinColumn(name = "id_dashboard", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmDashboard dashboard;

    @Transient
    private List<FilterType> filters;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getIdVisualization() {
        return idVisualization;
    }

    public UtmDashboardVisualization idVisualization(Long idVisualization) {
        this.idVisualization = idVisualization;
        return this;
    }

    public void setIdVisualization(Long idVisualization) {
        this.idVisualization = idVisualization;
    }

    public Long getIdDashboard() {
        return idDashboard;
    }

    public UtmDashboardVisualization idDashboard(Long idDashboard) {
        this.idDashboard = idDashboard;
        return this;
    }

    public void setIdDashboard(Long idDashboard) {
        this.idDashboard = idDashboard;
    }

    public Integer getOrder() {
        return order;
    }

    public UtmDashboardVisualization order(Integer order) {
        this.order = order;
        return this;
    }

    public void setOrder(Integer order) {
        this.order = order;
    }

    public Float getWidth() {
        return width;
    }

    public UtmDashboardVisualization width(Float width) {
        this.width = width;
        return this;
    }

    public void setWidth(Float width) {
        this.width = width;
    }

    public Float getHeight() {
        return height;
    }

    public UtmDashboardVisualization height(Float height) {
        this.height = height;
        return this;
    }

    public void setHeight(Float height) {
        this.height = height;
    }

    public Float getTop() {
        return top;
    }

    public UtmDashboardVisualization top(Float top) {
        this.top = top;
        return this;
    }

    public void setTop(Float top) {
        this.top = top;
    }

    public Float getLeft() {
        return left;
    }

    public UtmDashboardVisualization left(Float left) {
        this.left = left;
        return this;
    }

    public void setLeft(Float left) {
        this.left = left;
    }

    public UtmVisualization getVisualization() {
        return visualization;
    }

    public void setVisualization(UtmVisualization visualization) {
        this.visualization = visualization;
    }

    public UtmDashboard getDashboard() {
        return dashboard;
    }

    public void setDashboard(UtmDashboard dashboard) {
        this.dashboard = dashboard;
    }

    public Boolean getShowTimeFilter() {
        return showTimeFilter;
    }

    public void setShowTimeFilter(Boolean showTimeFilter) {
        this.showTimeFilter = showTimeFilter;
    }

    public String getDefaultTimeRange() {
        return defaultTimeRange;
    }

    public void setDefaultTimeRange(String defaultTimeRange) {
        this.defaultTimeRange = defaultTimeRange;
    }

    public String getGridInfo() {
        return gridInfo;
    }

    public void setGridInfo(String gridInfo) {
        this.gridInfo = gridInfo;
    }
}
