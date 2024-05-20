package com.park.utmstack.service.network_scan;

import com.park.utmstack.domain.UtmAssetMetrics;
import com.park.utmstack.domain.network_scan.AssetGroupFilter;
import com.park.utmstack.domain.network_scan.UtmAssetGroup;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.repository.UtmAssetMetricsRepository;
import com.park.utmstack.repository.network_scan.UtmAssetGroupRepository;
import com.park.utmstack.repository.network_scan.UtmNetworkScanRepository;
import com.park.utmstack.service.dto.network_scan.AssetGroupDTO;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.dao.InvalidDataAccessResourceUsageException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.EntityManager;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * Service Implementation for managing UtmAssetGroup.
 */
@Service
@Transactional
public class UtmAssetGroupService {

    private final Logger log = LoggerFactory.getLogger(UtmAssetGroupService.class);
    private static final String CLASSNAME = "UtmAssetGroupService";

    private final UtmAssetGroupRepository utmAssetGroupRepository;
    private final UtmAssetMetricsRepository assetMetricsRepository;
    private final UtmNetworkScanRepository networkScanRepository;
    private final EntityManager em;

    public UtmAssetGroupService(UtmAssetGroupRepository utmAssetGroupRepository,
                                UtmAssetMetricsRepository assetMetricsRepository,
                                UtmNetworkScanRepository networkScanRepository, EntityManager em) {
        this.utmAssetGroupRepository = utmAssetGroupRepository;
        this.assetMetricsRepository = assetMetricsRepository;
        this.networkScanRepository = networkScanRepository;
        this.em = em;
    }

    /**
     * Save a utmAssetGroup.
     *
     * @param utmAssetGroup the entity to save
     * @return the persisted entity
     */
    public UtmAssetGroup save(UtmAssetGroup utmAssetGroup) {
        log.debug("Request to save UtmAssetGroup : {}", utmAssetGroup);
        return utmAssetGroupRepository.save(utmAssetGroup);
    }

    /**
     * Get all the utmAssetGroups.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmAssetGroup> findAll(Pageable pageable) {
        log.debug("Request to get all UtmAssetGroups");
        return utmAssetGroupRepository.findAll(pageable);
    }

    public Page<AssetGroupDTO> searchGroupsByFilter(AssetGroupFilter filter, Pageable pageable) throws Exception {
        final String ctx = CLASSNAME + ".searchGroupsByFilter";
        try {
            String query = searchQueryBuilder(filter);
            String queryWithPaginationAndSort = paginateAndSort(query, pageable);
            BigInteger count = (BigInteger) em.createNativeQuery(String.format("SELECT count(*) FROM (%1$s) AS total", query)).getSingleResult();
            List<UtmAssetGroup> results = new ArrayList<>(em.createNativeQuery(queryWithPaginationAndSort, UtmAssetGroup.class).getResultList());

            if (!CollectionUtils.isEmpty(results)) {
                results.forEach(g -> {
                    Optional<List<UtmNetworkScan>> assetsOpt = networkScanRepository.findAllByGroupId(g.getId());

                    if (assetsOpt.isPresent()) {
                        g.setAssets(assetsOpt.get());
                        List<String> collect = assetsOpt.get().stream().map(UtmNetworkScan::getAssetName).collect(Collectors.toList());
                        List<UtmAssetMetrics> metrics = assetMetricsRepository.findAllByAssetNameIn(collect);
                        g.setMetrics(metrics);
                    }
                });
            }
            return new PageImpl<>(results.stream().map(AssetGroupDTO::new).collect(Collectors.toList()), pageable, count.longValue());
        } catch (InvalidDataAccessResourceUsageException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            throw new Exception(msg);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }


    /**
     * Get one utmAssetGroup by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmAssetGroup> findOne(Long id) {
        log.debug("Request to get UtmAssetGroup : {}", id);
        return utmAssetGroupRepository.findById(id);
    }

    /**
     * Delete the utmAssetGroup by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmAssetGroup : {}", id);
        utmAssetGroupRepository.deleteById(id);
    }

    private String searchQueryBuilder(AssetGroupFilter filters) {
        StringBuilder sb = new StringBuilder();
        sb.append(String.format("SELECT DISTINCT utm_asset_group.* FROM utm_asset_group LEFT JOIN utm_network_scan ON utm_asset_group.id = utm_network_scan.group_id where type =  '%s' \n", filters.getAssetType()));

        if (Objects.isNull(filters))
            return sb.toString();

        boolean where = false;

        // id
        if (Objects.nonNull(filters.getId())) {
            sb.append(String.format("WHERE utm_asset_group.id = %1$s\n", filters.getId()));
            where = false;
        }

        // groupName
        if (StringUtils.hasText(filters.getGroupName())) {
            sb.append(where ? "WHERE " : "AND ")
                .append(String.format("lower(utm_asset_group.group_name) LIKE '%%%1$s%%'\n",
                    filters.getGroupName().toLowerCase()));
            where = false;
        }

        // createdDate
        if (Objects.nonNull(filters.getInitDate()) && Objects.nonNull(filters.getEndDate())) {
            sb.append(where ? "WHERE " : "AND ")
                .append(String.format("(utm_asset_group.created_date BETWEEN '%1$s' AND '%2$s')\n",
                    filters.getInitDate(), filters.getEndDate()));
            where = false;
        }

        // assetType
        if (!CollectionUtils.isEmpty(filters.getType())) {
            String types = filters.getType().stream()
                .map(type -> String.format("'%1$s'", type)).collect(Collectors.joining(","));
            sb.append(where ? "WHERE " : "AND ")
                .append(String.format("utm_network_scan.asset_type_id IN (SELECT utm_asset_types.id FROM utm_asset_types WHERE utm_asset_types.type_name IN (%1$s))\n", types));
            where = false;
        }

        // serverName
        if (!CollectionUtils.isEmpty(filters.getProbe())) {
            String probes = filters.getProbe().stream()
                .map(probe -> String.format("'%1$s'", probe)).collect(Collectors.joining(","));
            sb.append(where ? "WHERE " : "AND ")
                .append(String.format("utm_network_scan.server_name IN (%1$s)\n", probes));
            where = false;
        }

        // assetOs
        if (!CollectionUtils.isEmpty(filters.getOs())) {
            String oss = filters.getOs().stream()
                .map(os -> String.format("'%1$s'", os)).collect(Collectors.joining(","));
            sb.append(where ? "WHERE " : "AND ")
                .append(String.format("utm_network_scan.asset_os IN (%1$s)\n", oss));
            where = false;
        }

        // assetIp
        if (!CollectionUtils.isEmpty(filters.getAssetIp())) {
            String ips = filters.getAssetIp().stream()
                .map(ip -> String.format("'%1$s'", ip)).collect(Collectors.joining(","));
            sb.append(where ? "WHERE " : "AND ")
                .append(String.format("utm_network_scan.asset_ip IN (%1$s)\n", ips));
            where = false;
        }

        // assetName
        if (!CollectionUtils.isEmpty(filters.getAssetName())) {
            String names = filters.getAssetName().stream()
                .map(name -> String.format("'%1$s'", name)).collect(Collectors.joining(","));
            sb.append(where ? "WHERE " : "AND ")
                .append(String.format("utm_network_scan.asset_name IN (%1$s)\n", names));
        }

        return sb.toString();
    }

    private String paginateAndSort(String query, Pageable pageable) {
        final String ctx = CLASSNAME + ".paginateAndSort";
        StringBuilder sb = new StringBuilder(query);

        try {
            Sort sort = pageable.getSort();

            if (sort.isSorted()) {
                sb.append("ORDER BY ");
                boolean firstProperty = true;

                List<Sort.Order> orders = sort.stream().collect(Collectors.toList());

                for (Sort.Order order : orders) {
                    sb.append(String.format(firstProperty ? "%1$s %2$s" : ", %1$s %2$s", order.getProperty(), order.getDirection().name()));
                    firstProperty = false;
                }
            }

            if (pageable.isPaged())
                sb.append(String.format(" OFFSET %1$s LIMIT %2$s", pageable.getOffset(), pageable.getPageSize()));

            return sb.toString();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
