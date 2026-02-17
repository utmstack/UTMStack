package com.park.utmstack.service.collectors;

import agent.CollectorOuterClass;
import com.park.utmstack.domain.collector.UtmCollector;
import com.park.utmstack.domain.network_scan.AssetGroupFilter;
import com.park.utmstack.domain.network_scan.UtmAssetGroup;
import com.park.utmstack.repository.collector.UtmCollectorRepository;
import com.park.utmstack.service.application_modules.UtmModuleGroupService;
import com.park.utmstack.service.dto.collectors.CollectorModuleEnum;
import com.park.utmstack.service.dto.collectors.dto.CollectorConfigDTO;
import com.park.utmstack.service.dto.collectors.dto.CollectorDTO;
import com.park.utmstack.service.dto.collectors.dto.ListCollectorsResponseDTO;
import com.park.utmstack.service.dto.network_scan.AssetGroupDTO;
import com.park.utmstack.service.grpc.ListRequest;
import com.park.utmstack.util.exceptions.ApiException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.EntityManager;
import javax.persistence.Query;
import java.util.*;
import java.util.stream.Collectors;

@Service
@Slf4j
@RequiredArgsConstructor
public class CollectorService {

    private final CollectorGrpcService collectorGrpcService;
    private final UtmModuleGroupService moduleGroupService;
    private final CollectorConfigBuilder CollectorConfigBuilder;
    private final UtmCollectorService utmCollectorService;
    private final UtmCollectorRepository collectorRepository;
    private final EntityManager em;
    private final String CLASSNAME = "CollectorService";

    private static final Set<String> ALLOWED_SORT_COLUMNS = Set.of(
            "id",
            "group_name",
            "created_date",
            "type"
    );


    public void upsertCollectorConfig(CollectorConfigDTO collectorConfig) {

        this.moduleGroupService.updateCollectorConfigurationKeys(collectorConfig);

        CollectorOuterClass.CollectorConfig collector = CollectorConfigBuilder.build(collectorConfig);
        collectorGrpcService.upsertCollectorConfig(collector);
    }

    public BulkCollectorConfigResponseDTO upsertCollectorsConfig(List<CollectorConfigDTO> collectors) {

        List<CollectorConfigResultDTO> results = collectors.stream()
                .map(this::processSingleCollectorConfig)
                .toList();

        return BulkCollectorConfigResponseDTO.builder()
                .results(results)
                .build();
    }

    public ListCollectorsResponseDTO listCollector(String hostname, Integer pageNumber, Integer pageSize, String sortBy, CollectorModuleEnum module) {

        String cts = CLASSNAME + ".listCollector";

        try {

            var request = buildListRequest(pageNumber, pageSize, module, hostname, sortBy);

            CollectorOuterClass.ListCollectorResponse collectorResponse = collectorGrpcService.listCollectors(request);
            return mapToListCollectorsResponseDTO(collectorResponse);

        } catch (Exception e) {
            log.error("{}: Error listing collectors with hostname={} and module={}: {}", cts, hostname, module, e.getMessage(), e);
            throw new ApiException(String.format("%s: Error listing collectors.", cts), HttpStatus.INTERNAL_SERVER_ERROR);
        }

    }

    private Optional<CollectorDTO> findCollectorByHostname(String hostname, CollectorModuleEnum module) {

        ListCollectorsResponseDTO response = this.listCollector(hostname, 0, 1, null, module);

        if (response.getCollectors() != null && !response.getCollectors().isEmpty()) {
            return Optional.of(response.getCollectors().get(0));
        } else {
            return Optional.empty();
        }
    }

    public void deleteCollector(Long id) {

        String ctx = CLASSNAME + ".deleteCollector";

        Optional<UtmCollector> collector = utmCollectorService.findById(id);

        if (collector.isEmpty()) {

            log.error("{}: Collector with id {} not found", ctx, id);
            throw new ApiException(String.format("%s: Collector with id %d not found", ctx, id), HttpStatus.NOT_FOUND);

        } /*else if (collector.get().isActive()) {

            var collectorToDelete = collector.get();

            Optional<CollectorDTO> collectorDTO = this.findCollectorByHostname(
                    collector.get().getHostname(),
                    CollectorModuleEnum.valueOf(collectorToDelete.getModule()));

            if (collectorDTO.isEmpty()) {

                log.error("{}: Collector with id {} not found in Agent Manager", ctx, id);
                throw new ApiException(String.format("%s: Collector with id %d not found in Agent Manager", ctx, id), HttpStatus.NOT_FOUND);

            } else {
                var c = collectorDTO.get();
                collectorGrpcService.deleteCollector(c.getId(), c.getCollectorKey());
            }

            this.moduleGroupService.deleteCollectorById(collectorToDelete.getId());

        }*/

        this.utmCollectorService.deleteCollector(id);
    }

    public Page<AssetGroupDTO> searchGroupsByFilter(AssetGroupFilter filter, Pageable pageable) {

        final String ctx = CLASSNAME + ".searchGroupsByFilter";
        try {

            Query countQuery = buildSearchQuery(filter, true, pageable);
            long total = ((Number) countQuery.getSingleResult()).longValue();

            Query dataQuery = buildSearchQuery(filter, false, pageable);
            List<UtmAssetGroup> groups = dataQuery.getResultList();

            enrichGroups(groups);

            List<AssetGroupDTO> dtos = groups.stream()
                    .map(AssetGroupDTO::new)
                    .toList();

            return new PageImpl<>(dtos, pageable, total);

        } catch (Exception ex) {
            log.error("{}: Error searching collector groups with filters {}: {}", ctx, filter, ex.getMessage(), ex);
            throw new ApiException(String.format("%s: Error searching asset groups with filters.", ctx), HttpStatus.INTERNAL_SERVER_ERROR);
        }


    }

    private String searchQueryBuilder(AssetGroupFilter filters) {
        StringBuilder sb = new StringBuilder();
        sb.append("SELECT DISTINCT ag.* FROM utm_asset_group ag ");
        sb.append("LEFT JOIN utm_collectors c ON ag.id = c.group_id ");

        List<String> conditions = new ArrayList<>();

        if (filters != null) {

            if (filters.getAssetType() != null) {
                conditions.add("ag.type = :type");
            }

            if (filters.getId() != null) {
                conditions.add("ag.id = :id");
            }

            if (StringUtils.hasText(filters.getGroupName())) {
                conditions.add("LOWER(ag.group_name) LIKE :groupName");
            }

            if (filters.getInitDate() != null && filters.getEndDate() != null) {
                conditions.add("ag.created_date BETWEEN :initDate AND :endDate");
            }

            if (!CollectionUtils.isEmpty(filters.getAssetIp())) {
                conditions.add("c.ip IN (:ips)");
            }

            if (!CollectionUtils.isEmpty(filters.getAssetName())) {
                conditions.add("c.hostname IN (:names)");
            }
        }

        if (!conditions.isEmpty()) {
            sb.append(" WHERE ");
            sb.append(String.join(" AND ", conditions));
        }

        return sb.toString();
    }


    private ListCollectorsResponseDTO getListCollector(ListRequest request) {

        CollectorOuterClass.ListCollectorResponse collectorResponse = collectorGrpcService.listCollectors(request);
        return mapToListCollectorsResponseDTO(collectorResponse);
    }


    private ListCollectorsResponseDTO mapToListCollectorsResponseDTO(CollectorOuterClass.ListCollectorResponse response) {
        final String ctx = CLASSNAME + ".mapToListCollectorsResponseDTO";
        try {
            ListCollectorsResponseDTO dto = new ListCollectorsResponseDTO();

            List<CollectorDTO> collectorDTOS = response.getRowsList().stream()
                    .map(this::protoToCollectorDto)
                    .collect(Collectors.toList());

            this.utmCollectorService.synchronize(collectorDTOS);

            dto.setCollectors(collectorDTOS);
            dto.setTotal(response.getTotal());

            return dto;
        } catch (Exception e) {
            log.error("{}: Error mapping ListCollectorResponse to ListCollectorsResponseDTO: {}", ctx, e.getMessage());
            throw new ApiException(String.format("%s: Error mapping ListCollectorResponse to ListCollectorsResponseDTO: %s", ctx, e.getMessage()), HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    private CollectorDTO protoToCollectorDto(CollectorOuterClass.Collector collector) {
        UtmCollector utmCollector = this.utmCollectorService.saveCollector(collector);
        return new CollectorDTO(utmCollector);
    }

    private CollectorConfigResultDTO processSingleCollectorConfig(CollectorConfigDTO dto) {

        try {
            this.upsertCollectorConfig(dto);

            return CollectorConfigResultDTO.builder()
                    .collectorId(dto.getCollector().getId())
                    .success(true)
                    .build();

        } catch (Exception e) {

            return CollectorConfigResultDTO.builder()
                    .collectorId(dto.getCollector().getId())
                    .success(false)
                    .errorMessage(e.getMessage())
                    .build();
        }
    }

    private Query buildSearchQuery(AssetGroupFilter filter, boolean countQuery, Pageable pageable) {

        StringBuilder sql = new StringBuilder();
        Map<String, Object> params = new HashMap<>();

        sql.append("""
                    SELECT DISTINCT ag.*
                    FROM utm_asset_group ag
                    LEFT JOIN utm_collectors c ON ag.id = c.group_id
                    WHERE 1=1
                """);

        if (filter.getAssetType() != null) {
            sql.append(" AND ag.type = :type ");
            params.put("type", filter.getAssetType());
        }

        if (filter.getId() != null) {
            sql.append(" AND ag.id = :id ");
            params.put("id", filter.getId());
        }

        if (StringUtils.hasText(filter.getGroupName())) {
            sql.append(" AND LOWER(ag.group_name) LIKE :groupName ");
            params.put("groupName", "%" + filter.getGroupName().toLowerCase() + "%");
        }

        if (filter.getInitDate() != null && filter.getEndDate() != null) {
            sql.append(" AND ag.created_date BETWEEN :initDate AND :endDate ");
            params.put("initDate", filter.getInitDate());
            params.put("endDate", filter.getEndDate());
        }

        if (!CollectionUtils.isEmpty(filter.getAssetIp())) {
            sql.append(" AND c.ip IN :ips ");
            params.put("ips", filter.getAssetIp());
        }

        if (!CollectionUtils.isEmpty(filter.getAssetName())) {
            sql.append(" AND c.hostname IN :names ");
            params.put("names", filter.getAssetName());
        }

        if (countQuery) {
            sql.insert(0, "SELECT COUNT(*) FROM (");
            sql.append(") AS total");
        } else {
            sql.append(buildOrderAndPagination(pageable));
        }

        Query q;

        if (countQuery) {
            q = em.createNativeQuery(sql.toString());
        } else {
            q = em.createNativeQuery(sql.toString(), UtmAssetGroup.class);
        }

        params.forEach(q::setParameter);

        return q;
    }


    private String buildOrderAndPagination(Pageable pageable) {
        StringBuilder sb = new StringBuilder(" ");

        Sort sort = pageable.getSort();

        if (sort.isSorted()) {
            sb.append(" ORDER BY ");

            List<String> clauses = sort.stream()
                    .map(order -> {
                        validateSortColumn(order.getProperty());
                        return order.getProperty() + " " + order.getDirection().name();
                    })
                    .toList();

            sb.append(String.join(", ", clauses));
        }

        if (pageable.isPaged()) {
            sb.append(" OFFSET ").append(pageable.getOffset());
            sb.append(" LIMIT ").append(pageable.getPageSize());
        }

        return sb.toString();
    }

    private void validateSortColumn(String column) {
        if (!ALLOWED_SORT_COLUMNS.contains(column)) {
            throw new IllegalArgumentException("Invalid sort column: " + column);
        }
    }

    private void enrichGroups(List<UtmAssetGroup> groups) {
        if (groups.isEmpty()) return;

        List<Long> ids = groups.stream().map(UtmAssetGroup::getId).toList();

        Map<Long, List<UtmCollector>> collectors =
                collectorRepository.findAllByGroupIdIn(ids)
                        .stream()
                        .collect(Collectors.groupingBy(UtmCollector::getGroupId));

        groups.forEach(g ->
                g.setCollectors(collectors.getOrDefault(g.getId(), List.of()))
        );
    }

    private ListRequest buildListRequest(Integer pageNumber, Integer pageSize,
                                         CollectorModuleEnum module, String hostName, String sortBy) {
        return ListRequest.newBuilder()
                .setPageNumber(pageNumber != null ? pageNumber : 0)
                .setPageSize(pageSize != null ? pageSize : 10)
                .setSearchQuery(buildSearchQuery(module, hostName))
                .setSortBy(sortBy != null ? sortBy : "")
                .build();
    }

    private String buildSearchQuery(CollectorModuleEnum module, String hostName) {
        List<String> conditions = new ArrayList<>();

        if (module != null) {
            conditions.add("module.Is=" + module.name());
        }

        if (StringUtils.hasText(hostName)) {
            conditions.add("hostname.Is=" + hostName);
        }

        return String.join("&", conditions);
    }


}

