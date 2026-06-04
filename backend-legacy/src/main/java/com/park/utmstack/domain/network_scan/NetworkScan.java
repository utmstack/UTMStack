package com.park.utmstack.domain.network_scan;

import java.util.List;
import java.util.Objects;

public class NetworkScan {
    private String ip;
    private String mac;
    private String os;
    private String name;
    private Boolean alive;
    private List<String> addressList;
    private List<String> aliasList;
    private List<Ports> ports;

    public String getIp() {
        return ip;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public String getMac() {
        return mac;
    }

    public void setMac(String mac) {
        this.mac = mac;
    }

    public String getOs() {
        return os;
    }

    public void setOs(String os) {
        this.os = os;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Boolean getAlive() {
        return alive;
    }

    public void setAlive(Boolean alive) {
        this.alive = alive;
    }

    public List<Ports> getPorts() {
        return ports;
    }

    public void setPorts(List<Ports> ports) {
        this.ports = ports;
    }

    public List<String> getAddressList() {
        return addressList;
    }

    public void setAddressList(List<String> addressList) {
        this.addressList = addressList;
    }

    public List<String> getAliasList() {
        return aliasList;
    }

    public void setAliasList(List<String> aliasList) {
        this.aliasList = aliasList;
    }

    public int getUUID() {
        return Objects.hash(ip, mac);
    }

    @Override
    public int hashCode() {
        return Objects.hash(ip, os, name, alive);
    }

    @Override
    public String toString() {
        return "NetworkScan{" +
            "ip='" + ip + '\'' +
            ", mac='" + mac + '\'' +
            ", os='" + os + '\'' +
            ", name='" + name + '\'' +
            ", alive=" + alive +
            '}';
    }
}
