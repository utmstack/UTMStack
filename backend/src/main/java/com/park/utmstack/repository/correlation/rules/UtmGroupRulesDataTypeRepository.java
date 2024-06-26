package com.park.utmstack.repository.correlation.rules;

import com.park.utmstack.domain.correlation.rules.UtmGroupRulesDataType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

/**
 * Spring Data  repository for the UtmGroupRulesDataType entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmGroupRulesDataTypeRepository extends JpaRepository<UtmGroupRulesDataType, Long> {
}
