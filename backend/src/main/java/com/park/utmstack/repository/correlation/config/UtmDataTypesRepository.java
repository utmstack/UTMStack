package com.park.utmstack.repository.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;

public interface UtmDataTypesRepository extends JpaRepository<UtmDataTypes, Long>, JpaSpecificationExecutor<UtmDataTypes> {
    @Query(nativeQuery = true, value = "SELECT nextval('utm_data_types_id_seq')")
    Long getNextId();
}
