package com.park.utmstack.repository.correlation.config;

import com.park.utmstack.domain.collector.UtmCollector;
import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.time.Instant;
import java.util.List;

public interface UtmDataTypesRepository extends JpaRepository<UtmDataTypes, Long>, JpaSpecificationExecutor<UtmDataTypes> {
    @Query(nativeQuery = true, value = "SELECT nextval('utm_data_types_id_seq')")
    Long getNextId();

    @Query(value = "SELECT dt FROM UtmDataTypes dt WHERE" +
            "(:search IS NULL OR ((dt.dataType LIKE :search OR lower(dt.dataType) LIKE lower(:search)) " +
            "OR (dt.dataTypeDescription LIKE :search OR lower(dt.dataTypeDescription) LIKE lower(:search))))")
    Page<UtmDataTypes> searchByFilters(@Param("search") String search,
                                       Pageable pageable);


}
