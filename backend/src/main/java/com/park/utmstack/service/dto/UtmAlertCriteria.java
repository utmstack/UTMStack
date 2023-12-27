package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.IntegerFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

public class UtmAlertCriteria implements Serializable {
    private static final long serialVersionUID = 1L;

    private LongFilter id;
    private IntegerFilter source;
    private StringFilter indexId;
    private StringFilter indexName;
    private InstantFilter timestamp;
    private StringFilter rule;
    private IntegerFilter severity;
    private StringFilter sensor;
    private StringFilter origin;
    private StringFilter destination;
    private IntegerFilter status;
    private IntegerFilter category;
    private LongFilter srcPort;
    private LongFilter destPort;
    private StringFilter proto;
    private StringFilter group;
    private StringFilter decoder;
    private StringFilter srcuser;
    private StringFilter dstuser;
    private StringFilter location;
    private StringFilter fullLog;
    private StringFilter url;
    private StringFilter hostname;
    private StringFilter programName;


    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public IntegerFilter getSource() {
        return source;
    }

    public void setSource(IntegerFilter source) {
        this.source = source;
    }

    public StringFilter getIndexId() {
        return indexId;
    }

    public void setIndexId(StringFilter indexId) {
        this.indexId = indexId;
    }

    public StringFilter getIndexName() {
        return indexName;
    }

    public void setIndexName(StringFilter indexName) {
        this.indexName = indexName;
    }

    public InstantFilter getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(InstantFilter timestamp) {
        this.timestamp = timestamp;
    }

    public StringFilter getRule() {
        return rule;
    }

    public void setRule(StringFilter rule) {
        this.rule = rule;
    }

    public IntegerFilter getSeverity() {
        return severity;
    }

    public void setSeverity(IntegerFilter severity) {
        this.severity = severity;
    }

    public StringFilter getSensor() {
        return sensor;
    }

    public void setSensor(StringFilter sensor) {
        this.sensor = sensor;
    }

    public StringFilter getOrigin() {
        return origin;
    }

    public void setOrigin(StringFilter origin) {
        this.origin = origin;
    }

    public StringFilter getDestination() {
        return destination;
    }

    public void setDestination(StringFilter destination) {
        this.destination = destination;
    }

    public IntegerFilter getStatus() {
        return status;
    }

    public void setStatus(IntegerFilter status) {
        this.status = status;
    }

    public IntegerFilter getCategory() {
        return category;
    }

    public void setCategory(IntegerFilter category) {
        this.category = category;
    }

    public LongFilter getSrcPort() {
        return srcPort;
    }

    public void setSrcPort(LongFilter srcPort) {
        this.srcPort = srcPort;
    }

    public LongFilter getDestPort() {
        return destPort;
    }

    public void setDestPort(LongFilter destPort) {
        this.destPort = destPort;
    }

    public StringFilter getProto() {
        return proto;
    }

    public void setProto(StringFilter proto) {
        this.proto = proto;
    }

    public StringFilter getGroup() {
        return group;
    }

    public void setGroup(StringFilter group) {
        this.group = group;
    }

    public StringFilter getDecoder() {
        return decoder;
    }

    public void setDecoder(StringFilter decoder) {
        this.decoder = decoder;
    }

    public StringFilter getSrcuser() {
        return srcuser;
    }

    public void setSrcuser(StringFilter srcuser) {
        this.srcuser = srcuser;
    }

    public StringFilter getDstuser() {
        return dstuser;
    }

    public void setDstuser(StringFilter dstuser) {
        this.dstuser = dstuser;
    }

    public StringFilter getLocation() {
        return location;
    }

    public void setLocation(StringFilter location) {
        this.location = location;
    }

    public StringFilter getFullLog() {
        return fullLog;
    }

    public void setFullLog(StringFilter fullLog) {
        this.fullLog = fullLog;
    }

    public StringFilter getUrl() {
        return url;
    }

    public void setUrl(StringFilter url) {
        this.url = url;
    }

    public StringFilter getHostname() {
        return hostname;
    }

    public void setHostname(StringFilter hostname) {
        this.hostname = hostname;
    }

    public StringFilter getProgramName() {
        return programName;
    }

    public void setProgramName(StringFilter programName) {
        this.programName = programName;
    }
}
