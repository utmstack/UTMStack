package com.park.utmstack.repository.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;

public interface UtmDataTypesRepository extends JpaRepository<UtmDataTypes, Long>, JpaSpecificationExecutor<UtmDataTypes> {
}
