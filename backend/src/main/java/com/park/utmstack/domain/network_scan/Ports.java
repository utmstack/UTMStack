package com.park.utmstack.domain.network_scan;

public class Ports {
    private Integer port;
    private String tcp;
    private String udp;

    public Integer getPort() {
        return port;
    }

    public void setPort(Integer port) {
        this.port = port;
    }

    public String getTcp() {
        return tcp;
    }

    public void setTcp(String tcp) {
        this.tcp = tcp;
    }

    public String getUdp() {
        return udp;
    }

    public void setUdp(String udp) {
        this.udp = udp;
    }
}
