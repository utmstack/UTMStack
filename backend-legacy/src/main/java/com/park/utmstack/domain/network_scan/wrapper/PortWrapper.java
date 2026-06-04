package com.park.utmstack.domain.network_scan.wrapper;

import com.park.utmstack.domain.network_scan.Ports;
import com.park.utmstack.domain.network_scan.UtmPorts;
import org.springframework.util.CollectionUtils;

import java.util.ArrayList;
import java.util.List;

public class PortWrapper {

    public static UtmPorts build(Ports ports, Long scanId) {
        UtmPorts result = new UtmPorts();
        return result.udp(ports.getUdp()).tcp(ports.getTcp()).port(ports.getPort()).scanId(scanId);
    }

    /**
     * @param ports
     * @param scanId
     * @return
     */
    public static List<UtmPorts> build(List<Ports> ports, Long scanId) {
        List<UtmPorts> result = new ArrayList<>();
        if (CollectionUtils.isEmpty(ports))
            return result;

        ports.forEach(port -> {
            UtmPorts utmPort = new UtmPorts();
            utmPort.udp(port.getUdp()).tcp(port.getTcp()).port(port.getPort()).scanId(scanId);
            result.add(utmPort);
        });
        return result;
    }
}
