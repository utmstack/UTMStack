package com.park.utmstack.domain.network_scan;


import com.park.utmstack.service.dto.network_scan.NetworkScanDTO;

import javax.persistence.*;
import java.io.Serializable;

/**
 * A UtmOpenPort.
 */
@Entity
@Table(name = "utm_ports")
public class UtmPorts implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "scan_id")
    private Long scanId;

    @Column(name = "port")
    private Integer port;

    @Column(name = "tcp")
    private String tcp;

    @Column(name = "udp")
    private String udp;

    public UtmPorts() {
    }

    public UtmPorts(NetworkScanDTO.Port port, Long assetId) {
        this.scanId = assetId;
        this.port = port.getPort();
        this.tcp = port.getTcp();
        this.udp = port.getUdp();
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getScanId() {
        return scanId;
    }

    public void setScanId(Long scanId) {
        this.scanId = scanId;
    }

    public UtmPorts scanId(Long scanId) {
        this.scanId = scanId;
        return this;
    }

    public Integer getPort() {
        return port;
    }

    public UtmPorts port(Integer port) {
        this.port = port;
        return this;
    }

    public void setPort(Integer port) {
        this.port = port;
    }

    public String getTcp() {
        return tcp;
    }

    public UtmPorts tcp(String tcp) {
        this.tcp = tcp;
        return this;
    }

    public void setTcp(String tcp) {
        this.tcp = tcp;
    }

    public String getUdp() {
        return udp;
    }

    public UtmPorts udp(String udp) {
        this.udp = udp;
        return this;
    }

    public void setUdp(String udp) {
        this.udp = udp;
    }

    @Override
    public String toString() {
        return "UtmOpenPort {" +
            "id=" + getId() +
            ", scanId=" + getScanId() +
            ", port=" + getPort() +
            ", tcp='" + getTcp() + "'" +
            ", udp='" + getUdp() + "'" +
            "}";
    }
}
