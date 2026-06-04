package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.coordinate_map;

public class CoordinateMapChartResult {
    private String name;
    private Double[] value = new Double[3];

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Double[] getValue() {
        return value;
    }

    public void setValue(Double[] value) {
        this.value = value;
    }

    public CoordinateMapChartResult addLatitude(Double latitude) {
        value[0] = latitude;
        return this;
    }

    public CoordinateMapChartResult addLongitude(Double longitude) {
        value[1] = longitude;
        return this;
    }

    public CoordinateMapChartResult addMetricValue(Double metricValue) {
        value[2] = metricValue;
        return this;
    }

    //    public static void main(String[] args) {
    //        String ip = "192.0.0.127", mask = "192.0.0.0/25";
    //        IpAddressMatcher m = new IpAddressMatcher(mask);
    //        System.out.println(m.matches(ip));
    //    }
}
