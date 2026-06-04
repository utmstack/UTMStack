package com.park.utmstack.domain.health_check.elasticsearch;

import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonSetter;

import java.util.Locale;
import java.util.Objects;

public class ElasticNode {
    private String name;
    private String cpuPercent;
    private String diskTotal;
    private String diskUsed;
    private String diskUsedPercent;
    private String diskAvailable;
    private String ramMax;
    private String ramCurrent;
    private String ramPercent;
    private String heapCurrent;
    private String heapMax;
    private String heapPercent;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    @JsonGetter("cpuPercent")
    public Float getCpuPercent() {
        return round(Float.parseFloat(cpuPercent));
    }

    @JsonSetter("cpu")
    public void setCpuPercent(String cpuPercent) {
        this.cpuPercent = cpuPercent;
    }

    @JsonGetter("diskTotal")
    public Float getDiskTotal() {
        return toGigabyte(this.diskTotal);
    }

    @JsonSetter("disk.total")
    public void setDiskTotal(String diskTotal) {
        this.diskTotal = diskTotal;
    }

    @JsonGetter("diskUsed")
    public Float getDiskUsed() {
        return toGigabyte(this.diskUsed);
    }

    @JsonSetter("disk.used")
    public void setDiskUsed(String diskUsed) {
        this.diskUsed = diskUsed;
    }

    @JsonGetter("diskUsedPercent")
    public Float getDiskUsedPercent() {
        return round(Float.parseFloat(diskUsedPercent));
    }

    @JsonSetter("disk.used_percent")
    public void setDiskUsedPercent(String diskUsedPercent) {
        this.diskUsedPercent = diskUsedPercent;
    }

    @JsonGetter("diskAvailable")
    public Float getDiskAvailable() {
        return toGigabyte(this.diskAvailable);
    }

    @JsonSetter("disk.avail")
    public void setDiskAvailable(String diskAvailable) {
        this.diskAvailable = diskAvailable;
    }

    @JsonGetter("ramMax")
    public Float getRamMax() {
        return toGigabyte(this.ramMax);
    }

    @JsonSetter("ram.max")
    public void setRamMax(String ramMax) {
        this.ramMax = ramMax;
    }

    @JsonGetter("ramCurrent")
    public Float getRamCurrent() {
        return toGigabyte(this.ramCurrent);
    }

    @JsonSetter("ram.current")
    public void setRamCurrent(String ramCurrent) {
        this.ramCurrent = ramCurrent;
    }

    @JsonGetter("ramPercent")
    public Float getRamPercent() {
        return round(Float.parseFloat(ramPercent));
    }

    @JsonSetter("ram.percent")
    public void setRamPercent(String ramPercent) {
        this.ramPercent = ramPercent;
    }

    @JsonGetter("heapCurrent")
    public Float getHeapCurrent() {
        return toGigabyte(this.heapCurrent);
    }

    @JsonSetter("heap.current")
    public void setHeapCurrent(String heapCurrent) {
        this.heapCurrent = heapCurrent;
    }

    @JsonGetter("heapMax")
    public Float getHeapMax() {
        return toGigabyte(this.heapMax);
    }

    @JsonSetter("heap.max")
    public void setHeapMax(String heapMax) {
        this.heapMax = heapMax;
    }

    @JsonGetter("heapPercent")
    public Float getHeapPercent() {
        return round(Float.parseFloat(heapPercent));
    }

    @JsonSetter("heap.percent")
    public void setHeapPercent(String heapPercent) {
        this.heapPercent = heapPercent;
    }

    @JsonIgnore
    private Float toGigabyte(String measure) {
        String[] value = measure.split("[a-zA-Z]+");
        String[] unit = measure.split("^[\\d]+[.]?([\\d]+)?");

        float tempValue = value.length > 0 ? Float.parseFloat(value[0]) : 0;
        String tempUnit = unit.length > 1 ? unit[1] : "";

        if (Objects.requireNonNull(tempUnit, "No unit measure is present").equals("tb"))
            tempValue *= Math.pow(2, 10); // 1024
        else if (Objects.requireNonNull(tempUnit, "No unit measure is present").equals("mb"))
            tempValue /= Math.pow(2, 10); // 1024
        else if (Objects.requireNonNull(tempUnit, "No unit measure is present").equals("kb"))
            tempValue /= Math.pow(2, 20); // 1024 * 1024 = 1048576

        return round(tempValue);
    }

    @JsonIgnore
    private Float round(Float value) {
        if (Objects.isNull(value))
            return (float) 0;
        return Float.parseFloat(String.format(Locale.US,"%.2f", value));
    }
}
