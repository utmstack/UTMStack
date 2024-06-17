package com.park.utmstack.service.collectors;

import agent.CollectorOuterClass;
import com.park.utmstack.domain.collector.UtmCollector;
import com.park.utmstack.domain.network_scan.NetworkScanFilter;
import com.park.utmstack.repository.collector.UtmCollectorRepository;
import com.park.utmstack.service.dto.collectors.dto.CollectorDTO;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.dao.InvalidDataAccessResourceUsageException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.Objects;

@Service
public class UtmCollectorService {

    private static final String CLASSNAME = "UtmCollectorService";

    private final UtmCollectorRepository utmCollectorRepository;
    private final DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    private final Logger log = LoggerFactory.getLogger(UtmCollectorService.class);

    public UtmCollectorService(UtmCollectorRepository utmCollectorRepository) {
        this.utmCollectorRepository = utmCollectorRepository;
    }

    public UtmCollector saveCollector(CollectorOuterClass.Collector collector){
        UtmCollector utmCollector = utmCollectorRepository.findById(Long.valueOf(collector.getId()))
                .orElse(new UtmCollector());

        if (utmCollector.getId() == null) {
            utmCollector.setId(Long.valueOf(collector.getId()));
        }

        utmCollector.setStatus(collector.getStatus().name());
        utmCollector.setLastSeen(LocalDateTime.parse(collector.getLastSeen(), this.formatter));
        utmCollector.setVersion(collector.getVersion());
        utmCollector.setIp(collector.getIp());
        utmCollector.setHostname(collector.getHostname());
        utmCollector.setCollectorKey(collector.getCollectorKey());
        utmCollector.setModule(collector.getModule().name());
        utmCollector.setActive(true);


        return this.utmCollectorRepository.save(utmCollector);

    }

    public void synchronize(List<CollectorDTO> collectorDTOS) {
        List<UtmCollector> collectors = utmCollectorRepository.findAll();

        for (UtmCollector collector : collectors) {
            CollectorDTO collectorDTO = collectorDTOS.stream()
                    .filter(c -> c.getId() == collector.getId())
                    .findFirst()
                    .orElse(null);

            if (!Objects.nonNull(collectorDTO)) {
                collector.setActive(false);
                collector.setStatus("OFFLINE");
                this.utmCollectorRepository.save(collector);
            }
        }
    }

    public Page<CollectorDTO> searchByFilters(NetworkScanFilter f, Pageable p) throws RuntimeException {
        final String ctx = CLASSNAME + ".searchByFilters";
        try {
            Page<UtmCollector> page = filter(f, p);
            return page.map(CollectorDTO::new);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private Page<UtmCollector> filter(NetworkScanFilter f, Pageable p) throws Exception {
        final String ctx = CLASSNAME + ".filter";
        try {
            return utmCollectorRepository.searchByFilters(
                    f.getAssetIpMacName() == null ? null : "%" + f.getAssetIpMacName() + "%",
                    f.getDiscoveredInitDate(),
                    f.getDiscoveredEndDate(),
                    f.getGroups(),p);

        } catch (InvalidDataAccessResourceUsageException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            throw new Exception(msg);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}
