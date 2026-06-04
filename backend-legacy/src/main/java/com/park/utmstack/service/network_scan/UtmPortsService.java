package com.park.utmstack.service.network_scan;

import com.park.utmstack.domain.network_scan.UtmPorts;
import com.park.utmstack.repository.network_scan.UtmPortsRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmOpenPort.
 */
@Service
@Transactional
public class UtmPortsService {

    private final Logger log = LoggerFactory.getLogger(UtmPortsService.class);

    private final UtmPortsRepository portsRepository;

    public UtmPortsService(UtmPortsRepository utmOpenPortRepository) {
        this.portsRepository = utmOpenPortRepository;
    }

    /**
     * Save a utmOpenPort.
     *
     * @param utmPorts the entity to save
     * @return the persisted entity
     */
    public UtmPorts save(UtmPorts utmPorts) {
        log.debug("Request to save UtmOpenPort : {}", utmPorts);
        return portsRepository.save(utmPorts);
    }

    /**
     * Get all the utmOpenPorts.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmPorts> findAll(Pageable pageable) {
        log.debug("Request to get all UtmOpenPorts");
        return portsRepository.findAll(pageable);
    }


    /**
     * Get one utmOpenPort by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmPorts> findOne(Long id) {
        log.debug("Request to get UtmOpenPort : {}", id);
        return portsRepository.findById(id);
    }

    /**
     * Delete the utmOpenPort by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmOpenPort : {}", id);
        portsRepository.deleteById(id);
    }

    public void deleteAllByScanId(Long scanId) {
        log.debug("Request to delete all ports of scan : {}", scanId);
        portsRepository.deleteAllByScanId(scanId);
    }

    public void deleteAllByScanIdIn(List<Long> scanIds) {
        portsRepository.deleteAllByScanIdIn(scanIds);
    }

    public List<UtmPorts> updateInBatch(Long scanId, List<UtmPorts> portList) {
        if (scanId != null)
            portsRepository.deleteAllByScanId(scanId);
        return portsRepository.saveAll(portList);
    }

    public List<UtmPorts> updateInBatch(List<Long> scanIds, List<UtmPorts> portList) {
        if (!CollectionUtils.isEmpty(scanIds))
            deleteAllByScanIdIn(scanIds);
        if (!CollectionUtils.isEmpty(portList))
            portsRepository.saveAll(portList);
        return portList;
    }

    public List<UtmPorts> saveInBatch(List<UtmPorts> portList) {
        return portsRepository.saveAll(portList);
    }
}
