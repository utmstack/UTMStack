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
import javax.persistence.Query;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Optional;
import java.util.Set;
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

    public Page<AssetGroupDTO> searchGroupsByFilter(AssetGroupFilter filter, Pageable pageable) {

        Map<String, Object> params = new HashMap<>();
        String query = searchQueryBuilder(filter, params);
        String queryWithPaginationAndSort = paginateAndSort(query, pageable);

        Query countQuery = em.createNativeQuery(String.format("SELECT count(*) FROM (%1$s) AS total", query));
        setQueryParams(countQuery, params);
        BigInteger count = (BigInteger) countQuery.getSingleResult();

        Query dataQuery = em.createNativeQuery(queryWithPaginationAndSort, UtmAssetGroup.class);
        setQueryParams(dataQuery, params);
        List<UtmAssetGroup> results = new ArrayList<>(dataQuery.getResultList());

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

    private static final Set<String> ALLOWED_SORT_COLUMNS = Set.of(
            "id", "group_name", "group_description", "created_date", "type"
    );

    private static void setQueryParams(Query query, Map<String, Object> params) {
        params.forEach(query::setParameter);
    }

    private String searchQueryBuilder(AssetGroupFilter filters, Map<String, Object> params) {
        StringBuilder sb = new StringBuilder();
        sb.append("SELECT DISTINCT utm_asset_group.* FROM utm_asset_group LEFT JOIN utm_network_scan ON utm_asset_group.id = utm_network_scan.group_id");

        if (Objects.isNull(filters))
            return sb.toString();

        List<String> conditions = new ArrayList<>();

        // id
        if (Objects.nonNull(filters.getId())) {
            conditions.add("utm_asset_group.id = :filterId");
            params.put("filterId", filters.getId());
        }

        // groupName
        if (StringUtils.hasText(filters.getGroupName())) {
            conditions.add("lower(utm_asset_group.group_name) LIKE :filterGroupName");
            params.put("filterGroupName", "%" + filters.getGroupName().toLowerCase() + "%");
        }

        // createdDate
        if (Objects.nonNull(filters.getInitDate()) && Objects.nonNull(filters.getEndDate())) {
            conditions.add("(utm_asset_group.created_date BETWEEN :filterInitDate AND :filterEndDate)");
            params.put("filterInitDate", filters.getInitDate());
            params.put("filterEndDate", filters.getEndDate());
        }

        // assetType
        if (!CollectionUtils.isEmpty(filters.getType())) {
            conditions.add("utm_network_scan.asset_type_id IN (SELECT utm_asset_types.id FROM utm_asset_types WHERE utm_asset_types.type_name IN (:filterTypes))");
            params.put("filterTypes", filters.getType());
        }

        // serverName
        if (!CollectionUtils.isEmpty(filters.getProbe())) {
            conditions.add("utm_network_scan.server_name IN (:filterProbes)");
            params.put("filterProbes", filters.getProbe());
        }

        // assetOs
        if (!CollectionUtils.isEmpty(filters.getOs())) {
            conditions.add("utm_network_scan.asset_os IN (:filterOs)");
            params.put("filterOs", filters.getOs());
        }

        // assetIp
        if (!CollectionUtils.isEmpty(filters.getAssetIp())) {
            conditions.add("utm_network_scan.asset_ip IN (:filterAssetIps)");
            params.put("filterAssetIps", filters.getAssetIp());
        }

        // assetName
        if (!CollectionUtils.isEmpty(filters.getAssetName())) {
            conditions.add("utm_network_scan.asset_name IN (:filterAssetNames)");
            params.put("filterAssetNames", filters.getAssetName());
        }

        // assetType (single value column)
        if (StringUtils.hasText(filters.getAssetType())) {
            conditions.add("type = :filterAssetType");
            params.put("filterAssetType", filters.getAssetType());
        }

        if (!conditions.isEmpty()) {
            sb.append(" WHERE ").append(String.join(" AND ", conditions));
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
                    if (!ALLOWED_SORT_COLUMNS.contains(order.getProperty()))
                        throw new IllegalArgumentException(ctx + ": Invalid sort column: " + order.getProperty());
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
