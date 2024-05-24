package com.park.utmstack.service.collectors;

import agent.CollectorOuterClass;
import com.park.utmstack.domain.collector.UtmCollector;
import com.park.utmstack.repository.collector.UtmCollectorRepository;
import com.park.utmstack.service.dto.collectors.dto.CollectorDTO;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.Objects;

@Service
public class UtmCollectorService {

    private final UtmCollectorRepository utmCollectorRepository;
    private final DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

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
                this.utmCollectorRepository.save(collector);
            }
        }
    }
}
