package com.park.utmstack.repository.correlation.config;

import com.park.utmstack.domain.correlation.config.UtmTenantConfig;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;


public interface UtmTenantConfigRepository extends JpaRepository<UtmTenantConfig, Long>, JpaSpecificationExecutor<UtmTenantConfig> {
}
