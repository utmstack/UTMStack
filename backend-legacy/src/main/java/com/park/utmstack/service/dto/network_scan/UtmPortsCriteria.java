package com.park.utmstack.service.dto.network_scan;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.IntegerFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmOpenPort entity. This class is used in UtmOpenPortResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-open-ports?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmPortsCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private LongFilter scanId;

    private IntegerFilter port;

    private StringFilter tcp;

    private StringFilter udp;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getScanId() {
        return scanId;
    }

    public void setScanId(LongFilter scanId) {
        this.scanId = scanId;
    }

    public IntegerFilter getPort() {
        return port;
    }

    public void setPort(IntegerFilter port) {
        this.port = port;
    }

    public StringFilter getTcp() {
        return tcp;
    }

    public void setTcp(StringFilter tcp) {
        this.tcp = tcp;
    }

    public StringFilter getUdp() {
        return udp;
    }

    public void setUdp(StringFilter udp) {
        this.udp = udp;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmPortsCriteria that = (UtmPortsCriteria) o;
        return
            Objects.equals(id, that.id) &&
                Objects.equals(port, that.port) &&
                Objects.equals(tcp, that.tcp) &&
                Objects.equals(udp, that.udp);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
            id,
            port,
            tcp,
            udp
        );
    }

    @Override
    public String toString() {
        return "UtmOpenPortCriteria{" +
            (id != null ? "id=" + id + ", " : "") +
            (port != null ? "port=" + port + ", " : "") +
            (tcp != null ? "tcp=" + tcp + ", " : "") +
            (udp != null ? "udp=" + udp + ", " : "") +
            "}";
    }

}
