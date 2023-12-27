package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.heat_map_chart;

import com.park.utmstack.util.exceptions.UtmChartBuilderException;

import java.util.ArrayList;
import java.util.List;

public class HeatMapChartResult {
    private static final String CLASS_NAME = "HeatmapChartResult";
    private List<String> xAxis = new ArrayList<>(), yAxis = new ArrayList<>();
    private List<Double[]> data = new ArrayList<>();

    public HeatMapChartResult addXAxis(String xAxis) {
        if (!this.xAxis.contains(xAxis))
            this.xAxis.add(xAxis);
        return this;
    }

    public HeatMapChartResult addXAxis(int index, String xAxis) {
        if (!this.xAxis.contains(xAxis))
            this.xAxis.add(index, xAxis);
        return this;
    }

    public HeatMapChartResult addYAxis(String yAxis) {
        if (!this.yAxis.contains(yAxis))
            this.yAxis.add(yAxis);
        return this;
    }

    public HeatMapChartResult addYAxis(int index, String yAxis) {
        if (!this.yAxis.contains(yAxis))
            this.yAxis.add(index, yAxis);
        return this;
    }

    public HeatMapChartResult addData(Double[] value) throws UtmChartBuilderException {
        final String ctx = CLASS_NAME + ".addData";
        if (value.length != 3)
            throw new UtmChartBuilderException(ctx + ": Parameter value don't satisfy structure [x, y, value]");
        this.data.add(value);
        return this;
    }

    public List<String> getxAxis() {
        return xAxis;
    }

    public void setxAxis(List<String> xAxis) {
        this.xAxis = xAxis;
    }

    public List<String> getyAxis() {
        return yAxis;
    }

    public void setyAxis(List<String> yAxis) {
        this.yAxis = yAxis;
    }

    public List<Double[]> getData() {
        return data;
    }

    public void setData(List<Double[]> data) {
        this.data = data;
    }
}
