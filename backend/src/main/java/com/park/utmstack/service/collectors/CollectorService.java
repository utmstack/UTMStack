package com.park.utmstack.service.collectors;

import agent.CollectorOuterClass;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.collector.UtmCollector;
import com.park.utmstack.service.application_modules.UtmModuleGroupConfigurationService;
import com.park.utmstack.service.dto.collectors.CollectorHostnames;
import com.park.utmstack.service.dto.collectors.CollectorModuleEnum;
import com.park.utmstack.service.dto.collectors.dto.CollectorConfigDTO;
import com.park.utmstack.service.dto.collectors.dto.CollectorDTO;
import com.park.utmstack.service.dto.collectors.dto.ListCollectorsResponseDTO;
import com.park.utmstack.service.grpc.ListRequest;
import com.park.utmstack.util.exceptions.ApiException;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.utmstack.grpc.exception.CollectorServiceGrpcException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.List;
import java.util.stream.Collectors;

import static com.park.utmstack.config.RestTemplateConfiguration.CLASSNAME;

@Service
@Slf4j
@RequiredArgsConstructor
public class CollectorService {

    private final CollectorGrpcService collectorGrpcService;
    private final UtmModuleGroupConfigurationService moduleGroupConfigurationService;
    private final CollectorConfigBuilder CollectorConfigBuilder;
    private final UtmCollectorService utmCollectorService;

    public void upsertCollectorConfig(CollectorConfigDTO collectorConfig) {

        this.moduleGroupConfigurationService.updateConfigurationKeys(collectorConfig.getModuleId(), collectorConfig.getKeys());

        CollectorOuterClass.CollectorConfig collector = CollectorConfigBuilder.build(collectorConfig);
        collectorGrpcService.upsertCollectorConfig(collector);
    }

    public ListCollectorsResponseDTO listCollector(ListRequest request) {
        return this.getListCollector(request);
    }

    public ListCollectorsResponseDTO listCollector(String hostname, CollectorModuleEnum module) {

        String query = "";
        if (module != null && StringUtils.hasText(hostname)) {
            query = "module.Is=" + module.name() + "&hostname.Is=" + hostname;
        } else if (StringUtils.hasText(hostname)) {
            query = "hostname.Is=" + hostname;
        } else if (module != null) {
            query = "module.Is=" + module.name();
        }

        var request = ListRequest.newBuilder()
                .setPageNumber(1)
                .setPageSize(1000000)
                .setSearchQuery(query)
                .setSortBy("id,desc")
                .build();

        CollectorOuterClass.ListCollectorResponse collectorResponse = collectorGrpcService.listCollectors(request);
        return mapToListCollectorsResponseDTO(collectorResponse);
    }

    public CollectorHostnames listCollectorHostnames(ListRequest request) {


        CollectorOuterClass.ListCollectorResponse response = collectorGrpcService.listCollectors(request);
        CollectorHostnames collectorHostnames = new CollectorHostnames();

        response.getRowsList().forEach(c -> {
            collectorHostnames.getHostname().add(c.getHostname());
        });

        return collectorHostnames;

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

}

