package com.park.utmstack.service.collectors;

import agent.CollectorOuterClass;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.collector.UtmCollector;
import com.park.utmstack.domain.network_scan.NetworkScanFilter;
import com.park.utmstack.repository.collector.UtmCollectorRepository;
import com.park.utmstack.service.dto.application_modules.ModuleActivationDTO;
import com.park.utmstack.service.dto.collectors.dto.CollectorDTO;
import com.park.utmstack.util.exceptions.ApiException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.dao.InvalidDataAccessResourceUsageException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.Assert;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.Objects;
import java.util.Optional;

import static com.park.utmstack.config.RestTemplateConfiguration.CLASSNAME;

@Service
public class UtmCollectorService {

    private static final String CLASSNAME = "UtmCollectorService";

    private final UtmCollectorRepository utmCollectorRepository;
    private final DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    private final Logger log = LoggerFactory.getLogger(UtmCollectorService.class);

    public UtmCollectorService(UtmCollectorRepository utmCollectorRepository) {
        this.utmCollectorRepository = utmCollectorRepository;
    }

    public UtmCollector saveCollector(CollectorOuterClass.Collector collector) {
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
            log.error("{}: Error searching collectors with filters {}", ctx, e.getMessage(), e);
            throw new ApiException(String.format("%s: Error searching collectors with filters", ctx), HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    private Page<UtmCollector> filter(NetworkScanFilter f, Pageable p) {

        return utmCollectorRepository.searchByFilters(
                f.getAssetIpMacName() == null ? null : "%" + f.getAssetIpMacName() + "%",
                f.getDiscoveredInitDate(),
                f.getDiscoveredEndDate(),
                f.getGroups(), p);

    }

    @Transactional
    public void updateGroup(List<Long> collectorsIds, Long assetGroupId) {
        String ctx = CLASSNAME + ".updateGroup";

        try {
             utmCollectorRepository.updateGroup(collectorsIds, assetGroupId);
        } catch (Exception ex) {
            log.error("{}: Error updating group for collectors {}: {}", ctx, collectorsIds, ex.getMessage(), ex);
            throw new ApiException(String.format("%s: Error updating group for collectors %s", ctx, collectorsIds), HttpStatus.INTERNAL_SERVER_ERROR);
        }

    }

    Optional<UtmCollector> findById(Long id) {
        return utmCollectorRepository.findById(id);
    }

    @Transactional
    public void deleteCollector(Long id) {
        utmCollectorRepository.deleteById(id);
    }
}
